package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
)

func main() {
	// Define subcommands
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	validateFile := validateCmd.String("config", "", "Path to config file (required)")
	validateVerbose := validateCmd.Bool("v", false, "Verbose output")

	testPatternCmd := flag.NewFlagSet("test-pattern", flag.ExitOnError)
	patternType := testPatternCmd.String("type", "prefix", "Pattern type: prefix, glob, or regex")
	pattern := testPatternCmd.String("pattern", "", "Pattern to test (required)")
	filePath := testPatternCmd.String("file", "", "File path to test against (required)")

	testTransformCmd := flag.NewFlagSet("test-transform", flag.ExitOnError)
	transformSource := testTransformCmd.String("source", "", "Source file path (required)")
	transformTemplate := testTransformCmd.String("template", "", "Transform template (required)")
	transformVars := testTransformCmd.String("vars", "", "Variables as key=value pairs, comma-separated")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initTemplate := initCmd.String("template", "basic", "Template to use: basic, glob, or regex")
	initOutput := initCmd.String("output", "workflow-config.yaml", "Output file path")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		validateCmd.Parse(os.Args[2:])
		if *validateFile == "" {
			fmt.Println("Error: -config is required")
			validateCmd.Usage()
			os.Exit(1)
		}
		validateConfig(*validateFile, *validateVerbose)

	case "test-pattern":
		testPatternCmd.Parse(os.Args[2:])
		if *pattern == "" || *filePath == "" {
			fmt.Println("Error: -pattern and -file are required")
			testPatternCmd.Usage()
			os.Exit(1)
		}
		testPattern(*patternType, *pattern, *filePath)

	case "test-transform":
		testTransformCmd.Parse(os.Args[2:])
		if *transformSource == "" || *transformTemplate == "" {
			fmt.Println("Error: -source and -template are required")
			testTransformCmd.Usage()
			os.Exit(1)
		}
		testTransform(*transformSource, *transformTemplate, *transformVars)

	case "init":
		initCmd.Parse(os.Args[2:])
		initConfig(*initTemplate, *initOutput)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Config Validator - Validate and test copier workflow configurations")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  config-validator <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  validate       Validate a workflow configuration file")
	fmt.Println("  test-pattern   Test a pattern against a file path")
	fmt.Println("  test-transform Test a path transformation")
	fmt.Println("  init           Initialize a new workflow config file from template")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  config-validator validate -config .copier/workflows/config.yaml -v")
	fmt.Println("  config-validator test-pattern -type glob -pattern 'examples/**/*.go' -file 'examples/go/main.go'")
	fmt.Println("  config-validator test-transform -source 'examples/go/main.go' -template 'code/${filename}'")
	fmt.Println("  config-validator init -template basic -output workflow-config.yaml")
}

func validateConfig(configFile string, verbose bool) {
	content, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("❌ Error reading config file: %v\n", err)
		os.Exit(1)
	}

	loader := services.NewConfigLoader()
	config, err := loader.LoadConfigFromContent(string(content), configFile)
	if err != nil {
		fmt.Printf("❌ Config validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Configuration is valid!")

	if verbose {
		fmt.Println()
		fmt.Printf("Number of Workflows: %d\n", len(config.Workflows))
		fmt.Println()

		for i, workflow := range config.Workflows {
			fmt.Printf("Workflow %d: %s\n", i+1, workflow.Name)
			fmt.Printf("  Source: %s @ %s\n", workflow.Source.Repo, workflow.Source.Branch)
			fmt.Printf("  Destination: %s @ %s\n", workflow.Destination.Repo, workflow.Destination.Branch)
			fmt.Printf("  Transformations: %d\n", len(workflow.Transformations))
			fmt.Printf("  Commit Strategy: %s\n", workflow.CommitStrategy.Type)
			if workflow.DeprecationCheck != nil && workflow.DeprecationCheck.Enabled {
				fmt.Printf("  Deprecation Tracking: enabled (%s)\n", workflow.DeprecationCheck.File)
			}
			fmt.Println()
		}
	}
}

func testPattern(patternType, pattern, filePath string) {
	var pt types.PatternType
	switch patternType {
	case "prefix":
		pt = types.PatternTypePrefix
	case "glob":
		pt = types.PatternTypeGlob
	case "regex":
		pt = types.PatternTypeRegex
	default:
		fmt.Printf("❌ Invalid pattern type: %s (must be prefix, glob, or regex)\n", patternType)
		os.Exit(1)
	}

	validator := services.NewConfigValidator()
	result, err := validator.TestPattern(pt, pattern, filePath)
	if err != nil {
		fmt.Printf("❌ Pattern validation error: %v\n", err)
		os.Exit(1)
	}

	if result.Matched {
		fmt.Println("✅ Pattern matched!")
		if len(result.Variables) > 0 {
			fmt.Println("\nExtracted variables:")
			for key, value := range result.Variables {
				fmt.Printf("  %s = %s\n", key, value)
			}
		}
	} else {
		fmt.Println("❌ Pattern did not match")
		os.Exit(1)
	}
}

func testTransform(source, template, varsStr string) {
	variables := make(map[string]string)
	if varsStr != "" {
		pairs := strings.Split(varsStr, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				variables[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	validator := services.NewConfigValidator()
	result, err := validator.TestTransform(source, template, variables)
	if err != nil {
		fmt.Printf("❌ Transform error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Transform successful!")
	fmt.Printf("Source: %s\n", source)
	fmt.Printf("Result: %s\n", result)
}

func initConfig(templateName, output string) {
	// Simple workflow config template
	template := `# Workflow Configuration
# This file defines workflows for copying code examples between repositories

workflows:
  - name: "example-workflow"
    source:
      repo: "mongodb/source-repo"
      branch: "main"
      path: "examples"
    destination:
      repo: "mongodb/dest-repo"
      branch: "main"
    transformations:
      - move:
          from: "examples"
          to: "code-examples"
    commit_strategy:
      type: "pr"
      pr_title: "Update code examples"
      pr_body: "Automated update from source repository"
`

	err := os.WriteFile(output, []byte(template), 0644)
	if err != nil {
		fmt.Printf("❌ Error writing config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created workflow config file: %s\n", output)
	fmt.Println("Edit this file to configure your workflows")
}
