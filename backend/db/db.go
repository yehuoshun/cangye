package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dbDir := filepath.Join(home, ".cangye")
		os.MkdirAll(dbDir, 0755)
		dbPath = filepath.Join(dbDir, "cangye.db")
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	if _, err := DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	return migrate()
}

func migrate() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS prefix_config (
			prefix TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			map_path TEXT,
			url_template TEXT,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS collections (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			icon TEXT DEFAULT '📁',
			parent_id TEXT,
			sort_order INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now')),
			FOREIGN KEY (parent_id) REFERENCES collections(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS collection_paths (
			id TEXT PRIMARY KEY,
			collection_id TEXT NOT NULL,
			path TEXT NOT NULL,
			auto_scan INTEGER DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now')),
			FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS virtual_files (
			id TEXT PRIMARY KEY,
			collection_id TEXT NOT NULL,
			path TEXT NOT NULL,
			display_name TEXT,
			size INTEGER DEFAULT 0,
			sort_order INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now')),
			FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS scan_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			collection_path_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_size INTEGER,
			mod_time TEXT,
			mime_type TEXT,
			cached_at TEXT DEFAULT (datetime('now')),
			FOREIGN KEY (collection_path_id) REFERENCES collection_paths(id) ON DELETE CASCADE,
			UNIQUE(collection_path_id, file_path)
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			color TEXT NOT NULL DEFAULT 'gray'
		)`,
		`CREATE TABLE IF NOT EXISTS file_tags (
			file_id TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			PRIMARY KEY (file_id, tag_id),
			FOREIGN KEY (file_id) REFERENCES virtual_files(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS collection_tags (
			collection_id TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			PRIMARY KEY (collection_id, tag_id),
			FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	}

	for _, s := range schemas {
		if _, err := DB.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w\nsql: %s", err, s)
		}
	}

	// init default prefixes
	defaultPrefixes := []struct{ prefix, ptype, mapPath, urlTemplate string }{
		{"115:", "local", "", ""},
		{"tg:", "web", "", ""},
		{"notion:", "web", "", ""},
	}
	for _, p := range defaultPrefixes {
		_, err := DB.Exec(
			"INSERT OR IGNORE INTO prefix_config (prefix, type, map_path, url_template) VALUES (?,?,?,?)",
			p.prefix, p.ptype, p.mapPath, p.urlTemplate,
		)
		if err != nil {
			return fmt.Errorf("init prefix %s: %w", p.prefix, err)
		}
	}

	// init default settings
	defaultSettings := map[string]string{
		"layout": "sidebar",
		"view":   "grid",
		"theme":  "dark",
	}
	for k, v := range defaultSettings {
		_, err := DB.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES (?,?)", k, v)
		if err != nil {
			return fmt.Errorf("init setting %s: %w", k, err)
		}
	}

	log.Println("[db] initialized:", dbPath)
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
