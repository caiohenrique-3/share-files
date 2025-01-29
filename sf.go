package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type Data struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

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

	log.Printf("Name: %+v\n", handler.Filename)
	log.Printf("Size: %+v\n", handler.Size)

	fileBytes, err := io.ReadAll(clientFile)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	dst, err := os.Create("./user-uploads/" + handler.Filename)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer dst.Close()

	dst.Write(fileBytes)

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Trim newline from UUID output
	uuidStr := string(uuid)
	uuidStr = uuidStr[:len(uuidStr)-1]

	log.Println("UUID:", uuidStr)

	err = checkFile("./data.json")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	jsonData := Data{uuidStr, handler.Filename}
	file, err := os.OpenFile("data.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer file.Close()

	var data []Data
	decoder := json.NewDecoder(file)
	decoder.Decode(&data)

	data = append(data, jsonData)

	file.Seek(0, 0)
	encoder := json.NewEncoder(file)
	encoder.Encode(data)

	fmt.Fprintln(w, "Successfully uploaded file!")
	fmt.Fprintln(w, "/download/"+uuidStr)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	log.Println("/download/ hit")

	uuid := r.URL.Path[len("/download/"):]

	var data []Data
	file, err := os.Open("./data.json")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.Decode(&data)

	for _, item := range data {
		if item.ID == uuid {
			filePath := "./user-uploads/" + item.Filename
			fileToDownload, err := os.Open(filePath)
			if err != nil {
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return
			}
			defer fileToDownload.Close()

			w.Header().Set("Content-Disposition", "attachment; filename="+item.Filename)
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", getSize(filePath)))

			io.Copy(w, fileToDownload)
			return
		}
	}

	http.Error(w, "File Not Found", 404)
}

func getSize(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

func checkFile(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			return err
		}
	}
	return nil
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

		log.Println("user-uploads dir created")
	}

	_, err = os.Stat("./data.json")
	if os.IsNotExist(err) {
		log.Println(err)

		_, err := os.Create("./data.json")
		if err != nil {
			log.Fatal(err)
		}

		log.Println("data.json file created")
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/download/", downloadFile)

	log.Println("Listening on :9090...")
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal(err)
	}
}
