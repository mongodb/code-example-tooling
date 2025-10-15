package services

import (
	"context"
	"testing"
	"time"

	"github.com/mongodb/code-example-tooling/code-copier/configs"
)

func TestNewServiceContainer(t *testing.T) {
	tests := []struct {
		name          string
		config        *configs.Config
		wantErr       bool
		checkServices bool
	}{
		{
			name: "valid config with audit disabled",
			config: &configs.Config{
				RepoOwner:      "test-owner",
				RepoName:       "test-repo",
				AuditEnabled:   false,
				SlackWebhookURL: "",
			},
			wantErr:       false,
			checkServices: true,
		},
		{
			name: "valid config with Slack enabled",
			config: &configs.Config{
				RepoOwner:       "test-owner",
				RepoName:        "test-repo",
				AuditEnabled:    false,
				SlackWebhookURL: "https://hooks.slack.com/services/TEST",
				SlackChannel:    "#test",
				SlackUsername:   "Test Bot",
				SlackIconEmoji:  ":robot:",
			},
			wantErr:       false,
			checkServices: true,
		},
		{
			name: "audit enabled without URI",
			config: &configs.Config{
				RepoOwner:      "test-owner",
				RepoName:       "test-repo",
				AuditEnabled:   true,
				MongoURI:       "",
			},
			wantErr:       true,
			checkServices: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := NewServiceContainer(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewServiceContainer() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewServiceContainer() error = %v, want nil", err)
			}

			if container == nil {
				t.Fatal("NewServiceContainer() returned nil container")
			}

			if tt.checkServices {
				// Check that all services are initialized
				if container.Config == nil {
					t.Error("Config is nil")
				}

				if container.FileStateService == nil {
					t.Error("FileStateService is nil")
				}

				if container.ConfigLoader == nil {
					t.Error("ConfigLoader is nil")
				}

				if container.PatternMatcher == nil {
					t.Error("PatternMatcher is nil")
				}

				if container.PathTransformer == nil {
					t.Error("PathTransformer is nil")
				}

				if container.MessageTemplater == nil {
					t.Error("MessageTemplater is nil")
				}

				if container.AuditLogger == nil {
					t.Error("AuditLogger is nil")
				}

				if container.MetricsCollector == nil {
					t.Error("MetricsCollector is nil")
				}

				if container.SlackNotifier == nil {
					t.Error("SlackNotifier is nil")
				}

				// Check that StartTime is set
				if container.StartTime.IsZero() {
					t.Error("StartTime is zero")
				}

				// Check that StartTime is recent (within last second)
				if time.Since(container.StartTime) > time.Second {
					t.Error("StartTime is not recent")
				}
			}
		})
	}
}

func TestServiceContainer_Close(t *testing.T) {
	tests := []struct {
		name        string
		config      *configs.Config
		wantErr     bool
	}{
		{
			name: "close with NoOp audit logger",
			config: &configs.Config{
				RepoOwner:      "test-owner",
				RepoName:       "test-repo",
				AuditEnabled:   false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := NewServiceContainer(tt.config)
			if err != nil {
				t.Fatalf("NewServiceContainer() error = %v", err)
			}

			ctx := context.Background()
			err = container.Close(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("Close() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Close() error = %v, want nil", err)
				}
			}
		})
	}
}

