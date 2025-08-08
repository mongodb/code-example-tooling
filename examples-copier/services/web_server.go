package services

import (
	"fmt"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/pkg/errors"
	"net/http"
	"os"
)

func SetupWebServerAndListen() {
	// Get environment file path from command line flag or environment variable
	envFilePath := os.Getenv("ENV_FILE")

	_, err := configs.LoadEnvironment(envFilePath)
	if err != nil {
		LogCritical(fmt.Sprintf("Failed to load environment: %v", err))
		return
	}

	InitializeGoogleLogger()
	http.HandleFunc(configs.WebserverPath, ParseWebhookData)
	port := configs.Port
	if port == "" {
		port = ":8080" // default port
	} else {
		port = ":" + port
	}

	LogInfo(fmt.Sprintf("Starting web server on port %s", port))

	e := http.ListenAndServe(port, nil)
	if e != nil && !errors.Is(e, http.ErrServerClosed) {
		LogCritical(fmt.Sprintf("Error starting server: %v", e))
	} else {
		LogInfo(fmt.Sprintf("Web server listening on " + configs.WebserverPath))
	}
}
