package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackNotifier handles sending notifications to Slack
type SlackNotifier interface {
	// NotifyPRProcessed sends a notification when a PR is successfully processed
	NotifyPRProcessed(ctx context.Context, event *PRProcessedEvent) error
	
	// NotifyError sends a notification when an error occurs
	NotifyError(ctx context.Context, event *ErrorEvent) error
	
	// NotifyFilesCopied sends a notification when files are copied
	NotifyFilesCopied(ctx context.Context, event *FilesCopiedEvent) error
	
	// NotifyDeprecation sends a notification when files are deprecated
	NotifyDeprecation(ctx context.Context, event *DeprecationEvent) error
	
	// IsEnabled returns true if Slack notifications are enabled
	IsEnabled() bool
}

// PRProcessedEvent contains information about a processed PR
type PRProcessedEvent struct {
	PRNumber      int
	PRTitle       string
	PRURL         string
	SourceRepo    string
	FilesMatched  int
	FilesCopied   int
	FilesFailed   int
	ProcessingTime time.Duration
}

// ErrorEvent contains information about an error
type ErrorEvent struct {
	Operation   string
	Error       error
	PRNumber    int
	SourceRepo  string
	AdditionalInfo map[string]interface{}
}

// FilesCopiedEvent contains information about copied files
type FilesCopiedEvent struct {
	PRNumber    int
	SourceRepo  string
	TargetRepo  string
	FileCount   int
	Files       []string
	RuleName    string
}

// DeprecationEvent contains information about deprecated files
type DeprecationEvent struct {
	PRNumber    int
	SourceRepo  string
	FileCount   int
	Files       []string
}

