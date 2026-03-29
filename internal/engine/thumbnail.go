package engine

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"golang.org/x/image/draw"
)

// ThumbnailCache generates and caches thumbnails
type ThumbnailCache struct {
	cacheDir string
	maxSize  int
}

func NewThumbnailCache(cacheDir string, maxSize int) *ThumbnailCache {
	if maxSize <= 0 {
		maxSize = 256
	}
	os.MkdirAll(cacheDir, 0o755)
	return &ThumbnailCache{
		cacheDir: cacheDir,
		maxSize:  maxSize,
	}
}

// GetOrCreate returns the path to a cached thumbnail, creating it if needed
func (tc *ThumbnailCache) GetOrCreate(imagePath string) (string, error) {
	// Generate cache key from path + size
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", imagePath, tc.maxSize)))
	cacheKey := fmt.Sprintf("%x.jpg", hash[:16])
	cachePath := filepath.Join(tc.cacheDir, cacheKey)

	// Return cached if exists
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	// Generate thumbnail
	if err := tc.generate(imagePath, cachePath); err != nil {
		return "", err
	}

	return cachePath, nil
}

func (tc *ThumbnailCache) generate(srcPath, dstPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	var newW, newH int
	if srcW > srcH {
		newW = tc.maxSize
		newH = int(float64(srcH) * float64(tc.maxSize) / float64(srcW))
	} else {
		newH = tc.maxSize
		newW = int(float64(srcW) * float64(tc.maxSize) / float64(srcH))
	}

	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	outFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create thumbnail: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, dst, &jpeg.Options{Quality: 80}); err != nil {
		os.Remove(dstPath)
		return fmt.Errorf("encode thumbnail: %w", err)
	}

	log.Debug().Str("src", srcPath).Str("dst", dstPath).Msg("thumbnail generated")
	return nil
}
