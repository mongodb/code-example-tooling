package main

import (
	"log"
	"os"
	"path/filepath"
)

// GetFiles traverses directories recursively from the startDirPath and adds file paths to an array of strings that it
// passes back to main.go to read into memory and categorize
func GetFiles() []string {
	// To traverse a different directory on your file system, change the path in `constants.go`
	// or create a new path variable that isn't relative to the root of this repo
	fullFilePath := SnippetsStartDirectory + ProjectName
	startDirPath, _ := filepath.Abs(fullFilePath)

	fileList := make([]string, 0)
	e := filepath.Walk(startDirPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("failed to traverse the file path: %v", err)
			return err
		}
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})

	if e != nil {
		log.Fatalf("failed to traverse the file path: %v", e)
	}

	return fileList
}
