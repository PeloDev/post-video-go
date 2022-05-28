package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-redis/redis/v8"
)

// TODO: refactor into coherent folder/file structure

// PubSub with Redis: https://dev.to/franciscomendes10866/using-redis-pub-sub-with-golang-mf9
// Currently using local redis-server. For a prod setup will need to install on server and probably use an init script:
// -> https://redis.io/docs/getting-started/#installing-redis-more-properly
var ctx = context.Background()
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

type ResponseOkay struct {
	Success bool `json:"success"`
}

func upload(w http.ResponseWriter, r *http.Request) {

	// check if POST request
	if r.Method != "POST" {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	// limit the size of the request body / file
	var maxFileSize int64 = 240
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize<<20+1024)

	// create a buffer for the file
	var buf bytes.Buffer

	// read file, or fail
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Println("err:", err)
		http.Error(w, "Sorry, your file could not be read.", http.StatusInternalServerError)
		return
	}

	// "defer" will call file.Close() before upload func returns
	defer file.Close()

	// copy file data to buffer
	io.Copy(&buf, file)

	// get mime type from bytes
	// http.DetectContentType([]byte) returns "application/octet-stream" for videos, using "github.com/gabriel-vasile/mimetype" instead...
	fileType := mimetype.Detect(buf.Bytes())

	validMimeTypes := []string{"video/mp4", "video/quicktime", "video/mpeg", "video/x-msvideo", "video/webm", "video/3gpp", "video/3gpp2"}

	// validate file type against valid mime types
	if !mimetype.EqualsAny(fileType.String(), validMimeTypes...) {
		http.Error(w, "Sorry, your file is not a valid video format.", http.StatusBadRequest)
		return
	}

	// Publish video
	pubRes := redisClient.Publish(ctx, "valid-upload", buf.Bytes())
	if pubRes.Err() != nil {
		fmt.Println("err:", pubRes.Err())
		http.Error(w, "Sorry, we had some trouble processing your video. Please try again", http.StatusInternalServerError)
	}

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResponseOkay{Success: true})
}

func main() {
	port := "5001"
	http.HandleFunc("/upload", upload)
	fmt.Printf("Server running on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
