package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"photoorg/internal/database"
	"photoorg/internal/sse"

	"github.com/rs/zerolog/log"
)

// MoveOneFile moves or copies a single file into its category subdirectory.
// Returns the new file path on success.
func MoveOneFile(file database.File, category string, mode string) (string, error) {
	targetDir := filepath.Join(filepath.Dir(file.OriginalPath), category)

	// Create target directory
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("create category dir %s: %w", targetDir, err)
	}

	// Mark as managed
	markerPath := filepath.Join(targetDir, MarkerFile)
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		os.WriteFile(markerPath, nil, 0o644)
	}

	// Determine target path with collision handling
	targetPath := filepath.Join(targetDir, file.Filename)
	counter := 1
	ext := filepath.Ext(file.Filename)
	base := file.Filename[:len(file.Filename)-len(ext)]
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
		targetPath = filepath.Join(targetDir, fmt.Sprintf("%s_%d%s", base, counter, ext))
		counter++
	}

	// Move or copy
	if mode == "copy" {
		data, err := os.ReadFile(file.OriginalPath)
		if err != nil {
			return "", fmt.Errorf("read file for copy: %w", err)
		}
		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			return "", fmt.Errorf("write copy: %w", err)
		}
	} else {
		if err := os.Rename(file.OriginalPath, targetPath); err != nil {
			return "", fmt.Errorf("move file: %w", err)
		}
	}

	return targetPath, nil
}

// Commit executes the actual file moves/copies for a reviewed job
func Commit(ctx context.Context, job *database.Job, db *database.DB, broker *sse.Broker) error {
	files, err := db.GetCategorizedFiles(job.ID)
	if err != nil {
		return fmt.Errorf("get categorized files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no categorized files to commit")
	}

	db.UpdateJobStatus(job.ID, "committing")

	var successCount int
	var errorCount int

	for _, f := range files {
		select {
		case <-ctx.Done():
			db.UpdateJobStatus(job.ID, "reviewing")
			return fmt.Errorf("commit cancelled")
		default:
		}

		newPath, err := MoveOneFile(f, f.FinalCategory, job.Mode)
		if err != nil {
			log.Error().Err(err).Str("file", f.Filename).Msg("commit file failed")
			errorCount++
			continue
		}

		// Update DB
		db.SetFileCommitted(f.ID, newPath)
		db.IncrementJobCommitted(job.ID)
		successCount++

		broker.PublishJSON("commit_progress", map[string]interface{}{
			"job_id":    job.ID,
			"file_id":   f.ID,
			"filename":  f.Filename,
			"committed": successCount,
			"total":     len(files),
		})
	}

	// Update error count on the job
	if errorCount > 0 {
		db.SetJobErrorCount(job.ID, errorCount)
	}

	// Determine final status
	if successCount == 0 {
		db.UpdateJobStatus(job.ID, "failed")
		broker.PublishJSON("commit_failed", map[string]interface{}{
			"job_id": job.ID,
			"error":  fmt.Sprintf("all %d files failed to commit", len(files)),
		})
		return fmt.Errorf("all %d files failed to commit", len(files))
	}

	db.UpdateJobStatus(job.ID, "committed")

	broker.PublishJSON("commit_completed", map[string]interface{}{
		"job_id":    job.ID,
		"committed": successCount,
		"failed":    errorCount,
		"total":     len(files),
	})

	log.Info().
		Str("job", job.ID).
		Int("committed", successCount).
		Int("failed", errorCount).
		Int("total", len(files)).
		Msg("commit complete")

	return nil
}

// Undo reverses a committed job by moving files back to their original locations
func Undo(ctx context.Context, job *database.Job, db *database.DB) error {
	files, err := db.GetCommittedFiles(job.ID)
	if err != nil {
		return fmt.Errorf("get committed files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no committed files to undo")
	}

	for _, f := range files {
		select {
		case <-ctx.Done():
			return fmt.Errorf("undo cancelled")
		default:
		}

		if f.NewPath == "" {
			continue
		}

		if job.Mode == "copy" {
			// For copies, just delete the copy
			if err := os.Remove(f.NewPath); err != nil {
				log.Error().Err(err).Str("file", f.Filename).Msg("failed to remove copy")
				continue
			}
		} else {
			// For moves, move back to original
			if err := os.Rename(f.NewPath, f.OriginalPath); err != nil {
				log.Error().Err(err).Str("file", f.Filename).Msg("failed to undo move")
				continue
			}
		}

		db.SetFileUndone(f.ID)
	}

	db.UpdateJobStatus(job.ID, "undone")

	// Clean up empty category directories
	cleanEmptyDirs(job.InputPath)

	log.Info().Str("job", job.ID).Msg("undo complete")
	return nil
}

// cleanEmptyDirs removes empty subdirectories and their marker files
func cleanEmptyDirs(rootPath string) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(rootPath, entry.Name())
		markerPath := filepath.Join(dirPath, MarkerFile)

		// Check if it's a managed directory
		if _, err := os.Stat(markerPath); err != nil {
			continue // not managed by us
		}

		// Check if directory is empty (only marker file)
		subEntries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		isEmpty := true
		for _, se := range subEntries {
			if se.Name() != MarkerFile {
				isEmpty = false
				break
			}
		}

		if isEmpty {
			os.Remove(markerPath)
			os.Remove(dirPath)
		}
	}
}
