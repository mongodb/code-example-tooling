package snooty

import (
	"testing"
)

func TestTimeStampShouldReturnNothing(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("timestamp.json")
	maybePage := GetPageFromResponse(inputJSON, GitHubUsernameNetlify)
	if maybePage != nil {
		t.Errorf("FAILED: got something, want nothing")
	}
}

func TestMetadataShouldReturnNothing(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("metadata.json")
	maybePage := GetPageFromResponse(inputJSON, GitHubUsernameNetlify)
	if maybePage != nil {
		t.Errorf("FAILED: got something, want nothing")
	}
}

func TestAssetShouldReturnNothing(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("asset.json")
	maybePage := GetPageFromResponse(inputJSON, GitHubUsernameNetlify)
	if maybePage != nil {
		t.Errorf("FAILED: got something, want nothing")
	}
}

func TestPageWithNetlifyShouldReturnPage(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("page-with-code-nodes.json")
	maybePage := GetPageFromResponse(inputJSON, GitHubUsernameNetlify)
	if maybePage == nil {
		t.Errorf("FAILED: got nothing, should have a page")
	}
}

func TestPageWithDocsBuilderBotShouldReturnPage(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("page-with-github-username-docs-builder-bot.json")
	maybePage := GetPageFromResponse(inputJSON, GitHubUsernameDocsBuilderBot)
	if maybePage == nil {
		t.Errorf("FAILED: got nothing, should have a page")
	}
}

func TestPageWithDocsBuilderBotAndNoGitHubUsernameShouldReturnPage(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("page-with-no-github-username.json")
	maybePage := GetPageFromResponse(inputJSON, GitHubUsernameDocsBuilderBot)
	if maybePage == nil {
		t.Errorf("FAILED: got nothing, should have a page")
	}
}

func TestPageWithWrongUsernameShouldReturnNothing(t *testing.T) {
	inputJSON := LoadJsonTestDataFromFile("page-with-code-nodes.json")
	maybePage := GetPageFromResponse(inputJSON, GitHubUsernameDocsBuilderBot)
	if maybePage != nil {
		t.Errorf("FAILED: got a page, should have nothing")
	}
}
