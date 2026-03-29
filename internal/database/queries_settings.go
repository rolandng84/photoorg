package database

import "time"

func (d *DB) GetAllConfig() ([]AppConfig, error) {
	configs := make([]AppConfig, 0)
	err := d.Select(&configs, `SELECT key, value, updated_at FROM app_config ORDER BY key`)
	return configs, err
}

func (d *DB) GetConfigValue(key, fallback string) string {
	var value string
	err := d.Get(&value, `SELECT value FROM app_config WHERE key = ?`, key)
	if err != nil {
		return fallback
	}
	return value
}

func (d *DB) GetConfigMap() (map[string]string, error) {
	configs, err := d.GetAllConfig()
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(configs))
	for _, c := range configs {
		m[c.Key] = c.Value
	}
	return m, nil
}

func (d *DB) SetConfigValue(key, value string) error {
	now := time.Now().UnixMilli()
	_, err := d.Exec(
		`INSERT INTO app_config (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?`,
		key, value, now, value, now,
	)
	return err
}

func (d *DB) SetConfigBulk(settings map[string]string) error {
	tx, err := d.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UnixMilli()
	stmt := `INSERT INTO app_config (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?`
	for k, v := range settings {
		if _, err := tx.Exec(stmt, k, v, now, v, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}
