package main

import (
	"flag"
	"fmt"
	"github.com/thompsch/app-tester/configs"
	. "github.com/thompsch/app-tester/services"
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

	configs.EnvFile = envFile
	configs.LoadEnvironment()
	SetupWebServerAndListen()

}
