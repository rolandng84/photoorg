package database

import "time"

func (d *DB) CreateJob(job *Job) error {
	_, err := d.NamedExec(`
		INSERT INTO jobs (id, input_path, status, mode, categories, provider, model, endpoint, concurrency, custom_prompt, instant_move, total_files, categorized, committed, error_count, created_at, updated_at)
		VALUES (:id, :input_path, :status, :mode, :categories, :provider, :model, :endpoint, :concurrency, :custom_prompt, :instant_move, :total_files, :categorized, :committed, :error_count, :created_at, :updated_at)
	`, job)
	return err
}

func (d *DB) GetJob(id string) (*Job, error) {
	var job Job
	err := d.Get(&job, `SELECT * FROM jobs WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (d *DB) ListJobs() ([]Job, error) {
	jobs := make([]Job, 0)
	err := d.Select(&jobs, `SELECT * FROM jobs ORDER BY created_at DESC`)
	return jobs, err
}

func (d *DB) UpdateJobStatus(id, status string) error {
	now := time.Now().UnixMilli()
	_, err := d.Exec(`UPDATE jobs SET status = ?, updated_at = ? WHERE id = ?`, status, now, id)
	return err
}

func (d *DB) IncrementJobCategorized(id string) error {
	now := time.Now().UnixMilli()
	_, err := d.Exec(`UPDATE jobs SET categorized = categorized + 1, updated_at = ? WHERE id = ?`, now, id)
	return err
}

func (d *DB) IncrementJobErrorCount(id string) error {
	now := time.Now().UnixMilli()
	_, err := d.Exec(`UPDATE jobs SET error_count = error_count + 1, updated_at = ? WHERE id = ?`, now, id)
	return err
}

func (d *DB) SetJobErrorCount(id string, count int) error {
	now := time.Now().UnixMilli()
	_, err := d.Exec(`UPDATE jobs SET error_count = ?, updated_at = ? WHERE id = ?`, count, now, id)
	return err
}

func (d *DB) IncrementJobCommitted(id string) error {
	now := time.Now().UnixMilli()
	_, err := d.Exec(`UPDATE jobs SET committed = committed + 1, updated_at = ? WHERE id = ?`, now, id)
	return err
}

func (d *DB) DeleteJob(id string) error {
	tx, err := d.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM files WHERE job_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM jobs WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
