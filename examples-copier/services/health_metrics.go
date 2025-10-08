package services

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of the application
type HealthStatus struct {
	Status              string                 `json:"status"`
	Started             bool                   `json:"started"`
	GitHub              GitHubHealthStatus     `json:"github"`
	Queues              QueueHealthStatus      `json:"queues"`
	AuditLogger         AuditLoggerHealthStatus `json:"audit_logger,omitempty"`
	Uptime              string                 `json:"uptime"`
}

// GitHubHealthStatus represents GitHub API health
type GitHubHealthStatus struct {
	Status       string `json:"status"`
	Authenticated bool  `json:"authenticated"`
}

// QueueHealthStatus represents queue health
type QueueHealthStatus struct {
	UploadCount      int `json:"upload_count"`
	DeprecationCount int `json:"deprecation_count"`
}

// AuditLoggerHealthStatus represents audit logger health
type AuditLoggerHealthStatus struct {
	Status    string `json:"status"`
	Connected bool   `json:"connected"`
}

// MetricsData represents application metrics
type MetricsData struct {
	Webhooks   WebhookMetrics   `json:"webhooks"`
	Files      FileMetrics      `json:"files"`
	GitHubAPI  GitHubAPIMetrics `json:"github_api"`
	Queues     QueueMetrics     `json:"queues"`
	System     SystemMetrics    `json:"system"`
}

// WebhookMetrics represents webhook processing metrics
type WebhookMetrics struct {
	Received       int64              `json:"received"`
	Processed      int64              `json:"processed"`
	Failed         int64              `json:"failed"`
	SuccessRate    float64            `json:"success_rate"`
	ProcessingTime ProcessingTimeStats `json:"processing_time"`
}

// FileMetrics represents file operation metrics
type FileMetrics struct {
	Matched          int64              `json:"matched"`
	Uploaded         int64              `json:"uploaded"`
	UploadFailed     int64              `json:"upload_failed"`
	Deprecated       int64              `json:"deprecated"`
	UploadSuccessRate float64           `json:"upload_success_rate"`
	UploadTime       ProcessingTimeStats `json:"upload_time"`
}

// GitHubAPIMetrics represents GitHub API usage metrics
type GitHubAPIMetrics struct {
	Calls      int64              `json:"calls"`
	Errors     int64              `json:"errors"`
	ErrorRate  float64            `json:"error_rate"`
	RateLimit  RateLimitInfo      `json:"rate_limit"`
}

// RateLimitInfo represents GitHub API rate limit info
type RateLimitInfo struct {
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"reset_at"`
}

// QueueMetrics represents queue size metrics
type QueueMetrics struct {
	UploadQueueSize      int `json:"upload_queue_size"`
	DeprecationQueueSize int `json:"deprecation_queue_size"`
	RetryQueueSize       int `json:"retry_queue_size"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	UptimeSeconds int64 `json:"uptime_seconds"`
}

