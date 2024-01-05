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

	"github.com/gin-gonic/gin"
	"github.com/lcrownover/duckpaste/internal/db"
)

const (
	defaultLifetime int    = 48
	containerName   string = "duckpaste"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

type PasteEntry struct {
	Id            string `json:"id" form:"id"`
	LifetimeHours int    `json:"lifetimeHours" form:"lifetimeHours"`
	Content       string `json:"content" form:"content"`
	DeleteOnRead  bool   `json:"deleteOnRead" form:"deleteOnRead"`
}

// func (p *PasteEntry) toDbItem() db.Item {
// 	return db.Item{
// 		Id:            db.ItemID(p.Id),
// 		LifetimeHours: p.LifetimeHours,
// 		Content:       db.ItemContent(p.Content),
// 		DeleteOnRead:  p.DeleteOnRead,
// 		Created:       db.GetCurrentTime(),
// 	}
// }

func NewPasteEntryFromDbItem(item db.Item) PasteEntry {
	return PasteEntry{
		Id:            string(item.Id),
		LifetimeHours: item.LifetimeHours,
		Content:       string(item.Content),
		DeleteOnRead:  item.DeleteOnRead,
	}
}

// var pastesDummy []PasteEntry
var dbClient *db.CosmosHandler

func StartServer() {
	dbConfig, err := db.GetConfig()
	if err != nil {
		slog.Error("failed to get Cosmos Config: "+err.Error(), "source", "StartServer")
	}
	dbClient, err = db.NewCosmosHandler(dbConfig)
	if err != nil {
		slog.Error("failed to create Cosmos Handler: "+err.Error(), "source", "StartServer")
	}
	err = dbClient.Init()
	if err != nil {
		slog.Error("failed to initialize Cosmos Handler: "+err.Error(), "source", "StartServer")
	}
	server := gin.Default()
	pattern := "templates/*html"
	LoadHTMLFromEmbedFS(server, templatesFS, pattern)
	server.GET("/api/paste", getPasteApi)
	server.POST("/api/paste", createPasteApi)
	server.GET("/", pasteFrontEnd)

	// drill down into static FS
	staticFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		slog.Error("failed to get sub FS: "+err.Error(), "source", "StartServer")
	}

	// serve embedded static files
	server.StaticFS("/static", http.FS(staticFS))

	err = server.Run()
	if err != nil {
		slog.Error("failed to start server: "+err.Error(), "source", "StartServer")
	}

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
	pasteEntry, err := dbClient.ReadItem(db.ItemID(id))
	if err != nil {
		return PasteEntry{}, fmt.Errorf("paste not found")
	}

	return NewPasteEntryFromDbItem(*pasteEntry), nil
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

	//convert
	newDbItem := dbClient.NewItem(newEntry.Content, newEntry.LifetimeHours, newEntry.DeleteOnRead)
	newEntry.Id = string(newDbItem.Id)

	// put it in the database
	err := dbClient.CreateItem(newDbItem.Id, newDbItem)
	if err != nil {
		return newEntry, err
	}

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
	c.HTML(http.StatusOK, "templates/index.html", nil)
}
