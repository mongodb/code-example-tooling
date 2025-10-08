package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v48/github"
)

func main() {
	// Command-line flags
	prNumber := flag.Int("pr", 0, "PR number to fetch from GitHub")
	owner := flag.String("owner", "", "Repository owner")
	repo := flag.String("repo", "", "Repository name")
	webhookURL := flag.String("url", "http://localhost:8080/webhook", "Webhook URL")
	secret := flag.String("secret", "", "Webhook secret for signature")
	payloadFile := flag.String("payload", "", "Path to custom payload JSON file")
	dryRun := flag.Bool("dry-run", false, "Print payload without sending")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	var payload []byte
	var err error

	// Option 1: Use custom payload file
	if *payloadFile != "" {
		payload, err = os.ReadFile(*payloadFile)
		if err != nil {
			fmt.Printf("Error reading payload file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Loaded payload from %s\n", *payloadFile)
	} else if *prNumber > 0 {
		// Option 2: Fetch PR data from GitHub
		if *owner == "" || *repo == "" {
			fmt.Println("Error: -owner and -repo are required when using -pr")
			os.Exit(1)
		}

		payload, err = fetchPRPayload(*owner, *repo, *prNumber)
		if err != nil {
			fmt.Printf("Error fetching PR data: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Fetched PR #%d from %s/%s\n", *prNumber, *owner, *repo)
	} else {
		// Option 3: Use example payload
		payload = createExamplePayload()
		fmt.Println("✓ Using example payload")
	}

	// Pretty print payload if dry-run
	if *dryRun {
		fmt.Println("\n=== Payload ===")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, payload, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(payload))
		}
		fmt.Println("\n=== Dry-run mode: Not sending webhook ===")
		return
	}

	// Send webhook
	if err := sendWebhook(*webhookURL, payload, *secret); err != nil {
		fmt.Printf("Error sending webhook: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Webhook sent successfully to %s\n", *webhookURL)
}

func printHelp() {
	fmt.Println(`Test Webhook Tool

Usage:
  test-webhook [options]

Options:
  -pr int         PR number to fetch from GitHub
  -owner string   Repository owner (required with -pr)
  -repo string    Repository name (required with -pr)
  -url string     Webhook URL (default: http://localhost:8080/webhook)
  -secret string  Webhook secret for HMAC signature
  -payload string Path to custom payload JSON file
  -dry-run        Print payload without sending
  -help           Show this help

Examples:

  # Use example payload
  test-webhook

  # Fetch real PR data
  test-webhook -pr 123 -owner myorg -repo myrepo

  # Use custom payload file
  test-webhook -payload webhook-payload.json

  # Dry-run to see payload
  test-webhook -pr 123 -owner myorg -repo myrepo -dry-run

  # Send to production with secret
  test-webhook -pr 123 -owner myorg -repo myrepo \
    -url https://myapp.appspot.com/webhook \
    -secret "my-webhook-secret"

Environment Variables:
  GITHUB_TOKEN    GitHub personal access token (for fetching PR data)
  WEBHOOK_SECRET  Default webhook secret (can be overridden with -secret)
`)
}

func fetchPRPayload(owner, repo string, prNumber int) ([]byte, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Fetch PR data
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var pr github.PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	// Fetch files changed in PR
	filesURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files", owner, repo, prNumber)
	filesReq, err := http.NewRequest("GET", filesURL, nil)
	if err != nil {
		return nil, err
	}

	filesReq.Header.Set("Authorization", "Bearer "+token)
	filesReq.Header.Set("Accept", "application/vnd.github.v3+json")

	filesResp, err := client.Do(filesReq)
	if err != nil {
		return nil, err
	}
	defer filesResp.Body.Close()

	var files []map[string]interface{}
	if err := json.NewDecoder(filesResp.Body).Decode(&files); err != nil {
		return nil, err
	}

	// Create webhook payload
	payload := map[string]interface{}{
		"action": "closed",
		"number": prNumber,
		"pull_request": map[string]interface{}{
			"number":       pr.GetNumber(),
			"state":        pr.GetState(),
			"merged":       pr.GetMerged(),
			"merge_commit_sha": pr.GetMergeCommitSHA(),
			"head": map[string]interface{}{
				"ref": pr.GetHead().GetRef(),
				"sha": pr.GetHead().GetSHA(),
				"repo": map[string]interface{}{
					"name":      pr.GetHead().GetRepo().GetName(),
					"full_name": pr.GetHead().GetRepo().GetFullName(),
				},
			},
			"base": map[string]interface{}{
				"ref": pr.GetBase().GetRef(),
				"repo": map[string]interface{}{
					"name":      pr.GetBase().GetRepo().GetName(),
					"full_name": pr.GetBase().GetRepo().GetFullName(),
				},
			},
		},
		"repository": map[string]interface{}{
			"name":      repo,
			"full_name": fmt.Sprintf("%s/%s", owner, repo),
			"owner": map[string]interface{}{
				"login": owner,
			},
		},
		"files": files,
	}

	return json.Marshal(payload)
}

func createExamplePayload() []byte {
	payload := map[string]interface{}{
		"action": "closed",
		"number": 42,
		"pull_request": map[string]interface{}{
			"number": 42,
			"state":  "closed",
			"merged": true,
			"merge_commit_sha": "abc123def456",
			"head": map[string]interface{}{
				"ref": "feature-branch",
				"sha": "abc123",
				"repo": map[string]interface{}{
					"name":      "source-repo",
					"full_name": "myorg/source-repo",
				},
			},
			"base": map[string]interface{}{
				"ref": "main",
				"repo": map[string]interface{}{
					"name":      "source-repo",
					"full_name": "myorg/source-repo",
				},
			},
		},
		"repository": map[string]interface{}{
			"name":      "source-repo",
			"full_name": "myorg/source-repo",
			"owner": map[string]interface{}{
				"login": "myorg",
			},
		},
	}

	data, _ := json.Marshal(payload)
	return data
}

func sendWebhook(url string, payload []byte, secret string) error {
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")

	// Add signature if secret provided
	if secret != "" {
		signature := generateSignature(payload, secret)
		req.Header.Set("X-Hub-Signature-256", signature)
		fmt.Printf("✓ Added signature: %s\n", signature[:20]+"...")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✓ Response: %d %s\n", resp.StatusCode, resp.Status)
	if len(body) > 0 {
		fmt.Printf("✓ Response body: %s\n", string(body))
	}

	return nil
}

func generateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

