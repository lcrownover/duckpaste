package web

import (
	"fmt"

	"github.com/lcrownover/duckpaste/internal/db"
)

var dbClient *db.CosmosHandler

func NewPasteEntryFromDbItem(item db.Item) PasteEntry {
	return PasteEntry{
		Id:              string(item.Id),
		ExpirationHours: item.LifetimeHours,
		Content:         string(item.Content),
		DeleteOnRead:    item.DeleteOnRead,
	}
}

func getPasteEntry(id string) (PasteEntry, error) {
	// get it
	pasteEntry, err := dbClient.ReadItem(db.ItemID(id))
	if err != nil {
		return PasteEntry{}, fmt.Errorf("paste not found")
	}

	return NewPasteEntryFromDbItem(*pasteEntry), nil
}

func createPasteEntry(content string, lifetimeHours int, deleteOnRead bool) (PasteEntry, error) {
	newEntry := PasteEntry{
		Content:         content,
		ExpirationHours: lifetimeHours,
		DeleteOnRead:    deleteOnRead,
	}

	if newEntry.ExpirationHours == 0 {
		newEntry.ExpirationHours = defaultLifetime
	}

	//convert
	newDbItem := dbClient.NewItem(newEntry.Content, newEntry.ExpirationHours, newEntry.DeleteOnRead)
	newEntry.Id = string(newDbItem.Id)

	// put it in the database
	err := dbClient.CreateItem(newDbItem.Id, newDbItem)
	if err != nil {
		return newEntry, err
	}

	return newEntry, nil
}
