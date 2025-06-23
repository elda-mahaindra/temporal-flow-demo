package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"svc-transaction/activity"
	"svc-transaction/api"
	"svc-transaction/service"
	"svc-transaction/store"
	"svc-transaction/util/config"
	"svc-transaction/worker"

	"github.com/sirupsen/logrus"
)

func start() {
	const op = "main.start"

	// --- Init logger ---
	var logger = logrus.New()
	logger.Formatter = new(logrus.JSONFormatter)
	logger.Formatter = new(logrus.TextFormatter)
	logger.Formatter.(*logrus.TextFormatter).DisableColors = true
	logger.Formatter.(*logrus.TextFormatter).DisableTimestamp = true
	logger.Level = logrus.DebugLevel
	logger.Out = os.Stdout

	// --- Load config ---
	config, err := config.LoadConfig(".")
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"scope": "LoadConfig",
			"err":   err.Error(),
		}).Error()

		os.Exit(1)
	}

	logger.WithFields(logrus.Fields{
		"[op]":   op,
		"config": fmt.Sprintf("%+v", config),
	}).Infof("Starting '%s' service ...", config.App.Name)

	// --- Init postgres pool ---
	postgresPool, err := createPostgresPool(logger, config.DB.Postgres)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error()

		os.Exit(1)
	}

	// --- Init store layer ---
	store := store.NewStore(logger, postgresPool)

	// --- Init service layer ---
	transactionService := service.NewService(logger, store)

	// --- Init activity ---
	activity := activity.NewActivity(logger, transactionService)

	// --- Create context for graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Init metrics server for Prometheus ---
	metricsServer := NewMetricsServer(logger, 8080)
	go func() {
		if err := metricsServer.Start(ctx); err != nil {
			logger.WithFields(logrus.Fields{
				"[op]":  op,
				"error": err.Error(),
			}).Error("Metrics server failed")
			cancel()
		}
	}()

	// --- Init api layer ---
	restApi := api.NewApi(logger, transactionService)

	// --- Start REST server in a goroutine ---
	go func() {
		runRestServer(config.App.Port, restApi)
	}()

	// --- Init temporal client and worker in a separate goroutine ---
	go func() {
		logger.Info("Attempting to connect to Temporal...")

		// Retry connecting to Temporal
		for {
			temporalClient, err := createTemporalClient(logger, config.Temporal)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"[op]":  op,
					"error": err.Error(),
				}).Warn("Failed to create Temporal client, retrying in 5 seconds...")

				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			// --- Init worker ---
			temporalWorker, err := worker.NewWorker(
				logger,
				temporalClient,
				config.Temporal.TaskQueue,
				activity,
				config.Temporal,
			)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"[op]":  op,
					"error": err.Error(),
				}).Error("Failed to create Temporal worker")
				temporalClient.Close()

				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			logger.Info("Temporal worker connected successfully")

			// --- Start Temporal worker ---
			if err := temporalWorker.Run(ctx); err != nil {
				logger.WithFields(logrus.Fields{
					"[op]":  op,
					"error": err.Error(),
				}).Error("Temporal worker failed")
				temporalWorker.Stop()
				temporalClient.Close()
			}

			return
		}
	}()

	// --- Wait for signal ---
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	logger.Info("Transaction service is running. Press Ctrl+C to exit.")

	// --- Block until signal is received ---
	<-ch

	logger.Info("Shutdown signal received, stopping service...")
	cancel()

	log.Printf("Transaction service stopped gracefully")
}
