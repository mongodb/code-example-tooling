package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
)

var googleInfoLogger *log.Logger
var googleWarningLogger *log.Logger
var googleErrorLogger *log.Logger
var googleCriticalLogger *log.Logger

// keep a reference to allow flushing/closing and to avoid re-initialization
var googleLoggingClient *logging.Client
var gcpLoggingEnabled bool

// InitializeGoogleLogger sets up Google Cloud Logging level loggers if not disabled.
// It is safe to call multiple times; initialization will only occur once per process.
func InitializeGoogleLogger() {
	// Allow disabling cloud logging for local/dev via env.
	if isCloudLoggingDisabled() {
		gcpLoggingEnabled = false
		return
	}
	if googleLoggingClient != nil {
		// already initialized
		gcpLoggingEnabled = true
		return
	}

	projectId := os.Getenv(configs.GoogleCloudProjectId)
	if projectId == "" {
		log.Printf("[WARN] GOOGLE_CLOUD_PROJECT_ID not set, disabling cloud logging\n")
		gcpLoggingEnabled = false
		return
	}

	client, err := logging.NewClient(context.Background(), projectId)
	if err != nil {
		log.Printf("[WARN] Failed to create Google logging client: %v\n", err)
		gcpLoggingEnabled = false
		return
	}
	googleLoggingClient = client
	gcpLoggingEnabled = true

	logName := os.Getenv(configs.CopierLogName)
	if logName == "" {
		logName = "code-copier-log" // fallback default
	}
	googleInfoLogger = client.Logger(logName).StandardLogger(logging.Info)
	googleWarningLogger = client.Logger(logName).StandardLogger(logging.Warning)
	googleErrorLogger = client.Logger(logName).StandardLogger(logging.Error)
	googleCriticalLogger = client.Logger(logName).StandardLogger(logging.Critical)
}

// CloseGoogleLogger flushes and closes the underlying Google logging client, if any.
func CloseGoogleLogger() {
	if googleLoggingClient != nil {
		_ = googleLoggingClient.Close()
	}
}

// LogDebug writes debug logs only when LOG_LEVEL=debug or COPIER_DEBUG=true.
func LogDebug(message string) {
	if !isDebugEnabled() {
		return
	}
	// Mirror to GCP as info if available, plus prefix to stdout
	if googleInfoLogger != nil && gcpLoggingEnabled {
		googleInfoLogger.Println("[DEBUG] " + message)
	}
	log.Println("[DEBUG] " + message)
}

func LogInfo(message string) {
	if googleInfoLogger != nil && gcpLoggingEnabled {
		googleInfoLogger.Println(message)
	}
	log.Println("[INFO] " + message)
}

func LogWarning(message string) {
	if googleWarningLogger != nil && gcpLoggingEnabled {
		googleWarningLogger.Println(message)
	}
	log.Println("[WARN] " + message)
}

func LogError(message string) {
	if googleErrorLogger != nil && gcpLoggingEnabled {
		googleErrorLogger.Println(message)
	}
	log.Println("[ERROR] " + message)
}

func LogCritical(message string) {
	if googleCriticalLogger != nil && gcpLoggingEnabled {
		googleCriticalLogger.Println(message)
	}
	log.Println("[CRITICAL] " + message)
}

func isDebugEnabled() bool {
	if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
		return true
	}
	return strings.EqualFold(os.Getenv("COPIER_DEBUG"), "true")
}

func isCloudLoggingDisabled() bool {
	return strings.EqualFold(os.Getenv("COPIER_DISABLE_CLOUD_LOGGING"), "true")
}


// Context-aware logging functions

// LogInfoCtx logs an info message with context and additional fields
func LogInfoCtx(ctx context.Context, message string, fields map[string]interface{}) {
	msg := formatLogMessage(ctx, message, fields)
	LogInfo(msg)
}

// LogWarningCtx logs a warning message with context and additional fields
func LogWarningCtx(ctx context.Context, message string, fields map[string]interface{}) {
	msg := formatLogMessage(ctx, message, fields)
	LogWarning(msg)
}

// LogErrorCtx logs an error message with context and additional fields
func LogErrorCtx(ctx context.Context, message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	msg := formatLogMessage(ctx, message, fields)
	LogError(msg)
}

// LogWebhookOperation logs webhook-related operations
func LogWebhookOperation(ctx context.Context, operation string, message string, err error, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	allFields["operation"] = operation

	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			allFields[k] = v
		}
	}

	if err != nil {
		LogErrorCtx(ctx, message, err, allFields)
	} else {
		LogInfoCtx(ctx, message, allFields)
	}
}

// LogFileOperation logs file-related operations
func LogFileOperation(ctx context.Context, operation string, sourcePath string, targetRepo string, message string, err error, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	allFields["operation"] = operation
	allFields["source_path"] = sourcePath
	if targetRepo != "" {
		allFields["target_repo"] = targetRepo
	}

	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			allFields[k] = v
		}
	}

	if err != nil {
		LogErrorCtx(ctx, message, err, allFields)
	} else {
		LogInfoCtx(ctx, message, allFields)
	}
}

// LogAndReturnError logs an error and returns
func LogAndReturnError(ctx context.Context, operation string, message string, err error) {
	LogErrorCtx(ctx, message, err, map[string]interface{}{
		"operation": operation,
	})
}

// formatLogMessage formats a log message with context and fields
func formatLogMessage(ctx context.Context, message string, fields map[string]interface{}) string {
	if fields == nil || len(fields) == 0 {
		return message
	}

	// Convert fields to JSON for structured logging
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		return fmt.Sprintf("%s | fields_error=%v", message, err)
	}

	return fmt.Sprintf("%s | %s", message, string(fieldsJSON))
}

// WithRequestID adds a request ID to the context and returns both the context and the ID
func WithRequestID(r *http.Request) (context.Context, string) {
	// Generate a simple request ID
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Add to context
	ctx := context.WithValue(r.Context(), "request_id", requestID)

	return ctx, requestID
}
