package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSlackNotifier(t *testing.T) {
	tests := []struct {
		name       string
		webhookURL string
		channel    string
		username   string
		iconEmoji  string
		wantEnabled bool
	}{
		{
			name:       "enabled with webhook URL",
			webhookURL: "https://hooks.slack.com/services/TEST",
			channel:    "#test",
			username:   "Test Bot",
			iconEmoji:  ":robot:",
			wantEnabled: true,
		},
		{
			name:       "disabled without webhook URL",
			webhookURL: "",
			channel:    "#test",
			username:   "Test Bot",
			iconEmoji:  ":robot:",
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifier := NewSlackNotifier(tt.webhookURL, tt.channel, tt.username, tt.iconEmoji)
			if notifier.IsEnabled() != tt.wantEnabled {
				t.Errorf("IsEnabled() = %v, want %v", notifier.IsEnabled(), tt.wantEnabled)
			}
		})
	}
}

func TestSlackNotifier_NotifyPRProcessed(t *testing.T) {
	tests := []struct {
		name        string
		event       *PRProcessedEvent
		wantColor   string
		wantEnabled bool
	}{
		{
			name: "successful PR with no failures",
			event: &PRProcessedEvent{
				PRNumber:       123,
				PRTitle:        "Add new feature",
				PRURL:          "https://github.com/test/repo/pull/123",
				SourceRepo:     "test/repo",
				FilesMatched:   5,
				FilesCopied:    5,
				FilesFailed:    0,
				ProcessingTime: 2 * time.Second,
			},
			wantColor:   "good",
			wantEnabled: true,
		},
		{
			name: "PR with some failures",
			event: &PRProcessedEvent{
				PRNumber:       124,
				PRTitle:        "Fix bug",
				PRURL:          "https://github.com/test/repo/pull/124",
				SourceRepo:     "test/repo",
				FilesMatched:   5,
				FilesCopied:    3,
				FilesFailed:    2,
				ProcessingTime: 3 * time.Second,
			},
			wantColor:   "warning",
			wantEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			var receivedMessage *SlackMessage
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &receivedMessage)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			notifier := NewSlackNotifier(server.URL, "#test", "Test Bot", ":robot:")
			ctx := context.Background()

			err := notifier.NotifyPRProcessed(ctx, tt.event)
			if err != nil {
				t.Errorf("NotifyPRProcessed() error = %v", err)
			}

			if receivedMessage == nil {
				t.Fatal("No message received")
			}

			if len(receivedMessage.Attachments) == 0 {
				t.Fatal("No attachments in message")
			}

			attachment := receivedMessage.Attachments[0]
			if attachment.Color != tt.wantColor {
				t.Errorf("Color = %v, want %v", attachment.Color, tt.wantColor)
			}

			expectedTitle := fmt.Sprintf("✅ PR #%d Processed", tt.event.PRNumber)
			if attachment.Title != expectedTitle {
				t.Errorf("Title = %v, want %v", attachment.Title, expectedTitle)
			}
		})
	}
}

func TestSlackNotifier_NotifyError(t *testing.T) {
	event := &ErrorEvent{
		Operation:  "file_copy",
		Error:      fmt.Errorf("test error"),
		PRNumber:   125,
		SourceRepo: "test/repo",
	}

	var receivedMessage *SlackMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedMessage)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewSlackNotifier(server.URL, "#test", "Test Bot", ":robot:")
	ctx := context.Background()

	err := notifier.NotifyError(ctx, event)
	if err != nil {
		t.Errorf("NotifyError() error = %v", err)
	}

	if receivedMessage == nil {
		t.Fatal("No message received")
	}

	if len(receivedMessage.Attachments) == 0 {
		t.Fatal("No attachments in message")
	}

	attachment := receivedMessage.Attachments[0]
	if attachment.Color != "danger" {
		t.Errorf("Color = %v, want danger", attachment.Color)
	}

	if attachment.Title != "❌ Error Occurred" {
		t.Errorf("Title = %v, want ❌ Error Occurred", attachment.Title)
	}
}

