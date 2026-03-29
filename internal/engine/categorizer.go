package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"photoorg/internal/database"
	"photoorg/internal/llm"
	"photoorg/internal/sse"

	"github.com/rs/zerolog/log"
)

// resolveCategory cleans up the LLM response and matches it to a known category.
// If no exact match, tries substring matching (e.g., "people in yellow shirts" matches "people").
// Falls back to "misc" if nothing matches.
func resolveCategory(raw string, lowerCategories map[string]bool) string {
	cleaned := strings.Trim(strings.ToLower(raw), `"' .!,`)

	// Exact match
	if lowerCategories[cleaned] {
		return cleaned
	}

	// Substring match: check if any category appears in the response
	for cat := range lowerCategories {
		if strings.Contains(cleaned, cat) {
			return cat
		}
	}

	log.Warn().Str("raw", raw).Str("cleaned", cleaned).Msg("no category match, falling back to misc")
	return "misc"
}

// Categorize runs the AI categorization phase for a job.
// It processes files concurrently using a semaphore and publishes SSE events.
// When job.InstantMove is true, files are moved immediately after categorization.
func Categorize(ctx context.Context, job *database.Job, db *database.DB, provider llm.VisionProvider, broker *sse.Broker, concurrency int) {
	if concurrency < 1 {
		concurrency = 1
	}

	categories := job.CategoriesList()
	lowerCategories := make(map[string]bool, len(categories))
	for _, c := range categories {
		lowerCategories[strings.ToLower(c)] = true
	}

	// Get pending files
	files, err := db.GetPendingFiles(job.ID)
	if err != nil {
		log.Error().Err(err).Str("job", job.ID).Msg("failed to get pending files")
		db.UpdateJobStatus(job.ID, "failed")
		broker.PublishJSON("job_failed", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
		return
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Mutex serializes file moves in instant_move mode to prevent
	// TOCTOU races on collision detection (check-then-rename).
	var moveMu sync.Mutex

	for _, f := range files {
		// Check for cancellation
		select {
		case <-ctx.Done():
			log.Info().Str("job", job.ID).Msg("categorization cancelled")
			db.UpdateJobStatus(job.ID, "cancelled")
			broker.PublishJSON("job_cancelled", map[string]interface{}{"job_id": job.ID})
			return
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // acquire

		go func(file database.File) {
			defer wg.Done()
			defer func() { <-sem }() // release

			// Check cancellation inside goroutine too
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Read image bytes
			imgBytes, err := os.ReadFile(file.OriginalPath)
			if err != nil {
				log.Error().Err(err).Str("file", file.Filename).Msg("failed to read image")
				db.SetFileError(file.ID, err.Error())
				db.IncrementJobErrorCount(job.ID)
				return
			}

			// Call LLM
			rawCategory, err := provider.CategorizeImage(ctx, imgBytes, categories, job.Model, job.CustomPrompt)
			if err != nil {
				log.Error().Err(err).Str("file", file.Filename).Msg("LLM categorization failed")
				db.SetFileError(file.ID, err.Error())
				db.IncrementJobErrorCount(job.ID)

				broker.PublishJSON("file_error", map[string]interface{}{
					"job_id":   job.ID,
					"file_id":  file.ID,
					"filename": file.Filename,
					"error":    err.Error(),
				})
				return
			}

			// Resolve category with substring fallback
			category := resolveCategory(rawCategory, lowerCategories)

			// Save to DB
			if err := db.SetFileCategorized(file.ID, category); err != nil {
				log.Error().Err(err).Str("file", file.Filename).Msg("failed to save categorization")
				return
			}
			db.IncrementJobCategorized(job.ID)

			if job.InstantMove {
				// Instant move: categorize + move + progress in a single SSE event
				moveMu.Lock()
				newPath, moveErr := MoveOneFile(file, category, job.Mode)
				moveMu.Unlock()

				if moveErr != nil {
					log.Error().Err(moveErr).Str("file", file.Filename).Msg("instant move failed")
					db.SetFileError(file.ID, moveErr.Error())
					db.IncrementJobErrorCount(job.ID)
				} else {
					db.SetFileCommitted(file.ID, newPath)
					db.IncrementJobCommitted(job.ID)
				}

				// Single combined event with everything the frontend needs
				updatedJob, _ := db.GetJob(job.ID)
				if updatedJob != nil {
					event := map[string]interface{}{
						"job_id":        job.ID,
						"file_id":       file.ID,
						"filename":      file.Filename,
						"original_path": file.OriginalPath,
						"folder":        filepath.Base(filepath.Dir(file.OriginalPath)),
						"category":      category,
						"categorized":   updatedJob.Categorized,
						"committed":     updatedJob.Committed,
						"total":         updatedJob.TotalFiles,
						"error_count":   updatedJob.ErrorCount,
					}
					if moveErr == nil {
						event["new_path"] = newPath
					}
					broker.PublishJSON("file_organized", event)
				}
			} else {
				// Standard flow: separate events for categorize + progress
				broker.PublishJSON("file_categorized", map[string]interface{}{
					"job_id":        job.ID,
					"file_id":       file.ID,
					"filename":      file.Filename,
					"original_path": file.OriginalPath,
					"category":      category,
				})

				updatedJob, _ := db.GetJob(job.ID)
				if updatedJob != nil {
					broker.PublishJSON("job_progress", map[string]interface{}{
						"job_id":      job.ID,
						"categorized": updatedJob.Categorized,
						"total":       updatedJob.TotalFiles,
						"error_count": updatedJob.ErrorCount,
					})
				}
			}
		}(f)
	}

	wg.Wait()

	// Check if cancelled during processing
	select {
	case <-ctx.Done():
		db.UpdateJobStatus(job.ID, "cancelled")
		broker.PublishJSON("job_cancelled", map[string]interface{}{"job_id": job.ID})
		return
	default:
	}

	finalJob, _ := db.GetJob(job.ID)

	if job.InstantMove {
		// Instant move: files already moved, set final status
		db.UpdateJobStatus(job.ID, "committed")

		if finalJob != nil {
			broker.PublishJSON("job_completed", map[string]interface{}{
				"job_id":       job.ID,
				"total":        finalJob.TotalFiles,
				"categorized":  finalJob.Categorized,
				"committed":    finalJob.Committed,
				"errors":       finalJob.ErrorCount,
				"instant_move": true,
			})
		}

		log.Info().Str("job", job.ID).Int("total", job.TotalFiles).Int("committed", finalJob.Committed).Msg("instant move complete")
	} else {
		// Standard flow: mark as reviewing for manual review
		db.UpdateJobStatus(job.ID, "reviewing")

		if finalJob != nil {
			broker.PublishJSON("job_completed", map[string]interface{}{
				"job_id":      job.ID,
				"total":       finalJob.TotalFiles,
				"categorized": finalJob.Categorized,
				"errors":      finalJob.ErrorCount,
			})
		}

		log.Info().Str("job", job.ID).Int("total", job.TotalFiles).Msg("categorization complete")
	}
}