// DefaultSlackNotifier implements SlackNotifier using Slack webhooks
type DefaultSlackNotifier struct {
	webhookURL string
	enabled    bool
	channel    string
	username   string
	iconEmoji  string
	httpClient *http.Client
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(webhookURL, channel, username, iconEmoji string) SlackNotifier {
	enabled := webhookURL != ""
	
	return &DefaultSlackNotifier{
		webhookURL: webhookURL,
		enabled:    enabled,
		channel:    channel,
		username:   username,
		iconEmoji:  iconEmoji,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsEnabled returns true if Slack notifications are enabled
func (sn *DefaultSlackNotifier) IsEnabled() bool {
	return sn.enabled
}

// NotifyPRProcessed sends a notification when a PR is successfully processed
func (sn *DefaultSlackNotifier) NotifyPRProcessed(ctx context.Context, event *PRProcessedEvent) error {
	if !sn.enabled {
		return nil
	}
	
	color := "good" // green
	if event.FilesFailed > 0 {
		color = "warning" // yellow
	}
	
	message := &SlackMessage{
		Channel:   sn.channel,
		Username:  sn.username,
		IconEmoji: sn.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      color,
				Title:      fmt.Sprintf("✅ PR #%d Processed", event.PRNumber),
				TitleLink:  event.PRURL,
				Text:       event.PRTitle,
				Fields: []SlackField{
					{Title: "Repository", Value: event.SourceRepo, Short: true},
					{Title: "Files Matched", Value: fmt.Sprintf("%d", event.FilesMatched), Short: true},
					{Title: "Files Copied", Value: fmt.Sprintf("%d", event.FilesCopied), Short: true},
					{Title: "Files Failed", Value: fmt.Sprintf("%d", event.FilesFailed), Short: true},
					{Title: "Processing Time", Value: event.ProcessingTime.String(), Short: true},
				},
				Footer:     "Examples Copier",
				FooterIcon: "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
	
	return sn.sendMessage(ctx, message)
}

// NotifyError sends a notification when an error occurs
func (sn *DefaultSlackNotifier) NotifyError(ctx context.Context, event *ErrorEvent) error {
	if !sn.enabled {
		return nil
	}
	
	fields := []SlackField{
		{Title: "Operation", Value: event.Operation, Short: true},
		{Title: "Error", Value: event.Error.Error(), Short: false},
	}
	
	if event.SourceRepo != "" {
		fields = append(fields, SlackField{Title: "Repository", Value: event.SourceRepo, Short: true})
	}
	
	if event.PRNumber > 0 {
		fields = append(fields, SlackField{Title: "PR Number", Value: fmt.Sprintf("#%d", event.PRNumber), Short: true})
	}
	
	message := &SlackMessage{
		Channel:   sn.channel,
		Username:  sn.username,
		IconEmoji: sn.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      "danger", // red
				Title:      "❌ Error Occurred",
				Text:       fmt.Sprintf("An error occurred during %s", event.Operation),
				Fields:     fields,
				Footer:     "Examples Copier",
				FooterIcon: "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
	
	return sn.sendMessage(ctx, message)
}

// NotifyFilesCopied sends a notification when files are copied
func (sn *DefaultSlackNotifier) NotifyFilesCopied(ctx context.Context, event *FilesCopiedEvent) error {
	if !sn.enabled {
		return nil
	}
	
	// Limit files shown to first 10
	filesText := ""
	displayFiles := event.Files
	if len(displayFiles) > 10 {
		displayFiles = displayFiles[:10]
		filesText = fmt.Sprintf("```\n%s\n... and %d more```", 
			formatFileList(displayFiles), 
			len(event.Files)-10)
	} else {
		filesText = fmt.Sprintf("```\n%s```", formatFileList(displayFiles))
	}
	
	message := &SlackMessage{
		Channel:   sn.channel,
		Username:  sn.username,
		IconEmoji: sn.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      "good", // green
				Title:      fmt.Sprintf("📋 Files Copied from PR #%d", event.PRNumber),
				Text:       filesText,
				Fields: []SlackField{
					{Title: "Source", Value: event.SourceRepo, Short: true},
					{Title: "Target", Value: event.TargetRepo, Short: true},
					{Title: "Rule", Value: event.RuleName, Short: true},
					{Title: "File Count", Value: fmt.Sprintf("%d", event.FileCount), Short: true},
				},
				Footer:     "Examples Copier",
				FooterIcon: "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
	
	return sn.sendMessage(ctx, message)
}

// NotifyDeprecation sends a notification when files are deprecated
func (sn *DefaultSlackNotifier) NotifyDeprecation(ctx context.Context, event *DeprecationEvent) error {
	if !sn.enabled {
		return nil
	}
	
	filesText := fmt.Sprintf("```\n%s```", formatFileList(event.Files))
	
	message := &SlackMessage{
		Channel:   sn.channel,
		Username:  sn.username,
		IconEmoji: sn.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      "warning", // yellow
				Title:      fmt.Sprintf("⚠️ Files Deprecated from PR #%d", event.PRNumber),
				Text:       filesText,
				Fields: []SlackField{
					{Title: "Repository", Value: event.SourceRepo, Short: true},
					{Title: "File Count", Value: fmt.Sprintf("%d", event.FileCount), Short: true},
				},
				Footer:     "Examples Copier",
				FooterIcon: "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
	
	return sn.sendMessage(ctx, message)
}

// sendMessage sends a message to Slack
func (sn *DefaultSlackNotifier) sendMessage(ctx context.Context, message *SlackMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", sn.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := sn.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned non-200 status: %d", resp.StatusCode)
	}
	
	return nil
}

// formatFileList formats a list of files for display
func formatFileList(files []string) string {
	result := ""
	for _, file := range files {
		result += "• " + file + "\n"
	}
	return result
}

// SlackMessage represents a Slack message
type SlackMessage struct {
	Channel     string             `json:"channel,omitempty"`
	Username    string             `json:"username,omitempty"`
	IconEmoji   string             `json:"icon_emoji,omitempty"`
	Text        string             `json:"text,omitempty"`
	Attachments []SlackAttachment  `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color      string       `json:"color,omitempty"`
	Title      string       `json:"title,omitempty"`
	TitleLink  string       `json:"title_link,omitempty"`
	Text       string       `json:"text,omitempty"`
	Fields     []SlackField `json:"fields,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	FooterIcon string       `json:"footer_icon,omitempty"`
	Timestamp  int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

