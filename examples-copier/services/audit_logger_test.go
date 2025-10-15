package services

import (
	"context"
	"testing"
	"time"
)

func TestNewMongoAuditLogger_Disabled(t *testing.T) {
	ctx := context.Background()
	
	// When enabled=false, should return NoOpAuditLogger
	logger, err := NewMongoAuditLogger(ctx, "", "testdb", "testcoll", false)
	if err != nil {
		t.Fatalf("NewMongoAuditLogger() error = %v, want nil", err)
	}

	if logger == nil {
		t.Fatal("NewMongoAuditLogger() returned nil logger")
	}

	// Should be NoOpAuditLogger
	_, ok := logger.(*NoOpAuditLogger)
	if !ok {
		t.Errorf("Expected NoOpAuditLogger when disabled, got %T", logger)
	}
}

func TestNewMongoAuditLogger_EnabledWithoutURI(t *testing.T) {
	ctx := context.Background()
	
	// When enabled=true but no URI, should return error
	_, err := NewMongoAuditLogger(ctx, "", "testdb", "testcoll", true)
	if err == nil {
		t.Error("NewMongoAuditLogger() expected error when enabled without URI, got nil")
	}

	expectedMsg := "MONGO_URI is required when audit logging is enabled"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestNoOpAuditLogger_LogCopyEvent(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	event := &AuditEvent{
		EventType:    AuditEventCopy,
		RuleName:     "test-rule",
		SourceRepo:   "test/source",
		SourcePath:   "test.go",
		TargetRepo:   "test/target",
		TargetPath:   "copied/test.go",
		CommitSHA:    "abc123",
		PRNumber:     123,
		Success:      true,
		DurationMs:   100,
		FileSize:     1024,
	}

	err := logger.LogCopyEvent(ctx, event)
	if err != nil {
		t.Errorf("LogCopyEvent() error = %v, want nil", err)
	}
}

func TestNoOpAuditLogger_LogDeprecationEvent(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	event := &AuditEvent{
		EventType:  AuditEventDeprecation,
		SourceRepo: "test/source",
		SourcePath: "deprecated.go",
		PRNumber:   124,
		Success:    true,
	}

	err := logger.LogDeprecationEvent(ctx, event)
	if err != nil {
		t.Errorf("LogDeprecationEvent() error = %v, want nil", err)
	}
}

func TestNoOpAuditLogger_LogErrorEvent(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	event := &AuditEvent{
		EventType:    AuditEventError,
		SourceRepo:   "test/source",
		SourcePath:   "error.go",
		ErrorMessage: "test error",
		Success:      false,
	}

	err := logger.LogErrorEvent(ctx, event)
	if err != nil {
		t.Errorf("LogErrorEvent() error = %v, want nil", err)
	}
}

func TestNoOpAuditLogger_GetRecentEvents(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	events, err := logger.GetRecentEvents(ctx, 10)
	if err != nil {
		t.Errorf("GetRecentEvents() error = %v, want nil", err)
	}

	if events == nil {
		t.Error("GetRecentEvents() returned nil, want empty slice")
	}

	if len(events) != 0 {
		t.Errorf("GetRecentEvents() returned %d events, want 0", len(events))
	}
}

func TestNoOpAuditLogger_GetFailedEvents(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	events, err := logger.GetFailedEvents(ctx, 10)
	if err != nil {
		t.Errorf("GetFailedEvents() error = %v, want nil", err)
	}

	if events == nil {
		t.Error("GetFailedEvents() returned nil, want empty slice")
	}

	if len(events) != 0 {
		t.Errorf("GetFailedEvents() returned %d events, want 0", len(events))
	}
}

func TestNoOpAuditLogger_GetEventsByRule(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	events, err := logger.GetEventsByRule(ctx, "test-rule", 10)
	if err != nil {
		t.Errorf("GetEventsByRule() error = %v, want nil", err)
	}

	if events == nil {
		t.Error("GetEventsByRule() returned nil, want empty slice")
	}

	if len(events) != 0 {
		t.Errorf("GetEventsByRule() returned %d events, want 0", len(events))
	}
}

func TestNoOpAuditLogger_GetStatsByRule(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	stats, err := logger.GetStatsByRule(ctx)
	if err != nil {
		t.Errorf("GetStatsByRule() error = %v, want nil", err)
	}

	if stats == nil {
		t.Error("GetStatsByRule() returned nil, want empty map")
	}

	if len(stats) != 0 {
		t.Errorf("GetStatsByRule() returned %d stats, want 0", len(stats))
	}
}

func TestNoOpAuditLogger_GetDailyVolume(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	stats, err := logger.GetDailyVolume(ctx, 7)
	if err != nil {
		t.Errorf("GetDailyVolume() error = %v, want nil", err)
	}

	if stats == nil {
		t.Error("GetDailyVolume() returned nil, want empty slice")
	}

	if len(stats) != 0 {
		t.Errorf("GetDailyVolume() returned %d stats, want 0", len(stats))
	}
}

func TestNoOpAuditLogger_Close(t *testing.T) {
	logger := &NoOpAuditLogger{}
	ctx := context.Background()

	err := logger.Close(ctx)
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestAuditEventTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType AuditEventType
		expected  string
	}{
		{"copy event", AuditEventCopy, "copy"},
		{"deprecation event", AuditEventDeprecation, "deprecation"},
		{"error event", AuditEventError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("Event type = %v, want %v", tt.eventType, tt.expected)
			}
		})
	}
}

