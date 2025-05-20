package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// InitLogger sets up the directory and new log file, returning the file and any error
func InitLogger(logDir string) (*os.File, error) {
	// Make sure dir exists, and create if needed
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating log directory %q: %w", logDir, err)
	}

	// Build timestamped log filename
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	logFile := filepath.Join(logDir, timestamp+"-app.log")

	// Create the log file only if it doesn't already exist (rare edge case)
	f, err := os.OpenFile(
		logFile,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("creating log file %q: %w", logFile, err)
	}

	// Send output to log file
	log.SetOutput(f)

	return f, nil
}
