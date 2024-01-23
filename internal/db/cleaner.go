package db

import (
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
	slog.Debug("starting cleaner", "source", "StartCleaner")
	if opts == nil {
		opts = NewCleanerOpts(60)
	}
	for {
		slog.Info("cleaner running", "source", "StartCleaner")
		allItems, err := h.GetAllItems()
		if err != nil {
			slog.Error("failed to get all items: "+err.Error(), "source", "StartCleaner")
			goto sleep
		}
		for _, item := range allItems {
			// Check if item is expired
			// itemCreatedTime, err := ParseTime(item.Created)
			if err != nil {
				slog.Error("failed to parse item time: "+err.Error(), "source", "StartCleaner")
				goto sleep
			}
			itemExpirationTime := item.Created.Add(time.Duration(item.LifetimeHours) * time.Hour)
			if time.Now().After(itemExpirationTime) {
				slog.Info("deleting expired item", "id", string(item.Id), "source", "StartCleaner")
				h.DeleteItem(item.Id)
			}
		}
	sleep:
		slog.Info("cleaner sleeping", "source", "StartCleaner")
		time.Sleep(opts.Interval)
	}
}
