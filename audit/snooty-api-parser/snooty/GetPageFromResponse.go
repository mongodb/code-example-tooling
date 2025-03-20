package snooty

import (
	"encoding/json"
	"log"
	"snooty-api-parser/types"
)

// GetPageFromResponse checks the "type" of the newline-delimited JSON blob, and if it is a "page",
// deserializes it to a page object and returns it. If the JSON blob is a timestamp, metadata, or asset, we ignore it.
func GetPageFromResponse(line []byte) *types.PageWrapper {
	var generic map[string]interface{}
	if err := json.Unmarshal(line, &generic); err != nil {
		log.Fatalf("Failed to unmarshal line: %v", err)
	}
	typeField, ok := generic["type"].(string)
	if !ok {
		log.Fatalf("Type field is missing or not a string in line: %s", line)
	}

	// Process based on typeField
	switch typeField {
	case "timestamp":
		var timestamp types.TimestampData
		if err := json.Unmarshal(line, &timestamp); err != nil {
			log.Fatalf("Failed to unmarshal TimestampData: %v", err)
		}
	case "metadata":
		var metadata types.ProjectMetadataWrapper
		if err := json.Unmarshal(line, &metadata); err != nil {
			log.Fatalf("Failed to unmarshal ProjectMetadata: %v", err)
		}
	case "page":
		var page types.PageWrapper
		if err := json.Unmarshal(line, &page); err != nil {
			log.Fatalf("Failed to unmarshal PageMetadata: %v", err)
		}
		return &page
		//// Because of the DOP bug duplicating pages with different GitHub usernames, we can pick which username to return.
		//// If the page doesn't match the GitHub username we want, we just don't return it.
		//if gitHubUserName == GitHubUsernameNetlify && page.Data.GitHubUsername == gitHubUserName {
		//	// If the name we're passing in when we call the function is netlify, and we want netlify, return the page
		//	return &page
		//} else if gitHubUserName == GitHubUsernameDocsBuilderBot {
		//	// If the name we're passing in when we call the function is docs-builder-bot, the response may either contain
		//	// the GiHubUsername "docs-builder-bot" or may omit this field entirely. In either of these cases, if we want
		//	// "docs-builder-bot", return the page.
		//	if page.Data.GitHubUsername == gitHubUserName || page.Data.GitHubUsername == "" {
		//		return &page
		//	}
		//} else {
		//	log.Printf("ISSUE: Tried to get a page for username %s but returned nothing", gitHubUserName)
		//	return nil
		//}
	case "asset":
		var fileAsset types.ProjectAsset
		if err := json.Unmarshal(line, &fileAsset); err != nil {
			log.Fatalf("Failed to unmarshal ProjectAsset: %v", err)
		}
	default:
		log.Printf("Unknown type: %s\n", typeField)
	}
	return nil
}
