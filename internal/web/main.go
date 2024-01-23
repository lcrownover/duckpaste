package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

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
	Id              string    `json:"id"`
	ExpirationHours int       `json:"expirationHours" form:"pasteExpirationHours"`
	Content         string    `json:"content" form:"pasteContent"`
	Password        string    `json:"password" form:"pastePassword"`
	DeleteOnRead    bool      `json:"deleteOnRead" form:"pasteDeleteOnRead"`
	Created         time.Time `json:"created"`
}

type WebConfig struct {
	Host string
	Port string
}

func (wc *WebConfig) Address() string {
	var host string
	if wc.Host == "localhost" {
		host = ""
	} else {
		host = wc.Host
	}
	return fmt.Sprintf("%s:%s", host, wc.Port)
}

type WebHandler struct {
	config *WebConfig
	server *gin.Engine
}

func (h *WebHandler) Run() error {
	return h.server.Run(h.config.Address())
}

func NewWebHandler(c *WebConfig, server *gin.Engine) *WebHandler {
	h := &WebHandler{config: c, server: server}
	pattern := "templates/*html"
	LoadHTMLFromEmbedFS(server, templatesFS, pattern)
	server.GET("/api/paste", h.getPasteApi)
	server.POST("/api/paste", h.createPasteApi)
	server.GET("/", h.getRoot)
	server.GET("/:pasteId", h.getPaste)
	server.GET("/about", h.getAbout)

	// drill down into static FS
	staticFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		slog.Error("failed to get sub FS: "+err.Error(), "source", "StartServer")
	}

	// serve embedded static files
	server.StaticFS("/static", http.FS(staticFS))

	return h
}

func GetWebConfig() *WebConfig {
	host, found := os.LookupEnv("SERVER_HOST")
	if !found {
		host = "localhost"
	}
	port, found := os.LookupEnv("SERVER_PORT")
	if !found {
		port = "8080"
	}
	return &WebConfig{
		Host: host,
		Port: port,
	}
}

func StartServer() {
	dbConfig, err := db.GetDBConfig()
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

	// get listen config from env
	wc := GetWebConfig()

	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()
	webHandler := NewWebHandler(wc, server)

	err = webHandler.Run()
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

// ENDPOINTS

func (h *WebHandler) createPasteApi(c *gin.Context) {
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

	paste, err = createPasteEntry(paste)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			fmt.Sprintf("failed to create paste entry: %s", err),
		})
		return
	}

	pasteUrl := fmt.Sprintf("/%s", paste.Id)
	c.Redirect(http.StatusFound, pasteUrl)
}

func (h *WebHandler) getPasteApi(c *gin.Context) {
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

	// If the paste is set to delete on read and it's been longer than 10sec since it was created, delete it
	if paste.DeleteOnRead && paste.Created.Add(time.Second*10).Before(time.Now()) {
		err = deletePasteEntry(paste)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				fmt.Sprintf("failed to delete paste: %s", err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, paste)

}

func (h *WebHandler) getRoot(c *gin.Context) {
	c.HTML(http.StatusOK, "templates/index.html", nil)
}

func (h *WebHandler) getPaste(c *gin.Context) {
	pasteID := c.Param("pasteId")
	paste, err := getPasteEntry(pasteID)
	if err != nil {
		c.HTML(http.StatusNotFound, "templates/notfound.html", nil)
		return
	}
	// If the paste is set to delete on read and it's been longer than 10sec since it was created, delete it
	if paste.DeleteOnRead && paste.Created.Add(time.Second*10).Before(time.Now()) {
		err = deletePasteEntry(paste)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				fmt.Sprintf("failed to delete paste: %s", err),
			})
			return
		}
	}
	decodedContent, err := db.DecodeContent(db.ItemContent(paste.Content))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			"failed to decode content of paste",
		})
	}
	// TODO(lcrown): fix https or http
	pasteURL := fmt.Sprintf("http://%s/%s", h.config.Address(), paste.Id)
	c.HTML(http.StatusOK, "templates/paste.html", gin.H{
		"pasteURL":     pasteURL,
		"pasteContent": decodedContent,
	})
}

func (h *WebHandler) getAbout(c *gin.Context) {
	c.HTML(http.StatusOK, "templates/about.html", nil)
}
