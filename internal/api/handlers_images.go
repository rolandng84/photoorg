package api

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) ServeImage(c *gin.Context) {
	imgPath := c.Query("path")
	jobID := c.Query("job_id")

	if imgPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	// Clean the path to prevent traversal
	cleanPath := filepath.Clean(imgPath)

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if !imageExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not an image file"})
		return
	}

	// If job_id is provided, validate path is within job's input directory
	if jobID != "" {
		job, err := h.db.GetJob(jobID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}

		jobDir := filepath.Clean(job.InputPath)
		if !strings.HasPrefix(cleanPath, jobDir) {
			c.JSON(http.StatusForbidden, gin.H{"error": "path outside job directory"})
			return
		}
	}

	// Determine content type
	contentType := "image/jpeg"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".webp":
		contentType = "image/webp"
	case ".bmp":
		contentType = "image/bmp"
	case ".heic", ".heif":
		contentType = "image/heic"
	}

	c.Header("Cache-Control", "private, max-age=3600")
	c.File(cleanPath)
	c.Header("Content-Type", contentType)
}

func (h *Handlers) ServeThumbnail(c *gin.Context) {
	imgPath := c.Query("path")
	jobID := c.Query("job_id")
	sizeStr := c.DefaultQuery("size", "256")

	if imgPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 32 || size > 512 {
		size = 256
	}

	cleanPath := filepath.Clean(imgPath)

	ext := strings.ToLower(filepath.Ext(cleanPath))
	if !imageExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not an image file"})
		return
	}

	if jobID != "" {
		job, err := h.db.GetJob(jobID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		jobDir := filepath.Clean(job.InputPath)
		if !strings.HasPrefix(cleanPath, jobDir) {
			c.JSON(http.StatusForbidden, gin.H{"error": "path outside job directory"})
			return
		}
	}

	_ = size // ThumbnailCache uses its own configured max size
	thumbPath, err := h.thumbCache.GetOrCreate(cleanPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate thumbnail"})
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.File(thumbPath)
}
