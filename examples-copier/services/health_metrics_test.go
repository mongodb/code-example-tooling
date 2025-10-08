package services_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mongodb/code-example-tooling/code-copier/services"
	"github.com/mongodb/code-example-tooling/code-copier/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCollector_WebhookMetrics(t *testing.T) {
	collector := services.NewMetricsCollector()

	// Record some webhooks
	collector.RecordWebhookReceived()
	collector.RecordWebhookReceived()
	collector.RecordWebhookReceived()

	collector.RecordWebhookProcessed(100 * time.Millisecond)
	collector.RecordWebhookProcessed(200 * time.Millisecond)

	collector.RecordWebhookFailed()

	// Get metrics
	fileStateService := services.NewFileStateService()
	metrics := collector.GetMetrics(fileStateService)

	assert.Equal(t, int64(3), metrics.Webhooks.Received)
	assert.Equal(t, int64(2), metrics.Webhooks.Processed)
	assert.Equal(t, int64(1), metrics.Webhooks.Failed)
	assert.InDelta(t, 66.67, metrics.Webhooks.SuccessRate, 0.1)
}

func TestMetricsCollector_FileMetrics(t *testing.T) {
	collector := services.NewMetricsCollector()

	// Record file operations
	collector.RecordFileMatched()
	collector.RecordFileMatched()
	collector.RecordFileMatched()

	collector.RecordFileUploaded(50 * time.Millisecond)
	collector.RecordFileUploaded(100 * time.Millisecond)

	collector.RecordFileUploadFailed()

	collector.RecordFileDeprecated()

	// Get metrics
	fileStateService := services.NewFileStateService()
	metrics := collector.GetMetrics(fileStateService)

	assert.Equal(t, int64(3), metrics.Files.Matched)
	assert.Equal(t, int64(2), metrics.Files.Uploaded)
	assert.Equal(t, int64(1), metrics.Files.UploadFailed)
	assert.Equal(t, int64(1), metrics.Files.Deprecated)
	assert.InDelta(t, 66.67, metrics.Files.UploadSuccessRate, 0.1)
}

func TestMetricsCollector_GitHubAPIMetrics(t *testing.T) {
	collector := services.NewMetricsCollector()

	// Record API calls
	collector.RecordGitHubAPICall()
	collector.RecordGitHubAPICall()
	collector.RecordGitHubAPICall()

	collector.RecordGitHubAPIError()

	// Get metrics
	fileStateService := services.NewFileStateService()
	metrics := collector.GetMetrics(fileStateService)

	assert.Equal(t, int64(3), metrics.GitHubAPI.Calls)
	assert.Equal(t, int64(1), metrics.GitHubAPI.Errors)
	// Error rate = errors / (calls + errors) = 1 / 4 = 25%
	assert.InDelta(t, 33.33, metrics.GitHubAPI.ErrorRate, 0.1)
}

func TestMetricsCollector_ProcessingTimePercentiles(t *testing.T) {
	collector := services.NewMetricsCollector()

	// Record processing times
	times := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}

	for _, d := range times {
		collector.RecordWebhookProcessed(d)
	}

	// Get metrics
	fileStateService := services.NewFileStateService()
	metrics := collector.GetMetrics(fileStateService)

	// Check percentiles are reasonable
	assert.Greater(t, metrics.Webhooks.ProcessingTime.P50Ms, float64(0))
	assert.Greater(t, metrics.Webhooks.ProcessingTime.P95Ms, metrics.Webhooks.ProcessingTime.P50Ms)
	assert.Greater(t, metrics.Webhooks.ProcessingTime.P99Ms, metrics.Webhooks.ProcessingTime.P95Ms)
}

func TestMetricsCollector_QueueSizes(t *testing.T) {
	collector := services.NewMetricsCollector()
	fileStateService := services.NewFileStateService()

	// Add some files to queues
	fileStateService.AddFileToUpload(
		types.UploadKey{RepoName: "org/repo", BranchPath: "refs/heads/main"},
		types.UploadFileContent{TargetBranch: "main"},
	)

	fileStateService.AddFileToDeprecate(
		"deprecated.json",
		types.DeprecatedFileEntry{FileName: "test.go"},
	)

	// Get metrics
	metrics := collector.GetMetrics(fileStateService)

	assert.Equal(t, 1, metrics.Queues.UploadQueueSize)
	assert.Equal(t, 1, metrics.Queues.DeprecationQueueSize)
}

