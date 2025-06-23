package main

import (
	"flowngine/util/config"
	"fmt"

	"go.temporal.io/sdk/client"
)

func createTemporalClient(config config.Temporal) (client.Client, error) {
	temporalClient, err := client.Dial(client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s grpc server: %w", config.HostPort, err)
	}

	return temporalClient, nil
}
