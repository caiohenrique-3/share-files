package main

import (
	"log"
	"os"
	"path/filepath"
)

func checkIfPathExists(path string) error {
	dirPath := filepath.Join(path)
	_, err := os.Stat(dirPath)
	return err
}

func createDir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Dir created:", "\""+path+"\"")
}

func createFile(filename string) (os.File, error) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("File created:", "\""+filename+"\"")
	return *file, err
}

func createFilesOnStartup() {
	err := checkIfPathExists("./user-uploads")
	if err != nil {
		createDir("./user-uploads")
	}

	err = checkIfPathExists("./data.json")
	if err != nil {
		createFile("./data.json")
	}
}
