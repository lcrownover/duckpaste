package db

import "time"

// Cleaner is a process that runs in a goroutine and periodically cleans up
// expired items from the database.

type CleanerOpts struct {
	// How often to run the cleaner
	Interval time.Duration
}

func NewCleanerOpts() *CleanerOpts {
	return &CleanerOpts{
		Interval: 1 * time.Hour,
	}
}

func StartCleaner(opts *CleanerOpts) {
	for {
		// Get all items
		// Filter out expired items
		// Delete expired items
		// Sleep
		time.Sleep(opts.Interval)
	}
}
