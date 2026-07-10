package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/caddy-webui/caddy-webui/internal/config"
)

func GetSetting(key string) (string, error) {
	var value string
	err := DB.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func SetSetting(key, value string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := DB.Exec(
		"INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?",
		key, value, now, value, now,
	)
	return err
}

func IsInitialized() (bool, error) {
	val, err := GetSetting("initialized")
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

func GetAdminUsername() (string, error) {
	return GetSetting("admin_username")
}

func GetAdminPasswordHash() (string, error) {
	return GetSetting("admin_password_hash")
}

func GetJWTSecret() (string, error) {
	secret, err := GetSetting("jwt_secret")
	if err != nil {
		return "", err
	}
	if secret == "" {
		secret = generateRandomHex(32)
		if err := SetSetting("jwt_secret", secret); err != nil {
			return "", err
		}
	}
	return secret, nil
}

func GetLoginFailCount() (int, error) {
	val, err := GetSetting("login_fail_count")
	if err != nil || val == "" {
		return 0, err
	}
	var count int
	fmt.Sscanf(val, "%d", &count)
	return count, nil
}

func SetLoginFailCount(count int) error {
	return SetSetting("login_fail_count", fmt.Sprintf("%d", count))
}

func GetLockedUntil() (string, error) {
	return GetSetting("locked_until")
}

func SetLockedUntil(until string) error {
	return SetSetting("locked_until", until)
}

func GetServerPort() (int, error) {
	val, err := GetSetting("server_port")
	if err != nil || val == "" {
		return config.Cfg.Server.Port, err
	}
	var port int
	fmt.Sscanf(val, "%d", &port)
	return port, nil
}

func SetServerPort(port int) error {
	return SetSetting("server_port", fmt.Sprintf("%d", port))
}

func GetLogLevel() (string, error) {
	val, err := GetSetting("log_level")
	if err != nil || val == "" {
		return config.Cfg.Log.Level, err
	}
	return val, nil
}

func SetLogLevel(level string) error {
	return SetSetting("log_level", level)
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
