package database

import (
	"database/sql"
	"encoding/json"
)

// Job represents a categorization job
type Job struct {
	ID            string       `db:"id" json:"id"`
	InputPath     string       `db:"input_path" json:"input_path"`
	Status        string       `db:"status" json:"status"`
	Mode          string       `db:"mode" json:"mode"`
	Categories    string       `db:"categories" json:"-"` // JSON array stored as text
	Provider      string       `db:"provider" json:"provider"`
	Model         string       `db:"model" json:"model"`
	Endpoint      string       `db:"endpoint" json:"endpoint"`
	Concurrency   int          `db:"concurrency" json:"concurrency"`
	CustomPrompt  string       `db:"custom_prompt" json:"custom_prompt"`
	InstantMove   bool         `db:"instant_move" json:"instant_move"`
	TotalFiles    int          `db:"total_files" json:"total_files"`
	Categorized   int          `db:"categorized" json:"categorized"`
	Committed     int          `db:"committed" json:"committed"`
	ErrorCount    int          `db:"error_count" json:"error_count"`
	CreatedAt     int64        `db:"created_at" json:"created_at"`
	UpdatedAt     int64        `db:"updated_at" json:"updated_at"`
}

// CategoriesList returns the categories as a string slice
func (j *Job) CategoriesList() []string {
	var cats []string
	if err := json.Unmarshal([]byte(j.Categories), &cats); err != nil {
		return make([]string, 0)
	}
	return cats
}

// JobJSON is the JSON representation including parsed categories
type JobJSON struct {
	Job
	CategoriesParsed []string `json:"categories"`
}

func (j *Job) ToJSON() JobJSON {
	return JobJSON{
		Job:              *j,
		CategoriesParsed: j.CategoriesList(),
	}
}

// File represents a file record within a job
type File struct {
	ID            int64          `db:"id" json:"id"`
	JobID         string         `db:"job_id" json:"job_id"`
	OriginalPath  string         `db:"original_path" json:"original_path"`
	Filename      string         `db:"filename" json:"filename"`
	FileSize      int64          `db:"file_size" json:"file_size"`
	AICategory    string         `db:"ai_category" json:"ai_category"`
	FinalCategory string         `db:"final_category" json:"final_category"`
	NewPath       string         `db:"new_path" json:"new_path"`
	Status        string         `db:"status" json:"status"`
	ErrorMessage  sql.NullString `db:"error_message" json:"-"`
	CategorizedAt sql.NullInt64  `db:"categorized_at" json:"-"`
	CommittedAt   sql.NullInt64  `db:"committed_at" json:"-"`
}

// FileJSON is the JSON-safe representation
type FileJSON struct {
	ID            int64   `json:"id"`
	JobID         string  `json:"job_id"`
	OriginalPath  string  `json:"original_path"`
	Filename      string  `json:"filename"`
	FileSize      int64   `json:"file_size"`
	AICategory    string  `json:"ai_category"`
	FinalCategory string  `json:"final_category"`
	NewPath       string  `json:"new_path"`
	Status        string  `json:"status"`
	ErrorMessage  *string `json:"error_message"`
	CategorizedAt *int64  `json:"categorized_at"`
	CommittedAt   *int64  `json:"committed_at"`
}

func (f *File) ToJSON() FileJSON {
	fj := FileJSON{
		ID:            f.ID,
		JobID:         f.JobID,
		OriginalPath:  f.OriginalPath,
		Filename:      f.Filename,
		FileSize:      f.FileSize,
		AICategory:    f.AICategory,
		FinalCategory: f.FinalCategory,
		NewPath:       f.NewPath,
		Status:        f.Status,
	}
	if f.ErrorMessage.Valid {
		fj.ErrorMessage = &f.ErrorMessage.String
	}
	if f.CategorizedAt.Valid {
		fj.CategorizedAt = &f.CategorizedAt.Int64
	}
	if f.CommittedAt.Valid {
		fj.CommittedAt = &f.CommittedAt.Int64
	}
	return fj
}

// AppConfig represents a config key-value pair
type AppConfig struct {
	Key       string `db:"key" json:"key"`
	Value     string `db:"value" json:"value"`
	UpdatedAt int64  `db:"updated_at" json:"updated_at"`
}

// CategorySummary represents a count per category for a job
type CategorySummary struct {
	Category string `db:"final_category" json:"category"`
	Count    int    `db:"count" json:"count"`
}
