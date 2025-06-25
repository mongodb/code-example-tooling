package snooty

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

// TODO: This is failing presumably for the same reason as the test in GetProjectDocuments_test.go. Figure out why.
func TestDocsBuilderBotStubShouldReturnPages(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("spark-connector-project-documents-stub.json")
	body := io.NopCloser(strings.NewReader(string(inputJSON)))
	reader := *bufio.NewReader(body)
	pages := ReadPagesForGitHubUser(reader)
	pageCount := len(pages)
	expectedPageCount := 13
	if pageCount != expectedPageCount {
		t.Errorf("FAILED: got %d, want %d", pageCount, expectedPageCount)
	}
}

func TestNetlifyStubShouldReturnPages(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("c-driver-project-documents-stub.json")
	body := io.NopCloser(strings.NewReader(string(inputJSON)))
	reader := *bufio.NewReader(body)
	pages := ReadPagesForGitHubUser(reader)
	pageCount := len(pages)
	expectedPageCount := 10
	if pageCount != expectedPageCount {
		t.Errorf("FAILED: got %d, want %d", pageCount, expectedPageCount)
	}
}

func TestWrongUsernameShouldReturnNoPages(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("c-driver-project-documents-stub.json")
	body := io.NopCloser(strings.NewReader(string(inputJSON)))
	reader := *bufio.NewReader(body)
	pages := ReadPagesForGitHubUser(reader)
	pageCount := len(pages)
	expectedPageCount := 0
	if pageCount != expectedPageCount {
		t.Errorf("FAILED: got %d, want %d", pageCount, expectedPageCount)
	}
}
