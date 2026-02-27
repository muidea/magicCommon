package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

// Exporter exports metrics via HTTP in various formats
type Exporter struct {
	mu sync.RWMutex

	collector *Collector
	config    *ExportConfig
	server    *http.Server

	// Custom labels applied to all metrics
	customLabels map[string]string

	// Statistics
	stats ExporterStats

	// Cache for formatted metrics
	cache struct {
		prometheus     string
		json           string
		metadata       string
		singleMetadata map[string]string
		prometheusTime time.Time
		jsonTime       time.Time
		metadataTime   time.Time
		mu             sync.RWMutex
	}
}

// ExporterStats holds exporter statistics
type ExporterStats struct {
	RequestsTotal  int64            `json:"requests_total"`
	RequestsByPath map[string]int64 `json:"requests_by_path"`
	ErrorsTotal    int64            `json:"errors_total"`
	LastRequest    time.Time        `json:"last_request"`
	StartTime      time.Time        `json:"start_time"`
	Uptime         time.Duration    `json:"uptime"`
	CacheHits      int64            `json:"cache_hits"`
	CacheMisses    int64            `json:"cache_misses"`
}

// NewExporter creates a new metrics exporter
func NewExporter(collector *Collector, config *ExportConfig) (*Exporter, *types.Error) {
	if collector == nil {
		return nil, types.NewCollectorNotInitializedError()
	}

	if config == nil {
		defaultConfig := DefaultExportConfig()
		config = &defaultConfig
	}

	exporter := &Exporter{
		collector:    collector,
		config:       config,
		customLabels: make(map[string]string),
		stats: ExporterStats{
			StartTime:      time.Now(),
			RequestsByPath: make(map[string]int64),
		},
	}

	// Initialize cache
	exporter.cache.prometheusTime = time.Time{}
	exporter.cache.jsonTime = time.Time{}
	exporter.cache.metadataTime = time.Time{}
	exporter.cache.singleMetadata = make(map[string]string)

	return exporter, nil
}

// Start starts the HTTP server for metric export
// This method is thread-safe.
func (e *Exporter) Start() *types.Error {
	if !e.config.Enabled {
		return nil
	}

	mux := http.NewServeMux()

	// Register handlers
	if e.config.EnablePrometheus {
		mux.HandleFunc(e.config.Path, e.prometheusHandler)
	}

	if e.config.EnableJSON {
		mux.HandleFunc(e.config.MetricsPath, e.jsonHandler)
	}

	mux.HandleFunc(e.config.HealthCheckPath, e.healthHandler)
	mux.HandleFunc(e.config.InfoPath, e.infoHandler)
	mux.HandleFunc(e.config.MetadataPath, e.metadataHandler)
	mux.HandleFunc(e.config.MetadataPath+"/", e.singleMetadataHandler)

	// Add middleware for authentication and statistics
	handler := e.withMiddleware(mux)

	e.mu.Lock()
	e.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", e.config.Port),
		Handler:      handler,
		ReadTimeout:  e.config.ScrapeTimeout,
		WriteTimeout: e.config.ScrapeTimeout,
	}
	server := e.server
	e.mu.Unlock()

	// Start server in background
	// Capture server reference to avoid race condition with Stop()
	go func() {
		var err error
		if e.config.EnableTLS {
			err = server.ListenAndServeTLS(e.config.TLSCertPath, e.config.TLSKeyPath)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			// Log error (in production, you'd want proper logging)
			slog.Error("Exporter server error", "error", err)
		}
	}()

	return nil
}

// Stop stops the HTTP server
// This method is thread-safe.
func (e *Exporter) Stop() *types.Error {
	e.mu.Lock()
	server := e.server
	e.server = nil
	e.mu.Unlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return types.NewExportFailedError("failed to shutdown server: " + err.Error())
	}

	// Clear cache on shutdown
	e.cache.mu.Lock()
	e.cache.prometheus = ""
	e.cache.json = ""
	e.cache.metadata = ""
	e.cache.singleMetadata = make(map[string]string)
	e.cache.prometheusTime = time.Time{}
	e.cache.jsonTime = time.Time{}
	e.cache.metadataTime = time.Time{}
	e.cache.mu.Unlock()

	return nil
}

