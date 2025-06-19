package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

type GetMetricsParams struct {
	Metrics string `json:"metrics"`
}

type MetricsResults struct {
	Metrics map[string]any `json:"metrics"`
}

func (service *Service) GetMetrics(ctx context.Context, params *GetMetricsParams) (*MetricsResults, error) {
	const op = "service.Service.GetMetrics"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info("Collecting system metrics")

	// Get runtime memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Initialize results with basic metrics
	results := &MetricsResults{
		Metrics: map[string]any{
			"service": map[string]any{
				"name":    "api-gateway",
				"version": "1.0.0",
				"uptime":  time.Since(time.Now().Add(-time.Hour)).String(), // Placeholder uptime
			},
			"runtime": map[string]any{
				"goroutines":   runtime.NumGoroutine(),
				"memory_alloc": memStats.Alloc,
				"memory_total": memStats.TotalAlloc,
				"memory_sys":   memStats.Sys,
				"gc_cycles":    memStats.NumGC,
			},
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info("Metrics collected successfully")

	return results, nil
}
