package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"svc-balance/activity"
	"svc-balance/service"
	"svc-balance/store"
	"svc-balance/util/config"
	"svc-balance/worker"

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
	}).Infof("Starting '%s' Temporal worker ...", config.App.Name)

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
	balanceService := service.NewService(logger, store)

	// --- Init activity ---
	activity := activity.NewActivity(logger, balanceService)

	// --- Init temporal client ---
	temporalClient, err := createTemporalClient(logger, config.Temporal)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error("Failed to create Temporal client")

		os.Exit(1)
	}

	// --- Init worker ---
	temporalWorker, err := worker.NewWorker(
		logger,
		temporalClient,
		config.Temporal.TaskQueue,
		activity,
	)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error("Failed to create Temporal worker")

		os.Exit(1)
	}

	// --- Create context for graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Start Temporal worker following the established pattern ---
	go func() {
		if err := temporalWorker.Run(ctx); err != nil {
			logger.WithFields(logrus.Fields{
				"[op]":  op,
				"error": err.Error(),
			}).Error("Temporal worker failed")
			cancel()
		}
	}()

	// --- Wait for signal ---
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	logger.Info("Balance service Temporal worker is running. Press Ctrl+C to exit.")

	// --- Block until signal is received ---
	<-ch

	logger.Info("Shutdown signal received, stopping worker...")
	cancel()
	temporalWorker.Stop()

	log.Printf("Balance service worker stopped gracefully")
}
