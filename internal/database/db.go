package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

type DB struct {
	*sqlx.DB
}

func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	dsn := fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-64000)&_pragma=foreign_keys(ON)",
		dbPath,
	)

	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	d := &DB{DB: db}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	log.Info().Str("path", dbPath).Msg("database initialized")
	return d, nil
}

func (d *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id            TEXT PRIMARY KEY,
		input_path    TEXT NOT NULL,
		status        TEXT NOT NULL DEFAULT 'categorizing'
			CHECK (status IN ('categorizing','reviewing','committing','committed','undone','failed','cancelled')),
		mode          TEXT NOT NULL DEFAULT 'move'
			CHECK (mode IN ('move','copy')),
		categories    TEXT NOT NULL DEFAULT '[]',
		provider      TEXT NOT NULL,
		model         TEXT NOT NULL,
		endpoint      TEXT NOT NULL DEFAULT '',
		concurrency   INTEGER NOT NULL DEFAULT 4,
		custom_prompt TEXT NOT NULL DEFAULT '',
		total_files   INTEGER NOT NULL DEFAULT 0,
		categorized   INTEGER NOT NULL DEFAULT 0,
		committed     INTEGER NOT NULL DEFAULT 0,
		error_count   INTEGER NOT NULL DEFAULT 0,
		instant_move  INTEGER NOT NULL DEFAULT 0,
		created_at    INTEGER NOT NULL,
		updated_at    INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS files (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id         TEXT NOT NULL REFERENCES jobs(id),
		original_path  TEXT NOT NULL,
		filename       TEXT NOT NULL,
		file_size      INTEGER NOT NULL DEFAULT 0,
		ai_category    TEXT NOT NULL DEFAULT '',
		final_category TEXT NOT NULL DEFAULT '',
		new_path       TEXT NOT NULL DEFAULT '',
		status         TEXT NOT NULL DEFAULT 'pending'
			CHECK (status IN ('pending','categorized','error','committed','undone','skipped')),
		error_message  TEXT,
		categorized_at INTEGER,
		committed_at   INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_files_job ON files(job_id);
	CREATE INDEX IF NOT EXISTS idx_files_status ON files(status);
	CREATE INDEX IF NOT EXISTS idx_files_category ON files(job_id, final_category);

	CREATE TABLE IF NOT EXISTS app_config (
		key        TEXT PRIMARY KEY,
		value      TEXT NOT NULL DEFAULT '',
		updated_at INTEGER NOT NULL DEFAULT 0
	);
	`

	if _, err := d.Exec(schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	// Seed default config values
	now := time.Now().UnixMilli()
	defaults := []struct {
		Key   string
		Value string
	}{
		{"provider", "ollama"},
		{"model", "llava:7b"},
		{"endpoint", "http://localhost:11434"},
		{"api_key", ""},
		{"concurrency", "4"},
		{"mode", "move"},
		{"categories", `["people","food","landscape","animals","documents","misc"]`},
		{"custom_prompt", ""},
		{"instant_move", "false"},
	}

	stmt := `INSERT OR IGNORE INTO app_config (key, value, updated_at) VALUES (?, ?, ?)`
	for _, cfg := range defaults {
		if _, err := d.Exec(stmt, cfg.Key, cfg.Value, now); err != nil {
			return fmt.Errorf("seed config %s: %w", cfg.Key, err)
		}
	}

	// Additive migrations for existing databases
	alterStatements := []string{
		`ALTER TABLE jobs ADD COLUMN instant_move INTEGER NOT NULL DEFAULT 0`,
	}
	for _, stmt := range alterStatements {
		if _, err := d.Exec(stmt); err != nil {
			if !strings.Contains(err.Error(), "duplicate column") {
				return fmt.Errorf("alter table: %w", err)
			}
		}
	}

	// Drop stale auto_commit column (replaced by instant_move)
	var autoCommitSQL string
	acRow := d.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='jobs'`)
	if err := acRow.Scan(&autoCommitSQL); err == nil && strings.Contains(autoCommitSQL, "auto_commit") {
		cols := `id, input_path, status, mode, categories, provider, model, endpoint, concurrency, custom_prompt, total_files, categorized, committed, error_count, instant_move, created_at, updated_at`
		d.Exec(`PRAGMA foreign_keys=OFF`)
		d.Exec(`PRAGMA legacy_alter_table=ON`)
		d.Exec(`ALTER TABLE jobs RENAME TO jobs_autocommit_old`)
		d.Exec(schema)
		d.Exec(fmt.Sprintf(`INSERT INTO jobs (%s) SELECT %s FROM jobs_autocommit_old`, cols, cols))
		d.Exec(`DROP TABLE jobs_autocommit_old`)
		d.Exec(`PRAGMA legacy_alter_table=OFF`)
		d.Exec(`PRAGMA foreign_keys=ON`)
	}

	// Migration: add 'committing' to job status CHECK constraint
	// SQLite can't ALTER CHECK constraints, so recreate the table if needed
	var createSQL string
	row := d.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='jobs'`)
	if err := row.Scan(&createSQL); err == nil && !strings.Contains(createSQL, "committing") {
		cols := `id, input_path, status, mode, categories, provider, model, endpoint, concurrency, custom_prompt, total_files, categorized, committed, error_count, instant_move, created_at, updated_at`
		d.Exec(`PRAGMA foreign_keys=OFF`)
		d.Exec(`PRAGMA legacy_alter_table=ON`) // prevent SQLite from updating FK references in other tables
		d.Exec(`ALTER TABLE jobs RENAME TO jobs_old`)
		d.Exec(schema) // recreates jobs table with new CHECK constraint
		d.Exec(fmt.Sprintf(`INSERT INTO jobs (%s) SELECT id, input_path, status, mode, categories, provider, model, endpoint, concurrency, custom_prompt, total_files, categorized, committed, error_count, COALESCE(instant_move, auto_commit, 0), created_at, updated_at FROM jobs_old`, cols))
		d.Exec(`DROP TABLE jobs_old`)
		d.Exec(`PRAGMA legacy_alter_table=OFF`)
		d.Exec(`PRAGMA foreign_keys=ON`)
	}

	// Fix files table FK if prior migration corrupted it (pointed to jobs_old instead of jobs)
	var filesSQL string
	fkRow := d.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='files'`)
	if err := fkRow.Scan(&filesSQL); err == nil && strings.Contains(filesSQL, "jobs_old") {
		d.Exec(`PRAGMA foreign_keys=OFF`)
		d.Exec(`ALTER TABLE files RENAME TO files_old`)
		d.Exec(schema) // recreates files table with correct FK
		d.Exec(`INSERT INTO files SELECT * FROM files_old`)
		d.Exec(`DROP TABLE files_old`)
		d.Exec(`PRAGMA foreign_keys=ON`)
	}

	// Fix data corrupted by column-order mismatch in prior migration
	d.Exec(`UPDATE jobs SET created_at = updated_at, updated_at = instant_move, instant_move = 0 WHERE instant_move > 1000000000000`)

	return nil
}
