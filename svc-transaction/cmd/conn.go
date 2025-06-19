package main

import (
	"context"
	"fmt"

	"svc-transaction/util/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
)

func createPostgresPool(
	logger *logrus.Logger,
	postgresConfig config.PostgresConfig,
) (*pgxpool.Pool, error) {
	const op = "main.createPostgresPool"

	pgConfig, err := pgxpool.ParseConfig(postgresConfig.ConnectionString)
	if err != nil {
		err = fmt.Errorf("failed to parse postgres config: %w", err)

		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error()

		return nil, err
	}

	pgConfig.MaxConns = int32(postgresConfig.Pool.MaxConns)
	pgConfig.MinConns = int32(postgresConfig.Pool.MinConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error()

		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error()

		return nil, err
	}

	logger.Info("postgres pool created successfully")

	return pool, nil
}

func createTemporalClient(
	logger *logrus.Logger,
	temporalConfig config.Temporal,
) (client.Client, error) {
	const op = "main.createTemporalClient"

	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  temporalConfig.HostPort,
		Namespace: temporalConfig.Namespace,
	})
	if err != nil {
		err = fmt.Errorf("failed to create Temporal client: %w", err)

		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error()

		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"host_port": temporalConfig.HostPort,
		"namespace": temporalConfig.Namespace,
	}).Info("Temporal client created successfully")

	return temporalClient, nil
}
