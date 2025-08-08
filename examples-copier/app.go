package main

import (
	"flag"
	"fmt"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
	. "github.com/mongodb/code-example-tooling/code-copier/services"
)

func main() {
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
		return
	}

	ConfigurePermissions()
	SetupWebServerAndListen()

}
