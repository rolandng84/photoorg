package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) ListFiles(c *gin.Context) {
	jobID := c.Param("id")
	category := c.Query("category")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	files, total, err := h.db.GetFilesByJob(jobID, category, status, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]interface{}, 0, len(files))
	for _, f := range files {
		result = append(result, f.ToJSON())
	}

	c.JSON(http.StatusOK, gin.H{
		"files": result,
		"total": total,
		"page":  page,
	})
}

func (h *Handlers) GetCategorySummary(c *gin.Context) {
	jobID := c.Param("id")
	summary, err := h.db.GetCategorySummary(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handlers) UpdateFileCategory(c *gin.Context) {
	fileID, err := strconv.ParseInt(c.Param("fileId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	var body struct {
		Category string `json:"category" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category is required"})
		return
	}

	if err := h.db.UpdateFileCategory(fileID, body.Category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) BulkUpdateFileCategory(c *gin.Context) {
	var body struct {
		FileIDs  []int64 `json:"file_ids" binding:"required"`
		Category string  `json:"category" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_ids and category are required"})
		return
	}

	if err := h.db.BulkUpdateFileCategory(body.FileIDs, body.Category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "count": len(body.FileIDs)})
}
