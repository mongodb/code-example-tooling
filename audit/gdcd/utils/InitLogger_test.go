package utils

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var testDir = "testDir"

// Helper to remove any leftover logs directory before/after tests.
func cleanupLogs(t *testing.T) {
	if err := os.RemoveAll(testDir); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}

func TestInitLogger_Success(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	// Call initLogger
	f, err := InitLogger(testDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer func() {
		f.Close()
	}()

	// logs directory must now exist
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("exists but is not a directory")
	}

	// File name must match the timestamped pattern _app.log
	filename := filepath.Base(f.Name())
	match, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-app\.log$`, filename)
	if !match {
		t.Errorf("unexpected log file name: %q", filename)
	}

	// Test that writing to the global logger goes into the file
	log.Println("hello-test")

	// Read file contents
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	if !regexp.MustCompile(`hello-test`).Match(data) {
		t.Errorf("expected log entry in file; got %q", string(data))
	}
}

func TestInitLogger_FailsWhenLogsIsAFile(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	// Create a file with same name so MkdirAll will fail
	if err := os.WriteFile(testDir, []byte{}, 0644); err != nil {
		t.Fatalf("couldn't create dummy logs file: %v", err)
	}

	_, err := InitLogger(testDir)
	if err == nil {
		t.Fatal("expected error when 'logs' exists as a file, got nil")
	}
}
