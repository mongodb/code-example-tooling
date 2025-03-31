package services

import (
	"cloud.google.com/go/logging"
	"context"
	"log"
)

var googleInfoLogger *log.Logger
var googleWarningLogger *log.Logger
var googleErrorLogger *log.Logger
var googleCriticalLogger *log.Logger

func InitializeGoogleLogger() {

	projectID := "github-copy-code-examples"

	loggingClient, err := logging.NewClient(context.Background(), projectID)
	if err != nil {
		log.Printf("Failed to create loggingClient: %v\n", err)
		return
	}
	//defer loggingClient.Close()
	logName := "copy-copier-log"
	googleInfoLogger = loggingClient.Logger(logName).StandardLogger(logging.Info)
	googleWarningLogger = loggingClient.Logger(logName).StandardLogger(logging.Warning)
	googleErrorLogger = loggingClient.Logger(logName).StandardLogger(logging.Error)
	googleCriticalLogger = loggingClient.Logger(logName).StandardLogger(logging.Critical)
}

func LogInfo(message string) {
	if googleInfoLogger != nil {
		googleInfoLogger.Println(message)
	}
	log.Println(message)

}
func LogWarning(message string) {
	if googleWarningLogger != nil {
		googleWarningLogger.Println(message)
	}
	log.Println(message)

}
func LogError(message string) {
	if googleErrorLogger != nil {
		googleErrorLogger.Println(message)
	}
	log.Println(message)

}
func LogCritical(message string) {
	if googleCriticalLogger != nil {
		googleCriticalLogger.Println(message)
	}
	log.Println(message)
}
