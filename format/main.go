package main

// NOTE: ffmpeg is required on server
import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os/exec"

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

func formatResolution(fileName string, resolution int, width, height float64, client *redis.Client) {

	inputPath := "/tmp/format/input_" + fileName + ".mp4"
	outputPath := "/tmp/format/output_" + fileName + "_" + strconv.Itoa(resolution) + ".mp4"

	var widthOverHeight float64 = width / height
	newHeight := resolution
	newWidth := int(math.Round(widthOverHeight * float64(newHeight)))
	if newWidth%2 != 0 {
		// width must be divisble by two
		newWidth--
	}

	fmt.Println("newDims:", newWidth, newHeight)

	cmd := exec.Command("ffmpeg", "-i", inputPath, "-filter:v", "scale="+strconv.Itoa(newWidth)+":"+strconv.Itoa(resolution), "-preset", "medium", "-crf", "24", "-c:a", "copy", outputPath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Oh no bro!", resolution, newWidth, newHeight, fmt.Sprint(err)+": "+stderr.String())
		return
	}
	// Get the output file in bytes
	dat, err := os.ReadFile(outputPath)
	if err != nil {
		fmt.Println("Read error:", err)
		panic(err)
	}

	// publish resolution file
	pubRes := redisClient.Publish(ctx, "format-video-"+strconv.Itoa(resolution), string(dat))
	if pubRes.Err() != nil {
		fmt.Println("Publish error:", pubRes.Err().Error())
		panic(pubRes.Err().Error())
	}

	// Remove tmp files
	os.Remove(inputPath)
	// TODO: remove output files
}

func listenForCompressedVideo(client *redis.Client) {
	// Subscribe to "valid-upload"
	subscriber := client.Subscribe(ctx, "compressed-video")
	defer subscriber.Close()
	var f *os.File
	defer f.Close()

	var err error
	var msg *redis.Message

	// use channels to handle multiple requests: https://stackoverflow.com/a/58237162/11278697
	controlCh := subscriber.Channel()
	fmt.Println("Format service is running")

	// use WaitGroup for Goroutines to run processing in parallel
	wg := &sync.WaitGroup{}
	// Continuously listen for valid-upload events
	for msg = range controlCh {
		wg.Add(1) // increment wait group counter by one every loop
		go func(nextMsg *redis.Message) {
			// Convert string video data to byte array
			videoBytes := []byte(nextMsg.Payload)

			// write input video to file system with "unique-ish" name (file-size_timestamp-ms.format)
			timestamp := time.Now().UnixMilli()
			fileName := strconv.Itoa(len(videoBytes)) + "_" + strconv.Itoa(int(timestamp))
			inputFilePath := "/tmp/format/input_" + fileName + ".mp4"
			err = os.MkdirAll("/tmp/format", 0750)
			if err != nil && !os.IsExist(err) {
				panic(err)
			}
			f, err = os.Create(inputFilePath)
			if err != nil {
				fmt.Println("Writer error:", err)
				panic(err)
			}
			f.Write(videoBytes)

			probeData, probeErr := fluentffmpeg.Probe(inputFilePath)
			if probeErr != nil {
				panic(probeErr)
			}

			vidWidth := probeData["streams"].([]interface{})[0].(map[string]interface{})["width"].(float64)
			vidHeight := probeData["streams"].([]interface{})[0].(map[string]interface{})["height"].(float64)
			fmt.Println("probeData:", vidWidth, vidHeight)

			if vidHeight >= 1080 {
				wg.Add(1) // increment wait group counter by one every loop
				go func() {
					formatResolution(fileName, 1080, vidWidth, vidHeight, redisClient)
					wg.Done() // decrement counter by one when iteration is complete
				}()
			}

			if vidHeight >= 720 {
				wg.Add(1) // increment wait group counter by one every loop
				go func() {
					formatResolution(fileName, 720, vidWidth, vidHeight, redisClient)
					wg.Done() // decrement counter by one when iteration is complete
				}()
			}

			if vidHeight >= 480 {
				wg.Add(1) // increment wait group counter by one every loop
				go func() {
					formatResolution(fileName, 480, vidWidth, vidHeight, redisClient)
					wg.Done() // decrement counter by one when iteration is complete
				}()
			}

			wg.Done() // decrement counter by one when iteration is complete

		}(msg)
	}
	wg.Wait() // blocks until wait group counter is 0
}

func main() {
	// TODO: consider checking current codec before compression, and how different files may be treated differently
	listenForCompressedVideo(redisClient)
}
