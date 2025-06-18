package service

import (
	"context"
	"fmt"

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

	logger.Info()

	// Initialize results
	results := &MetricsResults{}

	// TODO: Implement get metrics

	// TODO: Set results

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info()

	return results, nil
}
