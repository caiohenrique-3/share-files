package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	log.Println("/upload/ hit")

	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	clientFile, handler, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", 400)
		return
	}
	defer clientFile.Close()

	log.Printf("Uploaded File: %+v\n", handler.Filename)
	log.Printf("File Size: %+v\n", handler.Size)

	fileBytes, err := io.ReadAll(clientFile)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
	}

	dst, err := os.Create("./user-uploads/" + handler.Filename)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer dst.Close()

	dst.Write(fileBytes)

	fmt.Fprintf(w, "Successfully uploaded file\n")
}

func main() {
	dirPath := filepath.Join(".", "user-uploads")
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		log.Println(err)

		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("user-uploads dir created")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/upload", uploadFile)

	log.Println("Listening on :9090...")
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal(err)
	}
}
