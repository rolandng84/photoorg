package api

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sync"

	"photoorg/internal/config"
	"photoorg/internal/database"
	"photoorg/internal/engine"
	"photoorg/internal/sse"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handlers struct {
	db         *database.DB
	cfg        *config.Config
	broker     *sse.Broker
	thumbCache *engine.ThumbnailCache

	// Active job cancellation
	mu           sync.Mutex
	activeCancel map[string]context.CancelFunc
}

func NewRouter(
	db *database.DB,
	cfg *config.Config,
	broker *sse.Broker,
	thumbCache *engine.ThumbnailCache,
	frontendFS embed.FS,
) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(SecurityHeaders())
	r.Use(CORS(cfg.APICORSOrigins))
	r.Use(RequestLogger())

	h := &Handlers{
		db:           db,
		cfg:          cfg,
		broker:       broker,
		thumbCache:   thumbCache,
		activeCancel: make(map[string]context.CancelFunc),
	}

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Browse
		api.GET("/browse", h.Browse)

		// Settings
		api.GET("/settings", h.GetSettings)
		api.PUT("/settings", h.UpdateSettings)

		// LLM
		api.GET("/models", h.ListModels)

		// Images
		api.GET("/image", h.ServeImage)
		api.GET("/thumbnail", h.ServeThumbnail)

		// Jobs
		api.POST("/jobs", h.CreateJob)
		api.GET("/jobs", h.ListJobs)
		api.GET("/jobs/:id", h.GetJob)
		api.POST("/jobs/:id/cancel", h.CancelJob)
		api.POST("/jobs/:id/commit", h.CommitJob)
		api.POST("/jobs/:id/undo", h.UndoJob)
		api.DELETE("/jobs/:id", h.DeleteJob)

		// Files
		api.GET("/jobs/:id/files", h.ListFiles)
		api.GET("/jobs/:id/files/summary", h.GetCategorySummary)
		api.PUT("/jobs/:id/files/:fileId/category", h.UpdateFileCategory)
		api.PUT("/jobs/:id/files/bulk-category", h.BulkUpdateFileCategory)

		// SSE
		api.GET("/events", h.Events)
	}

	setupSPA(r, frontendFS)

	return r
}

func setupSPA(r *gin.Engine, frontendFS embed.FS) {
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Warn().Err(err).Msg("frontend dist not found, SPA disabled")
		return
	}

	httpFS := http.FS(distFS)

	// Serve static assets
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS(fmt.Sprintf("assets/%s", c.Param("filepath")), httpFS)
	})

	// Read index.html once at startup
	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		log.Warn().Err(err).Msg("index.html not found in dist")
		return
	}

	// SPA catch-all
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Try exact file
		if f, err := distFS.(fs.ReadFileFS).ReadFile(strings.TrimPrefix(path, "/")); err == nil {
			_ = f
			c.FileFromFS(strings.TrimPrefix(path, "/"), httpFS)
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}
