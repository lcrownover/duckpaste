package db

import (
	"fmt"
	"log/slog"
	"time"
)

// Cleaner is a process that runs in a goroutine and periodically cleans up
// expired items from the database.

type CleanerOpts struct {
	// How often to run the cleaner
	Interval time.Duration
}

func NewCleanerOpts(intervalMinutes int) *CleanerOpts {
	return &CleanerOpts{
		Interval: time.Duration(intervalMinutes) * time.Minute,
	}
}

func StartCleaner(h *CosmosHandler, opts *CleanerOpts) {
	slog.Debug("starting cleaner")
	if opts == nil {
		opts = NewCleanerOpts(60)
	}
	for {
		slog.Info("Cleaner running")
		allItems, err := h.GetAllItems()
		if err != nil {
			slog.Error("Failed to get all items", "error", err)
			goto sleep
		}
		for _, item := range allItems {
			// Check if item is expired
			fmt.Println(item)
		}
	sleep:
		slog.Info("Cleaner sleeping", "duration", opts.Interval)
		time.Sleep(opts.Interval)
	}
}
