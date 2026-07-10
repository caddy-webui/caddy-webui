package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/caddy-webui/caddy-webui/internal/config"
)

var DB *sql.DB

func Init() error {
	dbPath := filepath.Join("/opt/caddy-webui/data", "caddy-webui.db")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("创建数据库表失败: %w", err)
	}

	config.Info("数据库初始化成功: %s", dbPath)
	return nil
}

func createTables() error {
	ddl := `
	CREATE TABLE IF NOT EXISTS settings (
		key         TEXT    PRIMARY KEY,
		value       TEXT    NOT NULL,
		updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS sites (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		domain          TEXT    NOT NULL UNIQUE,
		enabled         INTEGER NOT NULL DEFAULT 1,
		proxy_target    TEXT    DEFAULT '',
		proxy_config    TEXT    DEFAULT '{}',
		cert_mode       TEXT    NOT NULL DEFAULT 'auto',
		cert_status     TEXT    NOT NULL DEFAULT 'none',
		cert_expires_at TEXT    DEFAULT NULL,
		cert_file_path  TEXT    DEFAULT NULL,
		key_file_path   TEXT    DEFAULT NULL,
		created_at      TEXT    NOT NULL DEFAULT (datetime('now')),
		updated_at      TEXT    NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_sites_domain ON sites(domain);
	CREATE INDEX IF NOT EXISTS idx_sites_enabled ON sites(enabled);
	CREATE INDEX IF NOT EXISTS idx_sites_cert_status ON sites(cert_status);
	CREATE INDEX IF NOT EXISTS idx_sites_cert_mode ON sites(cert_mode);
	`

	_, err := DB.Exec(ddl)
	return err
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