func TestAuditEvent_Structure(t *testing.T) {
	// Test that AuditEvent can be created with all fields
	now := time.Now()
	event := &AuditEvent{
		ID:             "test-id",
		Timestamp:      now,
		EventType:      AuditEventCopy,
		RuleName:       "test-rule",
		SourceRepo:     "test/source",
		SourcePath:     "source.go",
		TargetRepo:     "test/target",
		TargetPath:     "target.go",
		CommitSHA:      "abc123",
		PRNumber:       123,
		Success:        true,
		ErrorMessage:   "",
		DurationMs:     100,
		FileSize:       1024,
		AdditionalData: map[string]any{"key": "value"},
	}

	if event.EventType != AuditEventCopy {
		t.Errorf("EventType = %v, want %v", event.EventType, AuditEventCopy)
	}

	if event.Success != true {
		t.Error("Success should be true")
	}

	if event.PRNumber != 123 {
		t.Errorf("PRNumber = %d, want 123", event.PRNumber)
	}

	if event.AdditionalData["key"] != "value" {
		t.Error("AdditionalData not set correctly")
	}
}

func TestRuleStats_Structure(t *testing.T) {
	stats := RuleStats{
		RuleName:     "test-rule",
		TotalCopies:  100,
		SuccessCount: 95,
		FailureCount: 5,
		AvgDuration:  150.5,
	}

	if stats.RuleName != "test-rule" {
		t.Errorf("RuleName = %v, want test-rule", stats.RuleName)
	}

	if stats.TotalCopies != 100 {
		t.Errorf("TotalCopies = %d, want 100", stats.TotalCopies)
	}

	if stats.SuccessCount != 95 {
		t.Errorf("SuccessCount = %d, want 95", stats.SuccessCount)
	}

	if stats.FailureCount != 5 {
		t.Errorf("FailureCount = %d, want 5", stats.FailureCount)
	}
}

func TestDailyStats_Structure(t *testing.T) {
	stats := DailyStats{
		Date:         "2024-01-15",
		TotalCopies:  50,
		SuccessCount: 48,
		FailureCount: 2,
	}

	if stats.Date != "2024-01-15" {
		t.Errorf("Date = %v, want 2024-01-15", stats.Date)
	}

	if stats.TotalCopies != 50 {
		t.Errorf("TotalCopies = %d, want 50", stats.TotalCopies)
	}
}

