package snooty

import (
	"fmt"
	"log"
	"os"
)

func LoadJsonTestDataFromFile(filename string) []byte {
	testFile := fmt.Sprintf("./test-data/%s", filename)
	data, err := os.ReadFile(testFile)
	if err != nil {
		log.Fatalf("Failed to read test data file: %v", err)
	}
	return data
}
