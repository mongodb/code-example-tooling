package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

func containsString(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func main() {
	isDriverProject := false
	driversProjects := []string{"c", "cpp-driver", "csharp", "java", "java-rs", "kotlin", "kotlin-sync", "laravel", "mongoid", "node", "php-library", "pymongo", "pymongo-arrow", "ruby-driver", "rust", "scala"}
	for _, driver := range driversProjects {
		if ProjectName == driver {
			isDriverProject = true
		}
	}
	startTime := time.Now()
	files := GetFiles()
	totalFileCount := len(files)
	LogStartInfoToConsole(startTime, totalFileCount)

	var snippets []SnippetInfo
	counts := make(map[string]map[string]int)
	llmCategorizedCount := 0
	stringMatchedCount := 0
	filesProcessed := 0

	// To change the model, use a different model's string name here
	llm, err := ollama.New(ollama.WithModel(MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()

	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("failed to read file: %v\n", err)
			return
		}
		// Find the starting index of the project name to strip the earlier parts of the filepath
		startIndex := strings.Index(file, ProjectName)
		pagePath := file[startIndex:]
		ext := filepath.Ext(file)
		if !strings.Contains(file, ".DS_Store") {
			lang := GetLangFromExtension(ext)

			category, llmCategorized := ProcessSnippet(string(contents), lang, llm, ctx, isDriverProject)

			details := SnippetInfo{
				Page:           pagePath,
				Category:       category,
				Language:       lang,
				LLMCategorized: llmCategorized,
			}
			snippets = append(snippets, details)
			if _, exists := counts[details.Category]; !exists {
				counts[details.Category] = make(map[string]int)
			}
			// Increment the language count for the specific category
			counts[details.Category][details.Language]++
			if llmCategorized {
				llmCategorizedCount++
			} else if llmCategorized == false {
				stringMatchedCount++
			}
		}
		filesProcessed++
		if filesProcessed%100 == 0 {
			fmt.Println("Processed ", filesProcessed, " snippets")
		}
	}

	WriteSnippetReport(snippets, ProjectName)
	WriteCategoryCountsReport(totalFileCount, counts, llmCategorizedCount, stringMatchedCount, ProjectName, isDriverProject)
	LogFinishInfoToConsole(startTime, filesProcessed)
}
