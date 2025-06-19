package main

import (
	"flowngine/util/config"
	"fmt"

	"go.temporal.io/sdk/client"
)

func createTemporalClient(config config.Temporal) (client.Client, error) {
	hostPort := fmt.Sprintf("%s:%d", config.Host, config.Port)

	temporalClient, err := client.Dial(client.Options{
		HostPort: hostPort,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s grpc server: %w", hostPort, err)
	}

	return temporalClient, nil
}
