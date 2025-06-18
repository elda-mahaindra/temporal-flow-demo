package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"svc-transaction/api"
	"svc-transaction/service"
	"svc-transaction/store"
	"svc-transaction/util/config"

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
	service := service.NewService(logger, store)

	// --- Init api layer ---
	api := api.NewApi(logger, service)

	// --- Run servers ---
	runWorkerServer(api)

	// --- Wait for signal ---
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// --- Block until signal is received ---
	<-ch

	log.Printf("end of program...")
}
