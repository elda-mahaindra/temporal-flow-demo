package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type CheckHealthParams struct {
}

type CheckHealthResults struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

func (service *Service) CheckHealth(ctx context.Context, params *CheckHealthParams) (results *CheckHealthResults, err error) {
	const op = "service.Service.CheckHealth"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Initialize results
	results = &CheckHealthResults{}

	// TODO: Implement health check

	// TODO: Set results

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info()

	return
}
