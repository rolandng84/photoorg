package engine

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

const MarkerFile = ".photo_org_managed"

var validExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".bmp": true, ".heic": true, ".heif": true,
}

// FileInfo represents a discovered image file
type FileInfo struct {
	Path     string
	Filename string
	Size     int64
}

// ScanDirectory walks the directory tree and finds all image files,
// skipping directories marked with .photo_org_managed
func ScanDirectory(rootPath string) ([]FileInfo, error) {
	root := filepath.Clean(rootPath)

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}

	var files []FileInfo

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Warn().Err(err).Str("path", path).Msg("scan error, skipping")
			return filepath.SkipDir
		}

		if d.IsDir() {
			// Skip managed directories
			markerPath := filepath.Join(path, MarkerFile)
			if _, err := os.Stat(markerPath); err == nil {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if !validExtensions[ext] {
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			log.Warn().Err(err).Str("path", path).Msg("cannot stat file, skipping")
			return nil
		}

		files = append(files, FileInfo{
			Path:     path,
			Filename: d.Name(),
			Size:     fi.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	if files == nil {
		files = make([]FileInfo, 0)
	}

	return files, nil
}
