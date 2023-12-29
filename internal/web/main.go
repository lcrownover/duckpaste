package web

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultLifetime int = 48
)

type PasteEntry struct {
	Id            string `json:"id" form:"id"`
	LifetimeHours int    `json:"lifetimeHours" form:"lifetimeHours"`
	Content       string `json:"content" form:"content"`
	DeleteOnRead  bool   `json:"deleteOnRead" form:"deleteOnRead"`
}

var pastesDummy []PasteEntry

func StartServer() {
	server := gin.Default()
	server.GET("/api/paste", getPasteApi)
	server.POST("/api/paste", createPasteApi)
	server.Run()

}

func getPasteEntry(id string) (PasteEntry, error) {
	// get it
	for _, paste := range pastesDummy {
		if paste.Id == id {
			return paste, nil
		}
	}

	return PasteEntry{}, fmt.Errorf("paste not found")
}

func createPasteEntry(content string, lifetimeHours int, deleteOnRead bool) (PasteEntry, error) {
	newEntry := PasteEntry{
		Content:       content,
		LifetimeHours: lifetimeHours,
		DeleteOnRead:  deleteOnRead,
	}

	if newEntry.LifetimeHours == 0 {
		newEntry.LifetimeHours = defaultLifetime
	}

	// generate unique key
	uniqueKey := rand.Int()

	newEntry.Id = strconv.FormatInt(int64(uniqueKey), 10)

	// put it in the database
	pastesDummy = append(pastesDummy, newEntry)

	// put the key in newEntry

	return newEntry, nil
}

type errorResponse struct {
	Message string `json:"message"`
}

// ENDPOINTS

func createPasteApi(c *gin.Context) {
	var paste PasteEntry

	err := c.Bind(&paste)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			fmt.Sprintf("couldn't unmarshal payload to PasteEntry struct: %s", err),
		})
		return
	}

	if strings.TrimSpace(paste.Content) == "" {
		c.JSON(http.StatusBadRequest, errorResponse{
			"please provide actual content",
		})
		return
	}

	paste, err = createPasteEntry(paste.Content, paste.LifetimeHours, paste.DeleteOnRead)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			fmt.Sprintf("failed to create paste entry: %s", err),
		})
		return
	}

	c.JSON(http.StatusCreated, paste)

}

func getPasteApi(c *gin.Context) {
	pasteId := c.Query("id")
	if pasteId == "" {
		c.JSON(http.StatusNotFound, errorResponse{
			"could not retrieve \"id\" from request",
		})
		return
	}

	paste, err := getPasteEntry(pasteId)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			fmt.Sprintf("no paste found with id: %s", pasteId),
		})
		return
	}

	c.JSON(http.StatusOK, paste)

}
