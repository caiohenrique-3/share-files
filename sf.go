package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	log.Print("File upload endpoint hit")

	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	tempFile, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error retrieving file!")
		fmt.Println(err)
		return
	}
	defer tempFile.Close()

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	fileBytes, err := io.ReadAll(tempFile)
	if err != nil {
		fmt.Println(err)
	}

	dst, err := os.Create("./uploads/" + handler.Filename)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return
	}
	defer dst.Close()

	dst.Write(fileBytes)

	fmt.Fprintf(w, "Successfully uploaded file\n")

}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/upload", uploadFile)

	log.Print("Listening on :9090...")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal(err)
	}
}
