package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var EnvFile string
var Port string
var RepoName string
var RepoOwner string
var AppClientId string
var InstallationId string
var CommiterName string
var CommiterEmail string
var ConfigFile string
var DeprecationFile string
var WebserverPath string

func LoadEnvironment() {
	if RepoOwner == "" || RepoName == "" || AppClientId == "" || InstallationId == "" {
		err := godotenv.Load(EnvFile)
		if err != nil {
			log.Fatal("Error loading env file")
		}

		Port = os.Getenv("PORT")
		RepoName = os.Getenv("REPO_NAME")
		RepoOwner = os.Getenv("REPO_OWNER")
		AppClientId = os.Getenv("GITHUB_APP_CLIENT_ID")
		InstallationId = os.Getenv("INSTALLATION_ID")
		CommiterName = os.Getenv("COMMITER_NAME")
		CommiterEmail = os.Getenv("COMMITER_EMAIL")
		ConfigFile = os.Getenv("CONFIG_FILE")
		DeprecationFile = os.Getenv("DEPRECATION_FILE")
		WebserverPath = os.Getenv("WEBSERVER_PATH")
	}
}
