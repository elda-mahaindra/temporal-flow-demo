package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"flowngine/api"
	"flowngine/service"
	"flowngine/util/config"

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

	// --- Init Temporal client ---
	temporalClient, err := createTemporalClient(config.Temporal)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error()

		os.Exit(1)
	}
	defer temporalClient.Close()

	logger.Info("🎉 Temporal client initialized - ready for workflow orchestration!")

	// --- Init service layer ---
	service := service.NewService(logger, temporalClient)

	// --- Init api layer ---
	api := api.NewApi(logger, service)

	// --- Run server(s) ---
	runGrpcServer(config.App.Port, api)

	// --- Wait for signal ---
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// --- Block until signal is received ---
	<-ch

	log.Printf("end of program...")
}
