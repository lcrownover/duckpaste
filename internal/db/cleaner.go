package db

import (
	"time"

	"github.com/lcrownover/duckpaste/internal/message"
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

func StartCleaner(ch chan<- message.Message, h *CosmosHandler, opts *CleanerOpts) {
	ch <- message.Message{
		Status: message.Debug,
		Text:   "starting cleaner",
		Source: "StartCleaner",
	}
	if opts == nil {
		opts = NewCleanerOpts(60)
	}
	for {
		ch <- message.Message{
			Status: message.Info,
			Text:   "cleaner running",
			Source: "StartCleaner",
		}
		allItems, err := h.GetAllItems()
		if err != nil {
			ch <- message.Message{
				Status: message.Error,
				Text:   "failed to get all items: " + err.Error(),
				Source: "StartCleaner",
			}
			goto sleep
		}
		for _, item := range allItems {
			// Check if item is expired
			// itemCreatedTime, err := ParseTime(item.Created)
			if err != nil {
				ch <- message.Message{
					Status: message.Error,
					Text:   "failed to parse item time: " + err.Error(),
					Source: "StartCleaner",
				}
				goto sleep
			}
			itemExpirationTime := item.Created.Add(time.Duration(item.LifetimeHours) * time.Hour)
			if time.Now().After(itemExpirationTime) {
				ch <- message.Message{
					Status: message.Debug,
					Text:   "deleting expired item id: " + string(item.Id),
					Source: "StartCleaner",
				}

				h.DeleteItem(item.Id)
			}
		}
	sleep:
		ch <- message.Message{
			Status: message.Info,
			Text:   "cleaner sleeping",
			Source: "StartCleaner",
		}
		time.Sleep(opts.Interval)
	}
}
