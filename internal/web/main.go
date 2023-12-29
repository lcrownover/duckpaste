package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/gin-gonic/gin"
	"github.com/lcrownover/duckpaste/internal/db"
)

const (
	defaultLifetime int    = 48
	containerName   string = "duckpaste"
)

//go:embed templates
var templatesFS embed.FS

type PasteEntry struct {
	Id            string `json:"id" form:"id"`
	LifetimeHours int    `json:"lifetimeHours" form:"lifetimeHours"`
	Content       string `json:"content" form:"content"`
	DeleteOnRead  bool   `json:"deleteOnRead" form:"deleteOnRead"`
}

var pastesDummy []PasteEntry
var dbClient *azcosmos.Client

func StartServer() {
	dbConfig, err := db.GetConfig()
	if err != nil {
		slog.Error(err.Error())
	}
	dbClient, err = db.GetCostmosClient(dbConfig)
	if err != nil {
		slog.Error(err.Error())
	}
	server := gin.Default()
	pattern := "templates/*html"
	LoadHTMLFromEmbedFS(server, templatesFS, pattern)
	server.GET("/api/paste", getPasteApi)
	server.POST("/api/paste", createPasteApi)
	server.GET("/paste", pasteFrontEnd)
	server.Run()

}

func LoadHTMLFromEmbedFS(engine *gin.Engine, embedFS embed.FS, pattern string) {
	root := template.New("")
	tmpl := template.Must(root, LoadAndAddToRoot(engine.FuncMap, root, embedFS, pattern))
	engine.SetHTMLTemplate(tmpl)
}

func LoadAndAddToRoot(funcMap template.FuncMap, rootTemplate *template.Template, embedFS embed.FS, pattern string) error {
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	err := fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if matched, _ := regexp.MatchString(pattern, path); !d.IsDir() && matched {
			data, readErr := embedFS.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			t := rootTemplate.New(path).Funcs(funcMap)
			if _, parseErr := t.Parse(string(data)); parseErr != nil {
				return parseErr
			}
		}
		return nil
	})
	return err
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
	newEntry.Id = string(db.GetRandomID())

	// put it in the database
	err := db.CreateItem(dbClient, containerName, containerName, db.ItemID(newEntry.Id), &db.Item{})
	if err != nil {
		return newEntry, err
	}
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

func pasteFrontEnd(c *gin.Context) {
	c.HTML(http.StatusOK, "templates/paste.html", nil)
}
