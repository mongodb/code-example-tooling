package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/services"
	"golang.org/x/oauth2"
)

func init() {
	// Initialize the GitHub client early in application startup
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Set the global GitHub client using the wrapper
	services.GitHubClient = &services.GitHubWrapper{
		Client: client.Repositories,
	}
}

func main() {
	// Take the environment file path from command line arguments or default to "./configs/.env"

	var envFile string
	flag.StringVar(&envFile, "env", "./configs/.env", "env file")
	help := flag.Bool("help", false, "show help")
	flag.Parse()
	if help != nil && *help == true {
		fmt.Println("Usage: go run app -env [path to env file]")
		return
	}

	_, err := configs.LoadEnvironment(envFile)
	if err != nil {
		fmt.Printf("Error loading environment: %v\n", err)
		return
	}
	services.ConfigurePermissions()
	services.SetupWebServerAndListen()
}