// SetCustomLabel sets a custom label to be applied to all metrics
func (e *Exporter) SetCustomLabel(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.customLabels[key] = value
	// Invalidate cache when labels change
	e.cache.mu.Lock()
	e.cache.prometheusTime = time.Time{}
	e.cache.jsonTime = time.Time{}
	e.cache.metadataTime = time.Time{}
	e.cache.singleMetadata = make(map[string]string)
	e.cache.mu.Unlock()
}

// RemoveCustomLabel removes a custom label
func (e *Exporter) RemoveCustomLabel(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.customLabels, key)
	// Invalidate cache when labels change
	e.cache.mu.Lock()
	e.cache.prometheusTime = time.Time{}
	e.cache.jsonTime = time.Time{}
	e.cache.metadataTime = time.Time{}
	e.cache.singleMetadata = make(map[string]string)
	e.cache.mu.Unlock()
}

// GetStats returns exporter statistics
func (e *Exporter) GetStats() ExporterStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := e.stats
	stats.Uptime = time.Since(stats.StartTime)
	return stats
}

// ExportPrometheus exports metrics in Prometheus format
// This method is thread-safe and uses caching for performance.
func (e *Exporter) ExportPrometheus() (string, *types.Error) {
	// Check cache first
	e.cache.mu.RLock()
	if !e.cache.prometheusTime.IsZero() && time.Since(e.cache.prometheusTime) < e.config.RefreshInterval {
		e.cache.mu.RUnlock()
		e.mu.Lock()
		e.stats.CacheHits++
		e.mu.Unlock()
		return e.cache.prometheus, nil
	}
	e.cache.mu.RUnlock()

	// Cache miss, generate new export
	e.mu.Lock()
	e.stats.CacheMisses++
	e.mu.Unlock()

	metrics := e.collector.GetMetrics()
	definitions := e.collector.GetDefinitions()

	var builder strings.Builder

	// Export metric definitions (HELP and TYPE lines)
	for name, def := range definitions {
		// HELP line
		fmt.Fprintf(&builder, "# HELP %s %s\n", name, def.Help)

		// TYPE line
		fmt.Fprintf(&builder, "# TYPE %s %s\n", name, strings.ToLower(string(def.Type)))
	}

	// Export metric values
	for name, metricList := range metrics {
		if _, exists := definitions[name]; !exists {
			continue // Skip metrics without definitions
		}

		for _, metric := range metricList {
			// Merge custom labels with metric labels
			labels := e.mergeLabels(metric.Labels)

			// Build label string
			labelStr := e.buildLabelString(labels)

			// Write metric line
			value := formatFloat(metric.Value)
			if labelStr != "" {
				fmt.Fprintf(&builder, "%s{%s} %s\n", name, labelStr, value)
			} else {
				fmt.Fprintf(&builder, "%s %s\n", name, value)
			}
		}
	}

	result := builder.String()

	// Update cache
	e.cache.mu.Lock()
	e.cache.prometheus = result
	e.cache.prometheusTime = time.Now()
	e.cache.mu.Unlock()

	return result, nil
}

// ExportJSON exports metrics in JSON format
func (e *Exporter) ExportJSON() (string, *types.Error) {
	// Check cache first
	e.cache.mu.RLock()
	if !e.cache.jsonTime.IsZero() && time.Since(e.cache.jsonTime) < e.config.RefreshInterval {
		e.cache.mu.RUnlock()
		e.mu.Lock()
		e.stats.CacheHits++
		e.mu.Unlock()
		return e.cache.json, nil
	}
	e.cache.mu.RUnlock()

	// Cache miss, generate new export
	e.mu.Lock()
	e.stats.CacheMisses++
	e.mu.Unlock()

	metrics := e.collector.GetMetrics()
	definitions := e.collector.GetDefinitions()

	// Build JSON structure
	exportData := struct {
		Timestamp      time.Time                         `json:"timestamp"`
		Metrics        map[string][]types.Metric         `json:"metrics"`
		Definitions    map[string]types.MetricDefinition `json:"definitions"`
		Stats          ExporterStats                     `json:"stats"`
		CollectorStats CollectorStats                    `json:"collector_stats"`
	}{
		Timestamp:      time.Now(),
		Metrics:        metrics,
		Definitions:    definitions,
		Stats:          e.GetStats(),
		CollectorStats: e.collector.GetStats(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", types.NewExportFailedError("failed to marshal JSON: " + err.Error())
	}

	result := string(data)

	// Update cache
	e.cache.mu.Lock()
	e.cache.json = result
	e.cache.jsonTime = time.Now()
	e.cache.mu.Unlock()

	return result, nil
}

// HTTP handlers

func (e *Exporter) prometheusHandler(w http.ResponseWriter, r *http.Request) {
	e.updateRequestStats(r.URL.Path)

	metrics, err := e.ExportPrometheus()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		e.mu.Lock()
		e.stats.ErrorsTotal++
		e.mu.Unlock()
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(metrics))
}

