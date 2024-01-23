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
		Password:        string(item.Password),
		DeleteOnRead:    item.DeleteOnRead,
		Created:         item.Created,
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

func createPasteEntry(p PasteEntry) (PasteEntry, error) {
	if p.ExpirationHours == 0 {
		p.ExpirationHours = defaultLifetime
	}

	//convert
	newDbItem := dbClient.NewItem(p.Content, p.ExpirationHours, p.Password, p.DeleteOnRead)
	p.Id = string(newDbItem.Id)

	// put it in the database
	err := dbClient.CreateItem(newDbItem.Id, newDbItem)
	if err != nil {
		return p, err
	}

	return p, nil
}

func deletePasteEntry(p PasteEntry) error {
	// delete it
	err := dbClient.DeleteItem(db.ItemID(p.Id))
	if err != nil {
		return err
	}

	return nil
}
