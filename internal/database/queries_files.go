package database

import (
	"fmt"
	"time"
)

func (d *DB) InsertFiles(files []File) error {
	tx, err := d.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt := `INSERT INTO files (job_id, original_path, filename, file_size, status) VALUES (?, ?, ?, ?, 'pending')`
	for _, f := range files {
		if _, err := tx.Exec(stmt, f.JobID, f.OriginalPath, f.Filename, f.FileSize); err != nil {
			return fmt.Errorf("insert file %s: %w", f.Filename, err)
		}
	}
	return tx.Commit()
}

func (d *DB) GetPendingFiles(jobID string) ([]File, error) {
	files := make([]File, 0)
	err := d.Select(&files, `SELECT * FROM files WHERE job_id = ? AND status = 'pending'`, jobID)
	return files, err
}

func (d *DB) GetFilesByJob(jobID string, category string, status string, page, perPage int) ([]File, int, error) {
	if perPage <= 0 {
		perPage = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	where := "WHERE job_id = ?"
	args := []interface{}{jobID}

	if category != "" {
		where += " AND final_category = ?"
		args = append(args, category)
	}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := d.Get(&total, fmt.Sprintf("SELECT COUNT(*) FROM files %s", where), countArgs...); err != nil {
		return nil, 0, err
	}

	files := make([]File, 0)
	query := fmt.Sprintf("SELECT * FROM files %s ORDER BY filename ASC LIMIT ? OFFSET ?", where)
	args = append(args, perPage, offset)
	if err := d.Select(&files, query, args...); err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (d *DB) GetCategorySummary(jobID string) ([]CategorySummary, error) {
	summary := make([]CategorySummary, 0)
	err := d.Select(&summary, `
		SELECT final_category, COUNT(*) as count
		FROM files
		WHERE job_id = ? AND status IN ('categorized', 'committed')
		GROUP BY final_category
		ORDER BY count DESC
	`, jobID)
	return summary, err
}

func (d *DB) SetFileCategorized(fileID int64, category string) error {
	now := time.Now().UnixMilli()
	_, err := d.Exec(`
		UPDATE files SET ai_category = ?, final_category = ?, status = 'categorized', categorized_at = ?
		WHERE id = ?
	`, category, category, now, fileID)
	return err
}

func (d *DB) SetFileError(fileID int64, errMsg string) error {
	_, err := d.Exec(`
		UPDATE files SET status = 'error', error_message = ?
		WHERE id = ?
	`, errMsg, fileID)
	return err
}

func (d *DB) UpdateFileCategory(fileID int64, category string) error {
	_, err := d.Exec(`UPDATE files SET final_category = ? WHERE id = ?`, category, fileID)
	return err
}

func (d *DB) BulkUpdateFileCategory(fileIDs []int64, category string) error {
	if len(fileIDs) == 0 {
		return nil
	}

	tx, err := d.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range fileIDs {
		if _, err := tx.Exec(`UPDATE files SET final_category = ? WHERE id = ?`, category, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) GetCategorizedFiles(jobID string) ([]File, error) {
	files := make([]File, 0)
	err := d.Select(&files, `SELECT * FROM files WHERE job_id = ? AND status = 'categorized'`, jobID)
	return files, err
}

func (d *DB) GetCommittedFiles(jobID string) ([]File, error) {
	files := make([]File, 0)
	err := d.Select(&files, `SELECT * FROM files WHERE job_id = ? AND status = 'committed'`, jobID)
	return files, err
}

func (d *DB) SetFileCommitted(fileID int64, newPath string) error {
	now := time.Now().UnixMilli()
	_, err := d.Exec(`
		UPDATE files SET new_path = ?, status = 'committed', committed_at = ?
		WHERE id = ?
	`, newPath, now, fileID)
	return err
}

func (d *DB) SetFileUndone(fileID int64) error {
	_, err := d.Exec(`UPDATE files SET status = 'undone', new_path = '' WHERE id = ?`, fileID)
	return err
}
