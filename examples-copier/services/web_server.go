package services

import (
	"fmt"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
)

// SetupWebServerAndListen sets up the web server and listens for incoming webhook requests.
func SetupWebServerAndListen() {
	// Get environment file path from command line flag or environment variable
	envFilePath := os.Getenv("ENV_FILE")

	_, err := configs.LoadEnvironment(envFilePath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to load environment"))
	}

 InitializeGoogleLogger()
 path := os.Getenv(configs.WebserverPath)
	if path == "" {
		path = configs.NewConfig().WebserverPath
	}
	http.HandleFunc(path, ParseWebhookData)
	port := os.Getenv(configs.Port)
	if port == "" {
		port = ":8080" // default port
	} else {
		port = ":" + port
	}

	LogInfo(fmt.Sprintf("Starting web server on port %s; path %s", port, path))

	e := http.ListenAndServe(port, nil)
	if e != nil && !errors.Is(e, http.ErrServerClosed) {
		log.Fatal(errors.Wrap(e, "Error starting server"))
	} else {
		LogInfo(fmt.Sprintf("Web server listening on " + path))
	}
}