func (e *Exporter) jsonHandler(w http.ResponseWriter, r *http.Request) {
	e.updateRequestStats(r.URL.Path)

	metrics, err := e.ExportJSON()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		e.mu.Lock()
		e.stats.ErrorsTotal++
		e.mu.Unlock()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(metrics))
}

func (e *Exporter) healthHandler(w http.ResponseWriter, r *http.Request) {
	e.updateRequestStats(r.URL.Path)

	healthStatus := struct {
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
		Uptime    string    `json:"uptime"`
	}{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(e.stats.StartTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthStatus)
}

func (e *Exporter) infoHandler(w http.ResponseWriter, r *http.Request) {
	e.updateRequestStats(r.URL.Path)

	info := struct {
		Name        string    `json:"name"`
		Version     string    `json:"version"`
		Description string    `json:"description"`
		StartTime   time.Time `json:"start_time"`
		Uptime      string    `json:"uptime"`
		Endpoints   []string  `json:"endpoints"`
	}{
		Name:        "MagicORM Monitoring Exporter",
		Version:     "1.0.0",
		Description: "Exports monitoring metrics in various formats",
		StartTime:   e.stats.StartTime,
		Uptime:      time.Since(e.stats.StartTime).String(),
		Endpoints:   e.getEndpoints(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(info)
}

// ExportMetadata exports metric metadata in JSON format
func (e *Exporter) ExportMetadata() (string, *types.Error) {
	// Check cache first
	e.cache.mu.RLock()
	if !e.cache.metadataTime.IsZero() && time.Since(e.cache.metadataTime) < e.config.RefreshInterval {
		e.cache.mu.RUnlock()
		e.mu.Lock()
		e.stats.CacheHits++
		e.mu.Unlock()
		return e.cache.metadata, nil
	}
	e.cache.mu.RUnlock()

	// Cache miss, generate new export
	e.mu.Lock()
	e.stats.CacheMisses++
	e.mu.Unlock()

	// Get all metric definitions from collector
	definitions := e.collector.GetDefinitions()

	// Build metadata response
	metadata := make(map[string]types.MetricMetadata)
	for name, def := range definitions {
		metadata[name] = types.NewMetricMetadata(def)
	}

	response := struct {
		Timestamp time.Time                       `json:"timestamp"`
		Metrics   map[string]types.MetricMetadata `json:"metrics"`
		Count     int                             `json:"count"`
	}{
		Timestamp: time.Now(),
		Metrics:   metadata,
		Count:     len(metadata),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", types.NewExportFailedError("failed to marshal metadata: " + err.Error())
	}

	result := string(data)

	// Update cache
	e.cache.mu.Lock()
	e.cache.metadata = result
	e.cache.metadataTime = time.Now()
	e.cache.mu.Unlock()

	return result, nil
}

func (e *Exporter) metadataHandler(w http.ResponseWriter, r *http.Request) {
	e.updateRequestStats(r.URL.Path)

	metadata, err := e.ExportMetadata()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		e.mu.Lock()
		e.stats.ErrorsTotal++
		e.mu.Unlock()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(metadata))
}

// ExportSingleMetadata exports metadata for a single metric
func (e *Exporter) ExportSingleMetadata(metricName string) (string, *types.Error) {
	// Check cache first
	e.cache.mu.RLock()
	if cached, exists := e.cache.singleMetadata[metricName]; exists {
		// Check if cache is still valid
		if !e.cache.metadataTime.IsZero() && time.Since(e.cache.metadataTime) < e.config.RefreshInterval {
			e.cache.mu.RUnlock()
			e.mu.Lock()
			e.stats.CacheHits++
			e.mu.Unlock()
			return cached, nil
		}
		// Cache expired, remove it
		delete(e.cache.singleMetadata, metricName)
	}
	e.cache.mu.RUnlock()

	// Cache miss, generate new export
	e.mu.Lock()
	e.stats.CacheMisses++
	e.mu.Unlock()

	// Get metric definition
	def, err := e.collector.GetDefinition(metricName)
	if err != nil {
		return "", err
	}

	// Create metadata from definition
	metadata := types.NewMetricMetadata(def)

	// Marshal to JSON
	data, marshalErr := json.MarshalIndent(metadata, "", "  ")
	if marshalErr != nil {
		return "", types.NewExportFailedError("failed to marshal single metadata: " + marshalErr.Error())
	}

	result := string(data)

	// Update cache
	e.cache.mu.Lock()
	e.cache.singleMetadata[metricName] = result
	// Also update global metadata timestamp for cache validation
	e.cache.metadataTime = time.Now()
	e.cache.mu.Unlock()

	return result, nil
}

func (e *Exporter) singleMetadataHandler(w http.ResponseWriter, r *http.Request) {
	e.updateRequestStats(r.URL.Path)

	// Extract metric name from URL path
	metricName := strings.TrimPrefix(r.URL.Path, e.config.MetadataPath+"/")
	if metricName == "" {
		http.Error(w, "Metric name required", http.StatusBadRequest)
		return
	}

	// Validate metric name format
	if !types.IsValidMetricName(metricName) {
		http.Error(w, "Invalid metric name format", http.StatusBadRequest)
		return
	}

	// Get single metric metadata
	metadata, err := e.ExportSingleMetadata(metricName)
	if err != nil {
		if err.Code == types.MetricNotFound {
			http.Error(w, "Metric not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			e.mu.Lock()
			e.stats.ErrorsTotal++
			e.mu.Unlock()
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(metadata))
}

// Middleware

func (e *Exporter) withMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check allowed hosts
		if !e.checkHost(r) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			e.mu.Lock()
			e.stats.ErrorsTotal++
			e.mu.Unlock()
			return
		}

		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Add CORS headers if needed
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the actual handler
		handler.ServeHTTP(w, r)
	})
}

