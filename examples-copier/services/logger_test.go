package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLogDebug(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      string
		copierDebug   string
		message       string
		shouldLog     bool
	}{
		{
			name:        "debug enabled via LOG_LEVEL",
			logLevel:    "debug",
			copierDebug: "",
			message:     "test debug message",
			shouldLog:   true,
		},
		{
			name:        "debug enabled via COPIER_DEBUG",
			logLevel:    "",
			copierDebug: "true",
			message:     "test debug message",
			shouldLog:   true,
		},
		{
			name:        "debug disabled",
			logLevel:    "info",
			copierDebug: "false",
			message:     "test debug message",
			shouldLog:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.logLevel != "" {
				os.Setenv("LOG_LEVEL", tt.logLevel)
				defer os.Unsetenv("LOG_LEVEL")
			}
			if tt.copierDebug != "" {
				os.Setenv("COPIER_DEBUG", tt.copierDebug)
				defer os.Unsetenv("COPIER_DEBUG")
			}

			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			LogDebug(tt.message)

			output := buf.String()
			if tt.shouldLog {
				if !strings.Contains(output, "[DEBUG]") {
					t.Error("Expected [DEBUG] prefix in output")
				}
				if !strings.Contains(output, tt.message) {
					t.Errorf("Expected message %q in output", tt.message)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no output, got: %s", output)
				}
			}
		})
	}
}

func TestLogInfo(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	message := "test info message"
	LogInfo(message)

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Error("Expected [INFO] prefix in output")
	}
	if !strings.Contains(output, message) {
		t.Errorf("Expected message %q in output", message)
	}
}

func TestLogWarning(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	message := "test warning message"
	LogWarning(message)

	output := buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Error("Expected [WARN] prefix in output")
	}
	if !strings.Contains(output, message) {
		t.Errorf("Expected message %q in output", message)
	}
}

func TestLogError(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	message := "test error message"
	LogError(message)

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Error("Expected [ERROR] prefix in output")
	}
	if !strings.Contains(output, message) {
		t.Errorf("Expected message %q in output", message)
	}
}

func TestLogCritical(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	message := "test critical message"
	LogCritical(message)

	output := buf.String()
	if !strings.Contains(output, "[CRITICAL]") {
		t.Error("Expected [CRITICAL] prefix in output")
	}
	if !strings.Contains(output, message) {
		t.Errorf("Expected message %q in output", message)
	}
}

func TestLogInfoCtx(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	ctx := context.Background()
	message := "test context message"
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	LogInfoCtx(ctx, message, fields)

	output := buf.String()
	if !strings.Contains(output, message) {
		t.Errorf("Expected message %q in output", message)
	}
	if !strings.Contains(output, "key1") {
		t.Error("Expected field key1 in output")
	}
	if !strings.Contains(output, "value1") {
		t.Error("Expected field value1 in output")
	}
}

func TestLogWarningCtx(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	ctx := context.Background()
	message := "test warning context"
	fields := map[string]interface{}{
		"warning_type": "test",
	}

	LogWarningCtx(ctx, message, fields)

	output := buf.String()
	if !strings.Contains(output, message) {
		t.Errorf("Expected message %q in output", message)
	}
	if !strings.Contains(output, "warning_type") {
		t.Error("Expected field warning_type in output")
	}
}

func TestLogErrorCtx(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	ctx := context.Background()
	message := "test error context"
	err := fmt.Errorf("test error")
	fields := map[string]interface{}{
		"error_code": 500,
	}

	LogErrorCtx(ctx, message, err, fields)

	output := buf.String()
	if !strings.Contains(output, message) {
		t.Errorf("Expected message %q in output", message)
	}
	if !strings.Contains(output, "test error") {
		t.Error("Expected error message in output")
	}
	if !strings.Contains(output, "error_code") {
		t.Error("Expected field error_code in output")
	}
}

func TestLogWebhookOperation(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		message   string
		err       error
		wantLevel string
	}{
		{
			name:      "successful operation",
			operation: "webhook_received",
			message:   "webhook processed",
			err:       nil,
			wantLevel: "[INFO]",
		},
		{
			name:      "failed operation",
			operation: "webhook_parse",
			message:   "failed to parse webhook",
			err:       fmt.Errorf("parse error"),
			wantLevel: "[ERROR]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			ctx := context.Background()
			LogWebhookOperation(ctx, tt.operation, tt.message, tt.err)

			output := buf.String()
			if !strings.Contains(output, tt.wantLevel) {
				t.Errorf("Expected %s level in output", tt.wantLevel)
			}
			if !strings.Contains(output, tt.message) {
				t.Errorf("Expected message %q in output", tt.message)
			}
			if !strings.Contains(output, tt.operation) {
				t.Errorf("Expected operation %q in output", tt.operation)
			}
		})
	}
}

func TestLogFileOperation(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	ctx := context.Background()
	LogFileOperation(ctx, "copy", "source/file.go", "target/repo", "file copied", nil)

	output := buf.String()
	if !strings.Contains(output, "copy") {
		t.Error("Expected operation 'copy' in output")
	}
	if !strings.Contains(output, "source/file.go") {
		t.Error("Expected source path in output")
	}
	if !strings.Contains(output, "target/repo") {
		t.Error("Expected target repo in output")
	}
}

func TestWithRequestID(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	
	ctx, requestID := WithRequestID(req)
	
	if requestID == "" {
		t.Error("Expected non-empty request ID")
	}

	// Check that request ID is in context
	ctxValue := ctx.Value("request_id")
	if ctxValue == nil {
		t.Error("Expected request_id in context")
	}

	if ctxValue.(string) != requestID {
		t.Error("Context request_id doesn't match returned request ID")
	}
}

func TestFormatLogMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		fields  map[string]interface{}
		want    []string
	}{
		{
			name:    "no fields",
			message: "test message",
			fields:  nil,
			want:    []string{"test message"},
		},
		{
			name:    "with fields",
			message: "test message",
			fields: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			want: []string{"test message", "key1", "value1"},
		},
		{
			name:    "empty fields",
			message: "test message",
			fields:  map[string]interface{}{},
			want:    []string{"test message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := formatLogMessage(ctx, tt.message, tt.fields)

			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("formatLogMessage() missing %q in result: %s", want, result)
				}
			}
		})
	}
}

func TestIsDebugEnabled(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		copierDebug string
		want        bool
	}{
		{"debug via LOG_LEVEL", "debug", "", true},
		{"DEBUG via LOG_LEVEL", "DEBUG", "", true},
		{"debug via COPIER_DEBUG", "", "true", true},
		{"debug via COPIER_DEBUG uppercase", "", "TRUE", true},
		{"not enabled", "info", "false", false},
		{"neither set", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tt.logLevel)
			os.Setenv("COPIER_DEBUG", tt.copierDebug)
			defer os.Unsetenv("LOG_LEVEL")
			defer os.Unsetenv("COPIER_DEBUG")

			got := isDebugEnabled()
			if got != tt.want {
				t.Errorf("isDebugEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCloudLoggingDisabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"disabled lowercase", "true", true},
		{"disabled uppercase", "TRUE", true},
		{"enabled", "false", false},
		{"not set", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("COPIER_DISABLE_CLOUD_LOGGING", tt.value)
			defer os.Unsetenv("COPIER_DISABLE_CLOUD_LOGGING")

			got := isCloudLoggingDisabled()
			if got != tt.want {
				t.Errorf("isCloudLoggingDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

