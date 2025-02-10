package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/tmc/langchaingo/llms/ollama"
)

// TODO: Add tests for remaining examples

func TestCategorizeSnippetAPIMethod(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel(MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	testFilePath, err := filepath.Abs("../categorize-examples/examples/other/insertOne.sh")
	if err != nil {
		log.Fatalf("failed to construct the file path: %v", err)
	}
	contents, err := os.ReadFile(testFilePath)
	if err != nil {
		log.Fatalf("failed to read the file at %v: %v", testFilePath, err)
	}
	category, llmCategorized := ProcessSnippet(string(contents), SHELL, llm, ctx, false)
	if category != SyntaxExample {
		t.Errorf("For category, got %v, want %v", category, SyntaxExample)
	}
	if llmCategorized != true {
		t.Errorf("For llm-categorized, got %v, want %v", llmCategorized, true)
	}
}

func TestCategorizeSnippetAPIMethodWithValues(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel(MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	testFilePath, err := filepath.Abs("../categorize-examples/examples/other/api-method.go")
	if err != nil {
		log.Fatalf("failed to construct the file path: %v", err)
	}
	contents, err := os.ReadFile(testFilePath)
	if err != nil {
		log.Fatalf("failed to read the file at %v: %v", testFilePath, err)
	}
	category, llmCategorized := ProcessSnippet(string(contents), JAVASCRIPT, llm, ctx, false)
	if category != SyntaxExample {
		t.Errorf("For category, got %v, want %v", category, SyntaxExample)
	}
	if llmCategorized != true {
		t.Errorf("For llm-categorized, got %v, want %v", llmCategorized, true)
	}
}

func TestCategorizeConfigExample(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel(MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	testFilePath, err := filepath.Abs("../categorize-examples/examples/other/configExample.yaml")
	if err != nil {
		log.Fatalf("failed to construct the file path: %v", err)
	}
	contents, err := os.ReadFile(testFilePath)
	if err != nil {
		log.Fatalf("failed to read the file at %v: %v", testFilePath, err)
	}
	category, llmCategorized := ProcessSnippet(string(contents), YAML, llm, ctx, false)
	if category != ExampleConfigurationObject {
		t.Errorf("For category, got %v, want %v", category, ExampleConfigurationObject)
	}
	if llmCategorized != true {
		t.Errorf("For llm-categorized, got %v, want %v", llmCategorized, true)
	}
}

func TestCategorizeSimpleReturnExample(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel(MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	testFilePath, err := filepath.Abs("../categorize-examples/examples/other/returnExample.sh")
	if err != nil {
		log.Fatalf("failed to construct the file path: %v", err)
	}
	contents, err := os.ReadFile(testFilePath)
	if err != nil {
		log.Fatalf("failed to read the file at %v: %v", testFilePath, err)
	}
	category, llmCategorized := ProcessSnippet(string(contents), SHELL, llm, ctx, false)
	if category != ExampleReturnObject {
		t.Errorf("For category, got %v, want %v", category, ExampleReturnObject)
	}
	if llmCategorized != false {
		t.Errorf("For llm-categorized, got %v, want %v", llmCategorized, false)
	}
}

// This test is currently failing - the LLMs seem to assess multi return examples as Task-based usage
// Should further tweak prompt until this passes
func TestCategorizeMultiReturnExample(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel(MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	testFilePath, err := filepath.Abs("../categorize-examples/examples/other/runQueriesReturnExample.sh")
	if err != nil {
		log.Fatalf("failed to construct the file path: %v", err)
	}
	contents, err := os.ReadFile(testFilePath)
	if err != nil {
		log.Fatalf("failed to read the file at %v: %v", testFilePath, err)
	}
	category, llmCategorized := ProcessSnippet(string(contents), SHELL, llm, ctx, false)
	if category != ExampleReturnObject {
		t.Errorf("For category, got %v, want %v", category, ExampleReturnObject)
	}
	if llmCategorized != true {
		t.Errorf("For llm-categorized, got %v, want %v", llmCategorized, true)
	}
}

func TestCategorizeTaskBasedUsage(t *testing.T) {
	llm, err := ollama.New(ollama.WithModel(MODEL))
	if err != nil {
		log.Fatalf("failed to connect to ollama: %v", err)
	}
	ctx := context.Background()
	testFilePath, err := filepath.Abs("../categorize-examples/examples/manage-indexes/drop-index.go")
	if err != nil {
		log.Fatalf("failed to construct the file path: %v", err)
	}
	contents, err := os.ReadFile(testFilePath)
	if err != nil {
		log.Fatalf("failed to read the file at %v: %v", testFilePath, err)
	}
	category, llmCategorized := ProcessSnippet(string(contents), GO, llm, ctx, true)
	if category != UsageExample {
		t.Errorf("For category, got %v, want %v", category, UsageExample)
	}
	if llmCategorized != false {
		t.Errorf("For llm-categorized, got %v, want %v", llmCategorized, false)
	}
}