// Helper methods

func (e *Exporter) updateRequestStats(path string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stats.RequestsTotal++
	e.stats.RequestsByPath[path]++
	e.stats.LastRequest = time.Now()
}

func (e *Exporter) checkHost(r *http.Request) bool {
	if len(e.config.AllowedHosts) == 0 {
		return true
	}

	host := r.Host
	if host == "" {
		host = r.RemoteAddr
	}

	for _, allowedHost := range e.config.AllowedHosts {
		if host == allowedHost || strings.HasPrefix(host, allowedHost+":") {
			return true
		}
	}

	return false
}

func (e *Exporter) mergeLabels(metricLabels map[string]string) map[string]string {
	e.mu.RLock()
	customLabels := e.customLabels
	e.mu.RUnlock()

	if len(customLabels) == 0 {
		return metricLabels
	}

	merged := make(map[string]string)
	for k, v := range metricLabels {
		merged[k] = v
	}
	for k, v := range customLabels {
		merged[k] = v
	}

	return merged
}

func (e *Exporter) buildLabelString(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := labels[k]
		// Escape special characters in label values
		escapedValue := strings.ReplaceAll(v, `\`, `\\`)
		escapedValue = strings.ReplaceAll(escapedValue, `"`, `\"`)
		escapedValue = strings.ReplaceAll(escapedValue, "\n", `\n`)
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, escapedValue))
	}

	return strings.Join(parts, ",")
}

func (e *Exporter) getEndpoints() []string {
	endpoints := []string{
		e.config.HealthCheckPath,
		e.config.InfoPath,
		e.config.MetadataPath,
	}

	if e.config.EnablePrometheus {
		endpoints = append(endpoints, e.config.Path)
	}

	if e.config.EnableJSON {
		endpoints = append(endpoints, e.config.MetricsPath)
	}

	return endpoints
}

func formatFloat(value float64) string {
	// Format float for Prometheus export
	// Prometheus recommends not using scientific notation
	return fmt.Sprintf("%.6f", value)
}
