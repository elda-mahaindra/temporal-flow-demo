package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// MetricsServer provides a dedicated HTTP server for Prometheus metrics collection
type MetricsServer struct {
	logger *logrus.Logger
	server *http.Server
	port   int
}

// NewMetricsServer creates a new metrics server instance
func NewMetricsServer(logger *logrus.Logger, port int) *MetricsServer {
	return &MetricsServer{
		logger: logger,
		port:   port,
	}
}

// Start begins serving metrics on the dedicated port
func (ms *MetricsServer) Start(ctx context.Context) error {
	const op = "MetricsServer.Start"

	// Create HTTP mux for metrics endpoints
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.HandleFunc("/metrics", ms.handleMetrics)

	// Health check for metrics server itself
	mux.HandleFunc("/metrics/health", ms.handleMetricsHealth)

	// Create HTTP server
	ms.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ms.port),
		Handler: mux,
	}

	logger := ms.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"port":   ms.port,
		"server": "metrics",
	})

	logger.Info("Starting metrics server for Prometheus")

	// Start server in goroutine
	go func() {
		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("Metrics server failed")
		}
	}()

	logger.Info("✅ Metrics server started successfully")

	// Wait for context cancellation
	<-ctx.Done()
	return ms.Shutdown()
}

// Shutdown gracefully stops the metrics server
func (ms *MetricsServer) Shutdown() error {
	const op = "MetricsServer.Shutdown"

	logger := ms.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"server": "metrics",
	})

	logger.Info("Shutting down metrics server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ms.server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Failed to shutdown metrics server gracefully")
		return err
	}

	logger.Info("Metrics server shutdown completed")
	return nil
}

// handleMetrics provides Prometheus-compatible metrics
func (ms *MetricsServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	const op = "MetricsServer.handleMetrics"

	logger := ms.logger.WithField("[op]", op)

	// Set Prometheus content type
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// Get runtime stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Generate Prometheus format metrics
	metrics := fmt.Sprintf(`# HELP api_gateway_info Service information
# TYPE api_gateway_info gauge
api_gateway_info{service="api-gateway",version="1.0.0"} 1

# HELP api_gateway_up Service health status
# TYPE api_gateway_up gauge
api_gateway_up 1

# HELP api_gateway_goroutines Current number of goroutines
# TYPE api_gateway_goroutines gauge
api_gateway_goroutines %d

# HELP api_gateway_memory_alloc_bytes Current allocated memory in bytes
# TYPE api_gateway_memory_alloc_bytes gauge
api_gateway_memory_alloc_bytes %d

# HELP api_gateway_memory_total_alloc_bytes Total allocated memory in bytes
# TYPE api_gateway_memory_total_alloc_bytes counter
api_gateway_memory_total_alloc_bytes %d

# HELP api_gateway_memory_sys_bytes System memory in bytes
# TYPE api_gateway_memory_sys_bytes gauge
api_gateway_memory_sys_bytes %d

# HELP api_gateway_gc_cycles_total Total number of GC cycles
# TYPE api_gateway_gc_cycles_total counter
api_gateway_gc_cycles_total %d

# HELP api_gateway_last_scrape_timestamp_seconds Timestamp of last metrics scrape
# TYPE api_gateway_last_scrape_timestamp_seconds gauge
api_gateway_last_scrape_timestamp_seconds %d
`,
		runtime.NumGoroutine(),
		memStats.Alloc,
		memStats.TotalAlloc,
		memStats.Sys,
		memStats.NumGC,
		time.Now().Unix(),
	)

	if _, err := w.Write([]byte(metrics)); err != nil {
		logger.WithError(err).Error("Failed to write metrics response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Debug("Metrics served successfully")
}

// handleMetricsHealth provides health check for the metrics server
func (ms *MetricsServer) handleMetricsHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := `{"status":"healthy","service":"api-gateway-metrics","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`

	if _, err := w.Write([]byte(response)); err != nil {
		ms.logger.WithError(err).Error("Failed to write metrics health response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
