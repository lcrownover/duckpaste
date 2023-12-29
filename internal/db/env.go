package db

import (
	"fmt"
	"os"
)

type CosmosConfig struct {
	Endpoint string
	Key      string
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

	return &CosmosConfig{
		Endpoint: endpoint,
		Key:      key,
	}, nil
}
