package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/lcrownover/duckpaste/internal/db"
	"github.com/lcrownover/duckpaste/internal/message"
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

	cosmosHandler, err := db.NewCosmosHandler(cosmosConfig)
	if err != nil {
		slog.Error("Failed to create Cosmos Handler", "error", err)
		os.Exit(1)
	}
	cosmosHandler.Init()

	// Test payload
	content := "pretend-im-a-paste"
	item := cosmosHandler.NewItem(content, 24, false)

	// Try to create an item
	err = cosmosHandler.CreateItem(item.Id, item)
	if err != nil {
		slog.Error("Failed to create item", "error", err)
		os.Exit(1)
	}

	// Try to read an item
	item, err = cosmosHandler.ReadItem(item.Id)
	if err != nil {
		slog.Error("Failed to read item", "error", err)
		os.Exit(1)
	}
	fmt.Printf("Item: %+v\n", item)

	// Try to delete an item
	err = cosmosHandler.DeleteItem(item.Id)
	if err != nil {
		slog.Error("Failed to delete item", "error", err)
		os.Exit(1)
	}

	messagesCh := make(chan message.Message, 100)

	go db.StartCleaner(messagesCh, cosmosHandler, db.NewCleanerOpts(1))

	go web.StartServer(messagesCh)

	for msg := range messagesCh {
		switch {
		case msg.Status == message.Info:
			slog.Info(msg.Text, "source", msg.Source)
		case msg.Status == message.Warning:
			slog.Warn(msg.Text, "source", msg.Source)
		case msg.Status == message.Error:
			slog.Error(msg.Text, "source", msg.Source)
		}
	}
}
