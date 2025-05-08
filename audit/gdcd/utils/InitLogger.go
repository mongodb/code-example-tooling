package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// InitLogger sets up the directory and new log file, but fails fast if dir or file can't be created.
func InitLogger(logDir string) (*os.File, error) {
	// Make sure dir exists, and create if needed
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating %q: %w", logDir, err)
	}

	// Build timestamped log filename
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	logFile := filepath.Join(logDir, timestamp+"-app.log")

	// Create the log file only if it doesn't already exist (don't overwrite logs!)
	f, err := os.OpenFile(
		logFile,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("couldn't create new log file %q: %v", logFile, err)
	}

	// Set up multi-writer to send output to both the console and the log file
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	return f, nil
}
