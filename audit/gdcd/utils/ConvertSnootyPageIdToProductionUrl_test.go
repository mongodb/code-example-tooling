package utils

import "testing"

func TestConversionFuncCorrectlyCreatesProductionUrlFromSingleElementPageID(t *testing.T) {
	// This is the `page_id` field that comes in from the Snooty Data API documents response.
	pageId := "c/docsworker-xlarge/v1.30/aggregation"
	// This is the `fullUrl` field that comes in from the Snooty Data API Projects response. It's set as the
	// types.DocsProjectDetails `ProdUrl` field, and pulled from there to pass to this function.
	siteUrl := "https://mongodb.com/docs/languages/c/c-driver/current"
	productionUrl := ConvertSnootyPageIdToProductionUrl(pageId, siteUrl)
	expectedProductionUrl := "https://mongodb.com/docs/languages/c/c-driver/current/aggregation"
	if productionUrl != expectedProductionUrl {
		t.Errorf("FAILED: got %s for the production URL, want %s", productionUrl, expectedProductionUrl)
	}
}

func TestConversionFuncCorrectlyCreatesProductionUrlFromMultiElementPageID(t *testing.T) {
	// This is the `page_id` field that comes in from the Snooty Data API documents response.
	pageId := "c/docsworker-xlarge/v1.30/connect/connection-targets"
	// This is the `fullUrl` field that comes in from the Snooty Data API Projects response. It's set as the
	// types.DocsProjectDetails `ProdUrl` field, and pulled from there to pass to this function.
	siteUrl := "https://mongodb.com/docs/languages/c/c-driver/current"
	productionUrl := ConvertSnootyPageIdToProductionUrl(pageId, siteUrl)
	expectedProductionUrl := "https://mongodb.com/docs/languages/c/c-driver/current/connect/connection-targets"
	if productionUrl != expectedProductionUrl {
		t.Errorf("FAILED: got %s for the production URL, want %s", productionUrl, expectedProductionUrl)
	}
}
