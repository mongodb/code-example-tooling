package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

const testDir = "./temp"

func cleanupLogs(t *testing.T) {
	err := os.RemoveAll(testDir)
	if err != nil {
		t.Fatalf("failed to clean up temp directory: %v", err)
	}
}

func TestInitLogger_FailsWhenFileCannotBeCreated(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	// Create a directory with the same name as the log file to cause a failure
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	logFilePath := filepath.Join(testDir, timestamp+"-app.log")
	if err := os.MkdirAll(logFilePath, 0o755); err != nil {
		t.Fatalf("couldn't create temp directory: %v", err)
	}

	_, err := InitLogger(testDir)
	if err == nil {
		t.Fatal("expected error when log file cannot be created, got nil")
	}
}

func TestInitLogger_WritesToConsoleAndFile(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	// Capture console output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f, err := InitLogger(testDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer func() {
		f.Close()
		os.Stdout = oldStdout
	}()

	// Write a log message
	logMessage := "console-and-file-test"
	log.Println(logMessage)

	// Capture console output
	w.Close()
	consoleOutput, _ := io.ReadAll(r)

	// Read file contents
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}

	// Verify log message is in both console and file
	if !regexp.MustCompile(logMessage).Match(consoleOutput) {
		t.Errorf("expected log entry in console; got %q", string(consoleOutput))
	}
	if !regexp.MustCompile(logMessage).Match(data) {
		t.Errorf("expected log entry in file; got %q", string(data))
	}
}
