package db

import (
	"fmt"
	"os"
)

type CosmosConfig struct {
	Endpoint         string
	Key              string
	DatabaseName     string
	ContainerName    string
	PartitionKeyPath string
}

func GetConfig() (*CosmosConfig, error) {
	// Load Cosmos Endpoint and Key from environment variables
	endpoint, found := os.LookupEnv("COSMOS_ENDPOINT")
	if !found {
		return nil, fmt.Errorf("COSMOS_ENDPOINT environment variable not set")
	}
	key, found := os.LookupEnv("COSMOS_KEY")
	if !found {
		return nil, fmt.Errorf("COSMOS_KEY environment variable not set")
	}
	databaseName, found := os.LookupEnv("COSMOS_DATABASE")
	if !found {
		return nil, fmt.Errorf("COSMOS_DATABASE environment variable not set")
	}
	containerName, found := os.LookupEnv("COSMOS_CONTAINER")
	if !found {
		return nil, fmt.Errorf("COSMOS_CONTAINER environment variable not set")
	}
	partitionKeyPath, found := os.LookupEnv("COSMOS_PARTITION_KEY_PATH")
	if !found {
		return nil, fmt.Errorf("COSMOS_PARTITION_KEY_PATH environment variable not set")
	}

	return &CosmosConfig{
		Endpoint:         endpoint,
		Key:              key,
		DatabaseName:     databaseName,
		ContainerName:    containerName,
		PartitionKeyPath: partitionKeyPath,
	}, nil
}