func TestSlackNotifier_NotifyFilesCopied(t *testing.T) {
	tests := []struct {
		name      string
		fileCount int
		wantTruncated bool
	}{
		{
			name:      "few files",
			fileCount: 5,
			wantTruncated: false,
		},
		{
			name:      "many files",
			fileCount: 15,
			wantTruncated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := make([]string, tt.fileCount)
			for i := 0; i < tt.fileCount; i++ {
				files[i] = fmt.Sprintf("file%d.go", i)
			}

			event := &FilesCopiedEvent{
				PRNumber:   126,
				SourceRepo: "test/source",
				TargetRepo: "test/target",
				FileCount:  tt.fileCount,
				Files:      files,
				RuleName:   "test-rule",
			}

			var receivedMessage *SlackMessage
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &receivedMessage)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			notifier := NewSlackNotifier(server.URL, "#test", "Test Bot", ":robot:")
			ctx := context.Background()

			err := notifier.NotifyFilesCopied(ctx, event)
			if err != nil {
				t.Errorf("NotifyFilesCopied() error = %v", err)
			}

			if receivedMessage == nil {
				t.Fatal("No message received")
			}

			attachment := receivedMessage.Attachments[0]
			if tt.wantTruncated {
				// Should contain "... and X more"
				if !contains(attachment.Text, "and") || !contains(attachment.Text, "more") {
					t.Error("Expected truncation message not found")
				}
			}
		})
	}
}

func TestSlackNotifier_NotifyDeprecation(t *testing.T) {
	event := &DeprecationEvent{
		PRNumber:   127,
		SourceRepo: "test/repo",
		FileCount:  3,
		Files:      []string{"old1.go", "old2.go", "old3.go"},
	}

	var receivedMessage *SlackMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedMessage)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewSlackNotifier(server.URL, "#test", "Test Bot", ":robot:")
	ctx := context.Background()

	err := notifier.NotifyDeprecation(ctx, event)
	if err != nil {
		t.Errorf("NotifyDeprecation() error = %v", err)
	}

	if receivedMessage == nil {
		t.Fatal("No message received")
	}

	attachment := receivedMessage.Attachments[0]
	if attachment.Color != "warning" {
		t.Errorf("Color = %v, want warning", attachment.Color)
	}

	expectedTitle := fmt.Sprintf("⚠️ Files Deprecated from PR #%d", event.PRNumber)
	if attachment.Title != expectedTitle {
		t.Errorf("Title = %v, want %v", attachment.Title, expectedTitle)
	}
}

func TestSlackNotifier_DisabledNotifier(t *testing.T) {
	// Create notifier without webhook URL (disabled)
	notifier := NewSlackNotifier("", "#test", "Test Bot", ":robot:")
	ctx := context.Background()

	// All notification methods should return nil without error
	err := notifier.NotifyPRProcessed(ctx, &PRProcessedEvent{})
	if err != nil {
		t.Errorf("NotifyPRProcessed() error = %v, want nil", err)
	}

	err = notifier.NotifyError(ctx, &ErrorEvent{})
	if err != nil {
		t.Errorf("NotifyError() error = %v, want nil", err)
	}

	err = notifier.NotifyFilesCopied(ctx, &FilesCopiedEvent{})
	if err != nil {
		t.Errorf("NotifyFilesCopied() error = %v, want nil", err)
	}

	err = notifier.NotifyDeprecation(ctx, &DeprecationEvent{})
	if err != nil {
		t.Errorf("NotifyDeprecation() error = %v, want nil", err)
	}
}

func TestFormatFileList(t *testing.T) {
	files := []string{"file1.go", "file2.go", "file3.go"}
	result := formatFileList(files)

	for _, file := range files {
		if !contains(result, file) {
			t.Errorf("formatFileList() missing file %s", file)
		}
	}

	// Should have bullet points
	if !contains(result, "•") {
		t.Error("formatFileList() missing bullet points")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

