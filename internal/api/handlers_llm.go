package api

import (
	"net/http"

	"photoorg/internal/llm"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) ListModels(c *gin.Context) {
	provider := c.Query("provider")
	endpoint := c.Query("endpoint")
	apiKey := c.Query("api_key")

	var p llm.VisionProvider

	switch provider {
	case "ollama":
		p = llm.NewOllamaProvider(endpoint)
	case "openai-compatible":
		p = llm.NewOpenAICompatibleProvider(endpoint, apiKey)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}

	models, err := p.ListModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, make([]string, 0))
		return
	}

	c.JSON(http.StatusOK, models)
}
