package snooty

import (
	"gdcd/types"
	"net/http"
	"testing"
	"time"
)

// TODO: Figure out why this test is failing. The stub has 13 JSON blobs where "type":"page" - I'm getting one too few back. The first "page" response is always nil. This is not a problem for the C driver.
func TestSparkConnectorStubShouldReturnPages(t *testing.T) {
	testProject := types.ProjectDetails{
		ProjectName: "spark-connector",
		Version:     "",
		ProdUrl:     "",
	}
	projectDocuments := GetProjectPages(testProject, &http.Client{Timeout: 5 * time.Second})
	projectDocumentCount := len(projectDocuments)
	expectedProjectDocumentCount := 13
	if projectDocumentCount != expectedProjectDocumentCount {
		t.Errorf("FAILED: got %d project documents, want %d", projectDocumentCount, expectedProjectDocumentCount)
	}
}

func TestCDriverStubShouldReturnPages(t *testing.T) {
	testProject := types.ProjectDetails{
		ProjectName: "c",
		Version:     "",
		ProdUrl:     "",
	}
	projectDocuments := GetProjectPages(testProject, &http.Client{Timeout: 5 * time.Second})
	projectDocumentCount := len(projectDocuments)
	expectedProjectDocumentCount := 10
	if projectDocumentCount != expectedProjectDocumentCount {
		t.Errorf("FAILED: got %d project documents, want %d", projectDocumentCount, expectedProjectDocumentCount)
	}
}
