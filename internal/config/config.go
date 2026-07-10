package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type AppConfig struct {
	Server ServerConfig `toml:"server"`
	Log    LogConfig    `toml:"log"`
	Caddy  CaddyConfig  `toml:"caddy"`
}

type ServerConfig struct {
	Port int    `toml:"port"`
	Host string `toml:"host"`
}

type LogConfig struct {
	Level string `toml:"level"`
	Dir   string `toml:"dir"`
}

type CaddyConfig struct {
	BinaryPath  string `toml:"binary_path"`
	ConfigPath  string `toml:"config_path"`
	ServiceName string `toml:"service_name"`
	AdminAPI    string `toml:"admin_api"`
}

var (
	Cfg      *AppConfig
	cfgPath  string
)

func init() {
	flag.StringVar(&cfgPath, "config", "/opt/caddy-webui/config/config.toml", "配置文件路径")
}

func Load() error {
	flag.Parse()

	Cfg = &AppConfig{
		Server: ServerConfig{
			Port: 8729,
			Host: "127.0.0.1",
		},
		Log: LogConfig{
			Level: "INFO",
			Dir:   "/opt/caddy-webui/log/",
		},
		Caddy: CaddyConfig{
			BinaryPath:  "/usr/bin/caddy",
			ConfigPath:  "/opt/caddy-webui/config/Caddyfile",
			ServiceName: "caddy",
			AdminAPI:    "http://localhost:2019",
		},
	}

	if _, err := os.Stat(cfgPath); err == nil {
		if _, err := toml.DecodeFile(cfgPath, Cfg); err != nil {
			return fmt.Errorf("解析配置文件失败: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	return nil
}

func Addr() string {
	return fmt.Sprintf("%s:%d", Cfg.Server.Host, Cfg.Server.Port)
}
