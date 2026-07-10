package caddy

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/caddy-webui/caddy-webui/internal/config"
)

func StartCaddy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "start", config.Cfg.Caddy.ServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func StopCaddy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "stop", config.Cfg.Caddy.ServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func RestartCaddy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "restart", config.Cfg.Caddy.ServiceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func GetCaddyStatus() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "is-active", config.Cfg.Caddy.ServiceName)
	output, err := cmd.Output()
	status := strings.TrimSpace(string(output))

	if err != nil {
		if status == "inactive" || status == "stopped" {
			return "stopped", nil
		}
		return "unknown", nil
	}

	if status == "active" {
		return "running", nil
	}

	return status, nil
}

func GetCaddyVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, config.Cfg.Caddy.BinaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	parts := strings.Split(strings.TrimSpace(string(output)), " ")
	return parts[0]
}
