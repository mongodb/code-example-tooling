package snooty

import (
	"gdcd/types"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestStubbedProjectsReturnTheCorrectNumberOfProjects(t *testing.T) {
	projectDocuments := GetProjects(&http.Client{Timeout: 5 * time.Second})
	projectDocumentCount := len(projectDocuments)
	expectedProjectDocumentCount := 1
	if projectDocumentCount != expectedProjectDocumentCount {
		t.Errorf("FAILED: got %d project documents, want %d", projectDocumentCount, expectedProjectDocumentCount)
	}
}

func TestStubbedProjectsReturnCorrectProjectDetails(t *testing.T) {
	projectDocuments := GetProjects(&http.Client{Timeout: 5 * time.Second})
	expectedProjectDocument := types.DocsProjectDetails{
		ProjectName:  "spark-connector",
		ActiveBranch: "v10.4",
		ProdUrl:      "https://mongodb.com/docs/spark-connector/current",
	}
	if !reflect.DeepEqual(projectDocuments[0], expectedProjectDocument) {
		t.Errorf("FAILED: got %v, want %v", projectDocuments, expectedProjectDocument)
	}
}