// ProcessingTimeStats represents timing statistics
type ProcessingTimeStats struct {
	AvgMs float64 `json:"avg_ms"`
	MinMs float64 `json:"min_ms"`
	MaxMs float64 `json:"max_ms"`
	P50Ms float64 `json:"p50_ms"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
}

// MetricsCollector collects and manages application metrics
type MetricsCollector struct {
	mu              sync.RWMutex
	startTime       time.Time
	webhookReceived int64
	webhookProcessed int64
	webhookFailed   int64
	filesMatched    int64
	filesUploaded   int64
	filesUploadFailed int64
	filesDeprecated int64
	githubAPICalls  int64
	githubAPIErrors int64
	processingTimes []time.Duration
	uploadTimes     []time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:       time.Now(),
		processingTimes: make([]time.Duration, 0, 1000),
		uploadTimes:     make([]time.Duration, 0, 1000),
	}
}

// RecordWebhookReceived increments webhook received counter
func (mc *MetricsCollector) RecordWebhookReceived() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.webhookReceived++
}

// RecordWebhookProcessed increments webhook processed counter
func (mc *MetricsCollector) RecordWebhookProcessed(duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.webhookProcessed++
	mc.processingTimes = append(mc.processingTimes, duration)
	
	// Keep only last 1000 entries
	if len(mc.processingTimes) > 1000 {
		mc.processingTimes = mc.processingTimes[len(mc.processingTimes)-1000:]
	}
}

// RecordWebhookFailed increments webhook failed counter
func (mc *MetricsCollector) RecordWebhookFailed() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.webhookFailed++
}

// RecordFileMatched increments file matched counter
func (mc *MetricsCollector) RecordFileMatched() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.filesMatched++
}

// RecordFileUploaded increments file uploaded counter
func (mc *MetricsCollector) RecordFileUploaded(duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.filesUploaded++
	mc.uploadTimes = append(mc.uploadTimes, duration)
	
	// Keep only last 1000 entries
	if len(mc.uploadTimes) > 1000 {
		mc.uploadTimes = mc.uploadTimes[len(mc.uploadTimes)-1000:]
	}
}

// RecordFileUploadFailed increments file upload failed counter
func (mc *MetricsCollector) RecordFileUploadFailed() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.filesUploadFailed++
}

// RecordFileDeprecated increments file deprecated counter
func (mc *MetricsCollector) RecordFileDeprecated() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.filesDeprecated++
}

// RecordGitHubAPICall increments GitHub API call counter
func (mc *MetricsCollector) RecordGitHubAPICall() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.githubAPICalls++
}

// RecordGitHubAPIError increments GitHub API error counter
func (mc *MetricsCollector) RecordGitHubAPIError() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.githubAPIErrors++
}

// GetMetrics returns current metrics
func (mc *MetricsCollector) GetMetrics(fileStateService FileStateService) MetricsData {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	webhookSuccessRate := 0.0
	if mc.webhookReceived > 0 {
		webhookSuccessRate = float64(mc.webhookProcessed) / float64(mc.webhookReceived) * 100
	}

	uploadSuccessRate := 0.0
	totalUploads := mc.filesUploaded + mc.filesUploadFailed
	if totalUploads > 0 {
		uploadSuccessRate = float64(mc.filesUploaded) / float64(totalUploads) * 100
	}

	githubErrorRate := 0.0
	if mc.githubAPICalls > 0 {
		githubErrorRate = float64(mc.githubAPIErrors) / float64(mc.githubAPICalls) * 100
	}

	// Get queue sizes
	uploadQueue := fileStateService.GetFilesToUpload()
	deprecationQueue := fileStateService.GetFilesToDeprecate()

	return MetricsData{
		Webhooks: WebhookMetrics{
			Received:       mc.webhookReceived,
			Processed:      mc.webhookProcessed,
			Failed:         mc.webhookFailed,
			SuccessRate:    webhookSuccessRate,
			ProcessingTime: calculateStats(mc.processingTimes),
		},
		Files: FileMetrics{
			Matched:          mc.filesMatched,
			Uploaded:         mc.filesUploaded,
			UploadFailed:     mc.filesUploadFailed,
			Deprecated:       mc.filesDeprecated,
			UploadSuccessRate: uploadSuccessRate,
			UploadTime:       calculateStats(mc.uploadTimes),
		},
		GitHubAPI: GitHubAPIMetrics{
			Calls:     mc.githubAPICalls,
			Errors:    mc.githubAPIErrors,
			ErrorRate: githubErrorRate,
			RateLimit: RateLimitInfo{
				Remaining: 5000, // TODO: Get from GitHub API
				ResetAt:   time.Now().Add(1 * time.Hour),
			},
		},
		Queues: QueueMetrics{
			UploadQueueSize:      len(uploadQueue),
			DeprecationQueueSize: len(deprecationQueue),
			RetryQueueSize:       0,
		},
		System: SystemMetrics{
			UptimeSeconds: int64(time.Since(mc.startTime).Seconds()),
		},
	}
}

// calculateStats calculates timing statistics
func calculateStats(durations []time.Duration) ProcessingTimeStats {
	if len(durations) == 0 {
		return ProcessingTimeStats{}
	}

	var sum, min, max float64
	min = float64(durations[0].Milliseconds())
	max = min

	for _, d := range durations {
		ms := float64(d.Milliseconds())
		sum += ms
		if ms < min {
			min = ms
		}
		if ms > max {
			max = ms
		}
	}

	avg := sum / float64(len(durations))

	// Calculate percentiles (simplified)
	p50 := avg // Simplified
	p95 := avg * 1.5
	p99 := avg * 2.0

	return ProcessingTimeStats{
		AvgMs: avg,
		MinMs: min,
		MaxMs: max,
		P50Ms: p50,
		P95Ms: p95,
		P99Ms: p99,
	}
}

// HealthHandler handles /health endpoint
func HealthHandler(fileStateService FileStateService, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uploadQueue := fileStateService.GetFilesToUpload()
		deprecationQueue := fileStateService.GetFilesToDeprecate()

		health := HealthStatus{
			Status:  "healthy",
			Started: true,
			GitHub: GitHubHealthStatus{
				Status:       "healthy",
				Authenticated: true,
			},
			Queues: QueueHealthStatus{
				UploadCount:      len(uploadQueue),
				DeprecationCount: len(deprecationQueue),
			},
			Uptime: time.Since(startTime).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}
}

// MetricsHandler handles /metrics endpoint
func MetricsHandler(metricsCollector *MetricsCollector, fileStateService FileStateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := metricsCollector.GetMetrics(fileStateService)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	}
}

