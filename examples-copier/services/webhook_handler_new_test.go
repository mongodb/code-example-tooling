package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v48/github"
	"github.com/mongodb/code-example-tooling/code-copier/configs"
)

func TestSimpleVerifySignature(t *testing.T) {
	secret := []byte("test-secret")
	body := []byte(`{"test": "payload"}`)

	// Generate valid signature
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	validSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		sigHeader string
		body      []byte
		secret    []byte
		want      bool
	}{
		{
			name:      "valid signature",
			sigHeader: validSignature,
			body:      body,
			secret:    secret,
			want:      true,
		},
		{
			name:      "invalid signature",
			sigHeader: "sha256=invalid",
			body:      body,
			secret:    secret,
			want:      false,
		},
		{
			name:      "missing sha256 prefix",
			sigHeader: "invalid",
			body:      body,
			secret:    secret,
			want:      false,
		},
		{
			name:      "empty signature",
			sigHeader: "",
			body:      body,
			secret:    secret,
			want:      false,
		},
		{
			name:      "wrong secret",
			sigHeader: validSignature,
			body:      body,
			secret:    []byte("wrong-secret"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := simpleVerifySignature(tt.sigHeader, tt.body, tt.secret)
			if got != tt.want {
				t.Errorf("simpleVerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleWebhookWithContainer_MissingEventType(t *testing.T) {
	config := &configs.Config{
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		
		AuditEnabled:   false,
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	payload := []byte(`{"action": "closed"}`)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	// Missing X-GitHub-Event header

	w := httptest.NewRecorder()

	HandleWebhookWithContainer(w, req, config, container)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("missing event type")) {
		t.Error("Expected 'missing event type' in response body")
	}
}

func TestHandleWebhookWithContainer_InvalidSignature(t *testing.T) {
	config := &configs.Config{
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		
		WebhookSecret:  "test-secret",
		AuditEnabled:   false,
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	payload := []byte(`{"action": "closed"}`)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")

	w := httptest.NewRecorder()

	HandleWebhookWithContainer(w, req, config, container)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleWebhookWithContainer_ValidSignature(t *testing.T) {
	secret := "test-secret"
	config := &configs.Config{
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		
		WebhookSecret:  secret,
		AuditEnabled:   false,
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	// Create a valid pull_request event payload
	prEvent := &github.PullRequestEvent{
		Action: github.String("opened"),
		PullRequest: &github.PullRequest{
			Number: github.Int(123),
			Merged: github.Bool(false),
		},
	}

	payload, _ := json.Marshal(prEvent)

	// Generate valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", signature)

	w := httptest.NewRecorder()

	HandleWebhookWithContainer(w, req, config, container)

	// Should not return unauthorized
	if w.Code == http.StatusUnauthorized {
		t.Error("Valid signature was rejected")
	}
}

func TestHandleWebhookWithContainer_NonPREvent(t *testing.T) {
	config := &configs.Config{
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		
		AuditEnabled:   false,
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	// Create a push event (not a PR event)
	pushEvent := map[string]interface{}{
		"ref": "refs/heads/main",
	}
	payload, _ := json.Marshal(pushEvent)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "push")

	w := httptest.NewRecorder()

	HandleWebhookWithContainer(w, req, config, container)

	// Should return 204 No Content for non-PR events
	if w.Code != http.StatusNoContent {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleWebhookWithContainer_NonMergedPR(t *testing.T) {
	config := &configs.Config{
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		
		AuditEnabled:   false,
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	// Create a PR event that's not merged
	prEvent := &github.PullRequestEvent{
		Action: github.String("opened"),
		PullRequest: &github.PullRequest{
			Number: github.Int(123),
			Merged: github.Bool(false),
		},
	}
	payload, _ := json.Marshal(prEvent)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")

	w := httptest.NewRecorder()

	HandleWebhookWithContainer(w, req, config, container)

	// Should return 204 No Content for non-merged PRs
	if w.Code != http.StatusNoContent {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleWebhookWithContainer_MergedPR(t *testing.T) {
	config := &configs.Config{
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		
		AuditEnabled:   false,
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	// Create a merged PR event
	prEvent := &github.PullRequestEvent{
		Action: github.String("closed"),
		PullRequest: &github.PullRequest{
			Number:         github.Int(123),
			Merged:         github.Bool(true),
			MergeCommitSHA: github.String("abc123"),
		},
		Repo: &github.Repository{
			Name: github.String("test-repo"),
			Owner: &github.User{
				Login: github.String("test-owner"),
			},
		},
	}
	payload, _ := json.Marshal(prEvent)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")

	w := httptest.NewRecorder()

	HandleWebhookWithContainer(w, req, config, container)

	// Should return 202 Accepted for merged PRs
	if w.Code != http.StatusAccepted {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusAccepted)
	}

	// Check response body
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "accepted" {
		t.Errorf("Response status = %v, want accepted", response["status"])
	}
}

func TestRetrieveFileContentsWithConfigAndBranch(t *testing.T) {
	// This test would require mocking the GitHub client
	// For now, we document the expected behavior
	t.Skip("Skipping test that requires GitHub API mocking")

	// Expected behavior:
	// - Should call client.Repositories.GetContents with correct parameters
	// - Should use the specified branch in RepositoryContentGetOptions
	// - Should return file content on success
	// - Should return error on failure
}

func TestMaxWebhookBodyBytes(t *testing.T) {
	// Verify the constant is set correctly
	expected := 1 << 20 // 1MB
	if maxWebhookBodyBytes != expected {
		t.Errorf("maxWebhookBodyBytes = %d, want %d", maxWebhookBodyBytes, expected)
	}
}

func TestStatusDeleted(t *testing.T) {
	// Verify the constant is set correctly
	if statusDeleted != "DELETED" {
		t.Errorf("statusDeleted = %s, want DELETED", statusDeleted)
	}
}

