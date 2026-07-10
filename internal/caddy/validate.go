package caddy

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/caddy-webui/caddy-webui/internal/config"
)

func ValidateCaddyfile(content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, config.Cfg.Caddy.BinaryPath, "validate", "--config", "/dev/stdin")
	cmd.Stdin = bytes.NewReader([]byte(content))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}

	return nil
}