func TestServiceContainer_ConfigPropagation(t *testing.T) {
	config := &configs.Config{
		RepoOwner:       "test-owner",
		RepoName:        "test-repo",
		AuditEnabled:    false,
		SlackWebhookURL: "https://hooks.slack.com/services/TEST",
		SlackChannel:    "#test-channel",
		SlackUsername:   "Test Bot",
		SlackIconEmoji:  ":robot:",
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	// Verify config is stored correctly
	if container.Config != config {
		t.Error("Config not stored correctly in container")
	}

	if container.Config.RepoOwner != "test-owner" {
		t.Errorf("RepoOwner = %v, want test-owner", container.Config.RepoOwner)
	}

	if container.Config.SlackChannel != "#test-channel" {
		t.Errorf("SlackChannel = %v, want #test-channel", container.Config.SlackChannel)
	}
}

func TestServiceContainer_SlackNotifierConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		webhookURL      string
		channel         string
		username        string
		iconEmoji       string
		wantEnabled     bool
	}{
		{
			name:        "Slack enabled",
			webhookURL:  "https://hooks.slack.com/services/TEST",
			channel:     "#test",
			username:    "Bot",
			iconEmoji:   ":robot:",
			wantEnabled: true,
		},
		{
			name:        "Slack disabled",
			webhookURL:  "",
			channel:     "",
			username:    "",
			iconEmoji:   "",
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configs.Config{
				RepoOwner:       "test-owner",
				RepoName:        "test-repo",
				AuditEnabled:    false,
				SlackWebhookURL: tt.webhookURL,
				SlackChannel:    tt.channel,
				SlackUsername:   tt.username,
				SlackIconEmoji:  tt.iconEmoji,
			}

			container, err := NewServiceContainer(config)
			if err != nil {
				t.Fatalf("NewServiceContainer() error = %v", err)
			}

			if container.SlackNotifier.IsEnabled() != tt.wantEnabled {
				t.Errorf("SlackNotifier.IsEnabled() = %v, want %v",
					container.SlackNotifier.IsEnabled(), tt.wantEnabled)
			}
		})
	}
}

func TestServiceContainer_AuditLoggerConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		auditEnabled bool
		mongoURI     string
		wantType     string
		wantErr      bool
	}{
		{
			name:         "audit disabled",
			auditEnabled: false,
			mongoURI:     "",
			wantType:     "*services.NoOpAuditLogger",
			wantErr:      false,
		},
		{
			name:         "audit enabled without URI",
			auditEnabled: true,
			mongoURI:     "",
			wantType:     "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configs.Config{
				RepoOwner:      "test-owner",
				RepoName:       "test-repo",
				AuditEnabled:   tt.auditEnabled,
				MongoURI:       tt.mongoURI,
				AuditDatabase:  "test-db",
				AuditCollection: "test-coll",
			}

			container, err := NewServiceContainer(config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewServiceContainer() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewServiceContainer() error = %v", err)
			}

			// Check audit logger type - NoOp should be returned when disabled
			_, isNoOp := container.AuditLogger.(*NoOpAuditLogger)
			if tt.wantType == "*services.NoOpAuditLogger" && !isNoOp {
				t.Error("Expected NoOpAuditLogger when audit is disabled")
			}
		})
	}
}

func TestServiceContainer_MetricsCollectorInitialization(t *testing.T) {
	config := &configs.Config{
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		AuditEnabled:   false,
	}

	container, err := NewServiceContainer(config)
	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	if container.MetricsCollector == nil {
		t.Fatal("MetricsCollector is nil")
	}

	// Verify metrics collector is functional
	container.MetricsCollector.RecordWebhookReceived()
	container.MetricsCollector.RecordWebhookProcessed(time.Second)

	// Check that metrics were recorded using GetMetrics
	metrics := container.MetricsCollector.GetMetrics(container.FileStateService)
	if metrics.Webhooks.Received != 1 {
		t.Errorf("WebhooksReceived = %d, want 1", metrics.Webhooks.Received)
	}

	if metrics.Webhooks.Processed != 1 {
		t.Errorf("WebhooksProcessed = %d, want 1", metrics.Webhooks.Processed)
	}
}

func TestServiceContainer_StartTimeTracking(t *testing.T) {
	config := &configs.Config{
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		AuditEnabled:   false,
	}

	beforeCreate := time.Now()
	container, err := NewServiceContainer(config)
	afterCreate := time.Now()

	if err != nil {
		t.Fatalf("NewServiceContainer() error = %v", err)
	}

	// StartTime should be between beforeCreate and afterCreate
	if container.StartTime.Before(beforeCreate) {
		t.Error("StartTime is before container creation")
	}
	if container.StartTime.After(afterCreate) {
		t.Error("StartTime is after container creation")
	}
}

