package main

import (
	"fmt"
	"net/http"
)

func upload(w http.ResponseWriter, r *http.Request) {
	// 1. Get file from request
	// 2. Check file mime type (this should also avoid harmful uploads if done correctly)
	//	- video/mp4
	//	- video/mpeg
	//	- video/x-msvideo
	//	- video/webm
	//	- video/3gpp
	//	- video/3gpp2
	// 3. Send to queue if valid
}

func main() {
	port := "5001"
	http.HandleFunc("/upload", upload)
	fmt.Printf("Server running on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
