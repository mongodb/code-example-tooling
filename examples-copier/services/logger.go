package services

import (
	"context"
	"log"
	"os"
	"strings"

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

	projectId := configs.GoogleCloudProjectId

	client, err := logging.NewClient(context.Background(), projectId)
	if err != nil {
		log.Printf("[WARN] Failed to create Google logging client: %v\n", err)
		gcpLoggingEnabled = false
		return
	}
	googleLoggingClient = client
	gcpLoggingEnabled = true

	logName := configs.CopierLogName
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
