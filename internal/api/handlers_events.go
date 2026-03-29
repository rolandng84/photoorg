package api

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) Events(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch := h.broker.Subscribe()
	defer h.broker.Unsubscribe(ch)

	// Send initial comment to force header flush so EventSource.onopen fires immediately
	fmt.Fprintf(c.Writer, ": connected\n\n")
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-ch:
			if !ok {
				return false
			}
			data, _ := json.Marshal(event.Payload)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			return true
		case <-time.After(30 * time.Second):
			fmt.Fprintf(w, ": ping\n\n")
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
