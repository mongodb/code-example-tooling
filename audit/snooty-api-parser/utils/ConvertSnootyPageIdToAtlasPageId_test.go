package utils

import "testing"

func TestConversionFuncCorrectlyCreatesAtlasPageIdFromSingleElementPageID(t *testing.T) {
	// This is the `page_id` field that comes in from the Snooty Data API documents response.
	snootyPageId := "c/docsworker-xlarge/v1.30/aggregation"
	atlasPageId := ConvertSnootyPageIdToAtlasPageId(snootyPageId)
	expectedPageId := "aggregation"
	if atlasPageId != expectedPageId {
		t.Errorf("FAILED: got %s for the Atlas Page ID, want %s", atlasPageId, expectedPageId)
	}
}

func TestConversionFuncCorrectlyCreatesAtlasPageIdFromMultiElementPageID(t *testing.T) {
	// This is the `page_id` field that comes in from the Snooty Data API documents response.
	snootyPageId := "c/docsworker-xlarge/v1.30/connect/connection-targets"
	atlasPageId := ConvertSnootyPageIdToAtlasPageId(snootyPageId)
	expectedPageId := "connect|connection-targets"
	if atlasPageId != expectedPageId {
		t.Errorf("FAILED: got %s for the Atlas Page ID, want %s", atlasPageId, expectedPageId)
	}
}
