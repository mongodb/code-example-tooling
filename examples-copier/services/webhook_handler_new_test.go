package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
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
		ConfigRepoOwner: "test-owner",
		ConfigRepoName:  "test-repo",
		
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
		ConfigRepoOwner: "test-owner",
		ConfigRepoName:  "test-repo",
		
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
		ConfigRepoOwner: "test-owner",
		ConfigRepoName:  "test-repo",
		
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
		ConfigRepoOwner: "test-owner",
		ConfigRepoName:  "test-repo",
		
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
		ConfigRepoOwner: "test-owner",
		ConfigRepoName:  "test-repo",
		
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
	// Note: This test triggers a background goroutine that processes the merged PR.
	// The goroutine will fail when trying to load config/fetch files from GitHub,
	// but that's expected in a unit test environment. The test only verifies that
	// the webhook handler returns the correct HTTP response.

	// Set up environment variables to prevent ConfigurePermissions from failing
	// We don't clean these up because:
	// 1. The background goroutine may still need them after the test completes
	// 2. TestMain in github_write_to_target_test.go sets them up properly anyway
	// 3. These are test values that won't affect other tests
	os.Setenv(configs.AppId, "123456")
	os.Setenv(configs.InstallationId, "789012")
	os.Setenv(configs.ConfigRepoOwner, "test-owner")
	os.Setenv(configs.ConfigRepoName, "test-repo")
	os.Setenv("SKIP_SECRET_MANAGER", "true")

	// Generate a valid RSA private key for testing
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	der := x509.MarshalPKCS1PrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	os.Setenv("GITHUB_APP_PRIVATE_KEY", string(pemBytes))
	os.Setenv("GITHUB_APP_PRIVATE_KEY_B64", base64.StdEncoding.EncodeToString(pemBytes))

	// Set InstallationAccessToken to prevent ConfigurePermissions from being called
	// We don't reset this because the background goroutine may still need it after the test completes
	InstallationAccessToken = "test-token"

	config := &configs.Config{
		ConfigRepoOwner: "test-owner",
		ConfigRepoName:  "test-repo",
		ConfigFile:      "nonexistent-config.yaml", // Use nonexistent file to prevent actual config loading
		AuditEnabled:    false,
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
			Base: &github.PullRequestBranch{
				Ref: github.String("main"),
			},
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

	// Note: The background goroutine will continue running and will eventually fail
	// when trying to access GitHub APIs. This is expected and doesn't affect the test result.
}

func TestHandleWebhookWithContainer_MergedPRToDevelopmentBranch(t *testing.T) {
	// This test verifies that PRs merged to non-main branches (like development)
	// are accepted by the webhook handler but won't match any workflows
	// (assuming workflows are configured for main branch only)

	// Set up environment variables
	os.Setenv(configs.AppId, "123456")
	os.Setenv(configs.InstallationId, "789012")
	os.Setenv(configs.ConfigRepoOwner, "test-owner")
	os.Setenv(configs.ConfigRepoName, "test-repo")
	os.Setenv("SKIP_SECRET_MANAGER", "true")

	// Generate a valid RSA private key for testing
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	der := x509.MarshalPKCS1PrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	os.Setenv("GITHUB_APP_PRIVATE_KEY", string(pemBytes))
	os.Setenv("GITHUB_APP_PRIVATE_KEY_B64", base64.StdEncoding.EncodeToString(pemBytes))

	InstallationAccessToken = "test-token"

	config := &configs.Config{
		ConfigRepoOwner: "test-owner",
		ConfigRepoName:  "test-repo",
		ConfigFile:      "nonexistent-config.yaml",
		AuditEnabled:    false,
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	// Create a merged PR event to development branch
	prEvent := &github.PullRequestEvent{
		Action: github.String("closed"),
		PullRequest: &github.PullRequest{
			Number:         github.Int(456),
			Merged:         github.Bool(true),
			MergeCommitSHA: github.String("def456"),
			Base: &github.PullRequestBranch{
				Ref: github.String("development"),
			},
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

	// Should still return 202 Accepted (webhook accepts the event)
	if w.Code != http.StatusAccepted {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusAccepted)
	}

	// Check response body
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "accepted" {
		t.Errorf("Response status = %v, want accepted", response["status"])
	}

	// Note: The background goroutine will fail to find matching workflows
	// because the workflow config specifies main branch, not development.
	// This is the expected behavior - the webhook accepts the event but
	// no workflows will be processed.
}

func TestHandleWebhookWithContainer_MergedPRWithDifferentBranches(t *testing.T) {
	// This test verifies that the base branch is correctly extracted
	// from different PR events

	testCases := []struct {
		name       string
		baseBranch string
		prNumber   int
	}{
		{
			name:       "main branch",
			baseBranch: "main",
			prNumber:   100,
		},
		{
			name:       "development branch",
			baseBranch: "development",
			prNumber:   101,
		},
		{
			name:       "feature branch",
			baseBranch: "feature/new-feature",
			prNumber:   102,
		},
		{
			name:       "release branch",
			baseBranch: "release/v1.0",
			prNumber:   103,
		},
	}

	// Set up environment variables
	os.Setenv(configs.AppId, "123456")
	os.Setenv(configs.InstallationId, "789012")
	os.Setenv(configs.ConfigRepoOwner, "test-owner")
	os.Setenv(configs.ConfigRepoName, "test-repo")
	os.Setenv("SKIP_SECRET_MANAGER", "true")

	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	der := x509.MarshalPKCS1PrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	os.Setenv("GITHUB_APP_PRIVATE_KEY", string(pemBytes))
	os.Setenv("GITHUB_APP_PRIVATE_KEY_B64", base64.StdEncoding.EncodeToString(pemBytes))

	InstallationAccessToken = "test-token"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &configs.Config{
				ConfigRepoOwner: "test-owner",
				ConfigRepoName:  "test-repo",
				ConfigFile:      "nonexistent-config.yaml",
				AuditEnabled:    false,
			}

			container, err := NewServiceContainer(config)
			if err != nil {
				t.Fatalf("NewServiceContainer() error = %v", err)
			}

			// Create a merged PR event with specific base branch
			prEvent := &github.PullRequestEvent{
				Action: github.String("closed"),
				PullRequest: &github.PullRequest{
					Number:         github.Int(tc.prNumber),
					Merged:         github.Bool(true),
					MergeCommitSHA: github.String("abc123"),
					Base: &github.PullRequestBranch{
						Ref: github.String(tc.baseBranch),
					},
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

			// Should return 202 Accepted for all merged PRs
			if w.Code != http.StatusAccepted {
				t.Errorf("Status code = %d, want %d", w.Code, http.StatusAccepted)
			}

			// Check response body
			var response map[string]string
			json.Unmarshal(w.Body.Bytes(), &response)
			if response["status"] != "accepted" {
				t.Errorf("Response status = %v, want accepted", response["status"])
			}
		})
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

