package main

import (
	"flag"
	"fmt"

	"github.com/mongodb/code-example-tooling/code-copier/configs"
	"github.com/mongodb/code-example-tooling/code-copier/services"
)

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
