package main

// NOTE: ffmpeg is required on server
import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
)

var ctx = context.Background()
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func compressToCodec(codec, fileName string, rateFacor int, client *redis.Client) {

	inputPath := "/tmp/compression/input_" + fileName + ".mov"
	outputPath := "/tmp/compression/output_" + fileName + "_" + codec + ".mp4"

	buf := &bytes.Buffer{}
	cmd := fluentffmpeg.NewCommand("").
		// TODO: use PipeInput which accepts io.Reader, instead of InputPath - figure out why it won't accept "bytes.NewReader(videoBytes)""
		InputPath(inputPath).
		OutputPath(outputPath).
		ConstantRateFactor(rateFacor). // 17 - 28 is optimal range: https://trac.ffmpeg.org/wiki/Encode/H.264
		VideoCodec(codec).             // eg: libx264, libaom-av1, etc..
		Overwrite(true).
		OutputLogs(buf). // log results to buffer
		Build()
	cmd.Run()

	// Get the output file in bytes
	dat, err := os.ReadFile(outputPath)
	if err != nil {
		fmt.Println("Read error:", err)
		panic(err)
	}

	// publish compressed file
	pubRes := redisClient.Publish(ctx, "compressed-video", string(dat))
	if pubRes.Err() != nil {
		fmt.Println("Publish error:", pubRes.Err().Error())
		panic(pubRes.Err().Error())
	}

	out, _ := ioutil.ReadAll(buf) // read logs
	if len(out) > 0 {
		fmt.Println(codec+" Logs:", string(out))
	}

	// Remove tmp files
	os.Remove(inputPath)
	// TODO: remove output files
}

func listenForValidVideo(client *redis.Client) {
	// Subscribe to "valid-upload"
	subscriber := client.Subscribe(ctx, "valid-upload")
	defer subscriber.Close()
	var f *os.File
	defer f.Close()

	var err error
	var msg *redis.Message

	// use channels to handle multiple requests: https://stackoverflow.com/a/58237162/11278697
	controlCh := subscriber.Channel()
	fmt.Println("Compression service is running")

	// use WaitGroup for Goroutines to run processing in parallel
	wg := &sync.WaitGroup{}
	// Continuously listen for valid-upload events
	for msg = range controlCh {
		go func(nextMsg *redis.Message) {
			// Convert string video data to byte array
			videoBytes := []byte(nextMsg.Payload)

			// write input video to file system with "unique-ish" name (file-size_timestamp-ms.format)
			timestamp := time.Now().UnixMilli()
			fileName := strconv.Itoa(len(videoBytes)) + "_" + strconv.Itoa(int(timestamp))
			inputFilePath := "/tmp/compression/input_" + fileName + ".mov"
			err = os.MkdirAll("/tmp/compression", 0750)
			if err != nil && !os.IsExist(err) {
				panic(err)
			}
			f, err = os.Create(inputFilePath)
			if err != nil {
				fmt.Println("Writer error:", err)
				panic(err)
			}
			f.Write(videoBytes)

			wg.Add(1) // increment wait group counter by one every loop
			go func() {
				compressToCodec("libx264", fileName, 28, redisClient)
				wg.Done() // decrement counter by one when iteration is complete
			}()
		}(msg)
	}
	wg.Wait() // blocks until wait group counter is 0
}

func main() {
	// TODO: consider checking current codec before compression, and how different files may be treated differently
	listenForValidVideo(redisClient)
}
