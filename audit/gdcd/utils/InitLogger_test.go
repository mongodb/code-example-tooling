package utils

import (
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

func TestInitLogger_CreatesLogDir(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	// Create a nested test directory path
	nestedLogDir := filepath.Join(testDir, "nested", "logs")
	f, err := InitLogger(nestedLogDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer f.Close()

	// Check nested dir exists
	_, err = os.Stat(nestedLogDir)
	if err != nil {
		t.Fatalf("log directory wasn't created: %v", err)
	}
}

func TestInitLogger_CreatesTimestampedLogFile(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	f, err := InitLogger(testDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer f.Close()

	// Check filename matches expected pattern
	filename := filepath.Base(f.Name())
	pattern := `^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-app\.log$`
	matched, err := regexp.MatchString(pattern, filename)
	if err != nil {
		t.Fatalf("regexp error: %v", err)
	}
	if !matched {
		t.Errorf("log filename %q doesn't match expected pattern %q", filename, pattern)
	}
}

func TestInitLogger_FailsWhenDirCannotBeCreated(t *testing.T) {
	// Should fail when dir creation requires elevated permissions
	restrictedDir := "/root/logs"
	if os.Geteuid() == 0 {
		t.Skip("Test can't fail if running as root")
	}

	_, err := InitLogger(restrictedDir)
	if err == nil {
		t.Fatal("expected error when log directory cannot be created, got nil")
	}
}

func TestInitLogger_FailsWhenFileCannotBeCreated(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	// Should fail if directory has the same name as log file
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

func TestInitLogger_WriteToLogFile(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	// Resets destination for logger output after the test
	originalOutput := log.Writer()
	defer log.SetOutput(originalOutput)

	f, err := InitLogger(testDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer f.Close()

	// Write to the log after initializing
	testMessage := "test log message"
	log.Println(testMessage)

	// Read log file content
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("couldn't read log file: %v", err)
	}

	// Log should contain the expected test message
	if !regexp.MustCompile(testMessage).Match(content) {
		t.Errorf("log file doesn't contain expected message; got %q", string(content))
	}
}

func TestInitLogger_UsesCorrectFilePermissions(t *testing.T) {
	cleanupLogs(t)
	defer cleanupLogs(t)

	f, err := InitLogger(testDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer f.Close()

	// Get file info
	fileInfo, err := os.Stat(f.Name())
	if err != nil {
		t.Fatalf("couldn't get file info: %v", err)
	}

	// Check permission mode (0644 in octal)
	expectedMode := os.FileMode(0o644)
	if fileInfo.Mode().Perm() != expectedMode {
		t.Errorf("expected file permissions %v, got %v", expectedMode, fileInfo.Mode().Perm())
	}
}
