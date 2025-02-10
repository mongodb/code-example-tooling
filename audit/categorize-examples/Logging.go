package main

import (
	"fmt"
	"time"
)

func LogStartInfoToConsole(startTime time.Time, fileCount int) {
	fmt.Printf("Processing %d files for %s project\n", fileCount, ProjectName)
	fmt.Println("Starting at ", startTime)
	// On an M1 Max laptop from 2021 w/64GB of RAM, a single file takes ~750000000 to process
	// Adjust processing time as needed based on the hardware running this program
	//var processingTime = 738000000 // on DC personal laptop
	var processingTime = 1100000000
	var timeForJob = time.Duration(fileCount * processingTime)
	fmt.Printf("Estimated time to run: %s\n", timeForJob)
}

func LogFinishInfoToConsole(startTime time.Time, filesProcessed int) {
	endTime := time.Now()
	fmt.Println("Finished at ", endTime)
	fmt.Println("Completed in ", endTime.Sub(startTime))
	fmt.Println("Total snippets processed: ", filesProcessed)
}
