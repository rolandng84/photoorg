package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"photoorg/internal/database"
	"photoorg/internal/engine"
	"photoorg/internal/llm"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (h *Handlers) ListJobs(c *gin.Context) {
	jobs, err := h.db.ListJobs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]interface{}, 0, len(jobs))
	for _, j := range jobs {
		result = append(result, j.ToJSON())
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handlers) GetJob(c *gin.Context) {
	id := c.Param("id")
	job, err := h.db.GetJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, job.ToJSON())
}

func (h *Handlers) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	job, err := h.db.GetJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	if job.Status == "committed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a committed job, undo first"})
		return
	}

	// Cancel if still running
	h.mu.Lock()
	if cancel, ok := h.activeCancel[id]; ok {
		cancel()
		delete(h.activeCancel, id)
	}
	h.mu.Unlock()

	if err := h.db.DeleteJob(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

type createJobRequest struct {
	InputPath    string   `json:"input_path"`
	Provider     string   `json:"provider"`
	Model        string   `json:"model"`
	Endpoint     string   `json:"endpoint"`
	APIKey       string   `json:"api_key"`
	Mode         string   `json:"mode"`
	Categories   []string `json:"categories"`
	Concurrency  int      `json:"concurrency"`
	CustomPrompt string   `json:"custom_prompt"`
	InstantMove  bool     `json:"instant_move"`
}

func (h *Handlers) CreateJob(c *gin.Context) {
	var req createJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.InputPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input_path is required"})
		return
	}
	if len(req.Categories) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "categories are required"})
		return
	}
	if req.Provider == "" {
		req.Provider = "ollama"
	}
	if req.Model == "" {
		req.Model = "llava:7b"
	}
	if req.Mode == "" {
		req.Mode = "move"
	}
	if req.Concurrency < 1 {
		req.Concurrency = 4
	}

	// Scan directory for images
	scannedFiles, err := engine.ScanDirectory(req.InputPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to scan directory: " + err.Error()})
		return
	}
	if len(scannedFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no image files found in directory"})
		return
	}

	// Serialize categories
	catsJSON, _ := json.Marshal(req.Categories)

	now := time.Now().UnixMilli()
	jobID := uuid.New().String()

	job := &database.Job{
		ID:           jobID,
		InputPath:    req.InputPath,
		Status:       "categorizing",
		Mode:         req.Mode,
		Categories:   string(catsJSON),
		Provider:     req.Provider,
		Model:        req.Model,
		Endpoint:     req.Endpoint,
		Concurrency:  req.Concurrency,
		CustomPrompt: req.CustomPrompt,
		InstantMove:  req.InstantMove,
		TotalFiles:   len(scannedFiles),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.db.CreateJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job: " + err.Error()})
		return
	}

	// Insert file records
	dbFiles := make([]database.File, 0, len(scannedFiles))
	for _, sf := range scannedFiles {
		dbFiles = append(dbFiles, database.File{
			JobID:        jobID,
			OriginalPath: sf.Path,
			Filename:     sf.Filename,
			FileSize:     sf.Size,
		})
	}
	if err := h.db.InsertFiles(dbFiles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert files: " + err.Error()})
		return
	}

	// Create LLM provider
	var provider llm.VisionProvider
	switch req.Provider {
	case "openai":
		endpoint := req.Endpoint
		if endpoint == "" {
			endpoint = "https://api.openai.com/v1"
		}
		provider = llm.NewOpenAICompatibleProvider(endpoint, req.APIKey)
	default: // ollama
		endpoint := req.Endpoint
		if endpoint == "" {
			endpoint = "http://localhost:11434"
		}
		provider = llm.NewOllamaProvider(endpoint)
	}

	// Create cancellable context and start categorization
	ctx, cancel := context.WithCancel(context.Background())
	h.mu.Lock()
	h.activeCancel[jobID] = cancel
	h.mu.Unlock()

	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.activeCancel, jobID)
			h.mu.Unlock()
		}()
		engine.Categorize(ctx, job, h.db, provider, h.broker, req.Concurrency)
	}()

	log.Info().Str("job", jobID).Int("files", len(scannedFiles)).Msg("job created, categorization started")
	c.JSON(http.StatusCreated, job.ToJSON())
}

func (h *Handlers) CancelJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.db.GetJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	if job.Status != "categorizing" && job.Status != "committing" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job is not actively processing"})
		return
	}

	h.mu.Lock()
	cancel, ok := h.activeCancel[id]
	h.mu.Unlock()

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active categorization to cancel"})
		return
	}

	cancel()
	log.Info().Str("job", id).Msg("job cancellation requested")
	c.JSON(http.StatusOK, gin.H{"status": "cancelling"})
}

func (h *Handlers) CommitJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.db.GetJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	if job.Status != "reviewing" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job must be in reviewing status to commit"})
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.mu.Lock()
	h.activeCancel[id] = cancel
	h.mu.Unlock()

	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.activeCancel, id)
			h.mu.Unlock()
			cancel()
		}()
		if err := engine.Commit(ctx, job, h.db, h.broker); err != nil {
			log.Error().Err(err).Str("job", id).Msg("commit failed")
			// Only set failed if Commit() didn't already handle it
			currentJob, _ := h.db.GetJob(id)
			if currentJob != nil && currentJob.Status == "committing" {
				h.db.UpdateJobStatus(id, "failed")
				h.broker.PublishJSON("commit_failed", map[string]interface{}{
					"job_id": id,
					"error":  err.Error(),
				})
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "committing"})
}

func (h *Handlers) UndoJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.db.GetJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	if job.Status != "committed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can only undo committed jobs"})
		return
	}

	ctx := context.Background()
	if err := engine.Undo(ctx, job, h.db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "undo failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "undone"})
}
