package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/lcrownover/duckpaste/internal/db"
	"github.com/lcrownover/duckpaste/internal/web"
)

var debug bool

func main() {
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()

	// Logging verbosity
	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	cosmosConfig, err := db.GetConfig()
	if err != nil {
		slog.Error("Failed to get Cosmos Config", "error", err)
		os.Exit(1)
	}
	client, err := db.GetCostmosClient(cosmosConfig)
	if err != nil {
		slog.Error("Failed to create Azure Cosmos DB client", "error", err)
	}

	// Create the database
	_, err = db.CreateDatabase(client, "duckpaste")
	if err != nil {
		slog.Error("Failed to create database", "error", err)
		os.Exit(1)
	}

	// Create the container
	err = db.CreateContainer(client, "duckpaste", "duckpaste", "/id")
	if err != nil {
		slog.Error("Failed to create container", "error", err)
		os.Exit(1)
	}

	// Test payload
	itemId := db.GetRandomID()
	content := "pretend-im-a-paste"
	item := &db.Item{
		Id:            itemId,
		LifetimeHours: 24,
		Content:       db.EncodeContent(content),
		DeleteOnRead:  false,
		Created:       db.GetCurrentTime(),
	}

	// Try to create an item
	err = db.CreateItem(client, "duckpaste", "duckpaste", item.Id, item)
	if err != nil {
		slog.Error("Failed to create item", "error", err)
		os.Exit(1)
	}

	// Try to read an item
	item, err = db.ReadItem(client, "duckpaste", "duckpaste", item.Id)
	if err != nil {
		slog.Error("Failed to read item", "error", err)
		os.Exit(1)
	}
	fmt.Printf("Item: %+v\n", item)

	// Try to delete an item
	err = db.DeleteItem(client, "duckpaste", "duckpaste", item.Id)
	if err != nil {
		slog.Error("Failed to delete item", "error", err)
		os.Exit(1)
	}

	web.StartServer()
}
