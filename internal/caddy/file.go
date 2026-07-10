package caddy

import (
	"io"
	"os"

	"github.com/caddy-webui/caddy-webui/internal/config"
)

func ReadCaddyfile() (string, error) {
	path := config.Cfg.Caddy.ConfigPath
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func WriteCaddyfile(content string) error {
	path := config.Cfg.Caddy.ConfigPath

	if err := os.MkdirAll(configPath(), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		if err := backupCaddyfile(path); err != nil {
			return err
		}
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func RollbackCaddyfile() error {
	path := config.Cfg.Caddy.ConfigPath
	bakPath := path + ".bak"

	if _, err := os.Stat(bakPath); err != nil {
		return err
	}

	src, err := os.Open(bakPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func backupCaddyfile(path string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path + ".bak")
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func configPath() string {
	return config.Cfg.Caddy.ConfigPath[:len(config.Cfg.Caddy.ConfigPath)-len("/Caddyfile")]
}