func TestHealthHandler(t *testing.T) {
	fileStateService := services.NewFileStateService()
	startTime := time.Now().Add(-1 * time.Hour)

	handler := services.HealthHandler(fileStateService, startTime)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var health map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
	assert.True(t, health["started"].(bool))
	assert.NotNil(t, health["uptime"])
}

func TestMetricsHandler(t *testing.T) {
	collector := services.NewMetricsCollector()
	fileStateService := services.NewFileStateService()

	// Record some metrics
	collector.RecordWebhookReceived()
	collector.RecordWebhookProcessed(100 * time.Millisecond)
	collector.RecordFileMatched()
	collector.RecordFileUploaded(50 * time.Millisecond)

	handler := services.MetricsHandler(collector, fileStateService)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var metrics map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	require.NoError(t, err)

	webhooks := metrics["webhooks"].(map[string]interface{})
	files := metrics["files"].(map[string]interface{})

	assert.Equal(t, float64(1), webhooks["received"])
	assert.Equal(t, float64(1), webhooks["processed"])
	assert.Equal(t, float64(1), files["matched"])
	assert.Equal(t, float64(1), files["uploaded"])
}

func TestMetricsCollector_CircularBuffer(t *testing.T) {
	collector := services.NewMetricsCollector()

	// Record more than buffer size (1000) processing times
	for i := 0; i < 1500; i++ {
		collector.RecordWebhookProcessed(time.Duration(i) * time.Millisecond)
	}

	fileStateService := services.NewFileStateService()
	metrics := collector.GetMetrics(fileStateService)

	// Should still work and not crash
	assert.Greater(t, metrics.Webhooks.ProcessingTime.P50Ms, float64(0))
	assert.Greater(t, metrics.Webhooks.ProcessingTime.P95Ms, float64(0))
	assert.Greater(t, metrics.Webhooks.ProcessingTime.P99Ms, float64(0))
}

func TestMetricsCollector_ZeroValues(t *testing.T) {
	collector := services.NewMetricsCollector()
	fileStateService := services.NewFileStateService()

	// Get metrics without recording anything
	metrics := collector.GetMetrics(fileStateService)

	assert.Equal(t, int64(0), metrics.Webhooks.Received)
	assert.Equal(t, int64(0), metrics.Webhooks.Processed)
	assert.Equal(t, int64(0), metrics.Webhooks.Failed)
	assert.Equal(t, float64(0), metrics.Webhooks.SuccessRate)

	assert.Equal(t, int64(0), metrics.Files.Matched)
	assert.Equal(t, int64(0), metrics.Files.Uploaded)
	assert.Equal(t, int64(0), metrics.Files.UploadFailed)
	assert.Equal(t, float64(0), metrics.Files.UploadSuccessRate)
}

func TestMetricsCollector_SuccessRateCalculation(t *testing.T) {
	tests := []struct {
		name        string
		received    int
		processed   int
		failed      int
		wantRate    float64
	}{
		{
			name:      "all success",
			received:  10,
			processed: 10,
			failed:    0,
			wantRate:  100.0,
		},
		{
			name:      "all failed",
			received:  10,
			processed: 0,
			failed:    10,
			wantRate:  0.0,
		},
		{
			name:      "half success",
			received:  10,
			processed: 5,
			failed:    5,
			wantRate:  50.0,
		},
		{
			name:      "no operations",
			received:  0,
			processed: 0,
			failed:    0,
			wantRate:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := services.NewMetricsCollector()

			for i := 0; i < tt.received; i++ {
				collector.RecordWebhookReceived()
			}
			for i := 0; i < tt.processed; i++ {
				collector.RecordWebhookProcessed(10 * time.Millisecond)
			}
			for i := 0; i < tt.failed; i++ {
				collector.RecordWebhookFailed()
			}

			fileStateService := services.NewFileStateService()
			metrics := collector.GetMetrics(fileStateService)

			assert.InDelta(t, tt.wantRate, metrics.Webhooks.SuccessRate, 0.1)
		})
	}
}

func TestMetricsCollector_ConcurrentAccess(t *testing.T) {
	collector := services.NewMetricsCollector()
	fileStateService := services.NewFileStateService()

	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			collector.RecordWebhookReceived()
			collector.RecordWebhookProcessed(10 * time.Millisecond)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = collector.GetMetrics(fileStateService)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not crash and should have recorded metrics
	metrics := collector.GetMetrics(fileStateService)
	assert.Greater(t, metrics.Webhooks.Received, int64(0))
}

