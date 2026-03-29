package api

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".bmp": true, ".heic": true, ".heif": true,
}

type BrowseItem struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	IsDir      bool   `json:"is_dir"`
	IsDrive    bool   `json:"is_drive,omitempty"`
	ImageCount *int   `json:"image_count"` // null = not yet counted
}

func (h *Handlers) Browse(c *gin.Context) {
	startPath := c.Query("path")

	items := make([]BrowseItem, 0)

	var dirPath string
	if startPath == "" || startPath == "/" {
		home, _ := os.UserHomeDir()
		dirPath = home
	} else {
		dirPath = startPath
	}

	info, err := os.Stat(dirPath)
	if err != nil || !info.IsDir() {
		home, _ := os.UserHomeDir()
		dirPath = home
	}

	// Add Windows drives at root level
	if runtime.GOOS == "windows" && (dirPath == filepath.VolumeName(dirPath)+`\` || startPath == "") {
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			dp := string(drive) + `:\`
			if _, err := os.Stat(dp); err == nil {
				items = append(items, BrowseItem{
					Name:    dp,
					Path:    dp,
					IsDir:   true,
					IsDrive: true,
				})
			}
		}
	}

	// Add parent directory
	parent := filepath.Dir(dirPath)
	if parent != dirPath {
		items = append(items, BrowseItem{
			Name:  "..",
			Path:  parent,
			IsDir: true,
		})
	}

	// List directory entries
	entries, err := os.ReadDir(dirPath)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			fullPath := filepath.Join(dirPath, entry.Name())
			// Count images in this directory (non-recursive, fast)
			count := countImagesShallow(fullPath)
			items = append(items, BrowseItem{
				Name:       entry.Name(),
				Path:       fullPath,
				IsDir:      true,
				ImageCount: &count,
			})
		}
	}

	c.JSON(http.StatusOK, items)
}

// countImagesShallow counts image files in the immediate directory only (not recursive)
func countImagesShallow(dirPath string) int {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if imageExtensions[ext] {
			count++
		}
	}
	return count
}
