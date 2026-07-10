package caddy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/caddy-webui/caddy-webui/internal/config"
)

func ReloadCaddy(caddyfileContent string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payload := map[string]interface{}{
		"config": json.RawMessage(caddyfileContent),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.Cfg.Caddy.AdminAPI+"/load", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if rbErr := RollbackCaddyfile(); rbErr != nil {
			config.Error("回滚 Caddyfile 失败: %v", rbErr)
		}
		return fmt.Errorf("Caddy 重载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		if rbErr := RollbackCaddyfile(); rbErr != nil {
			config.Error("回滚 Caddyfile 失败: %v", rbErr)
		}
		return fmt.Errorf("Caddy 重载失败: %s", string(respBody))
	}

	return nil
}

func GetCertificates() ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", config.Cfg.Caddy.AdminAPI+"/certificates", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取证书信息失败: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	certs, ok := result["certificates"].([]map[string]interface{})
	if !ok {
		return nil, nil
	}

	return certs, nil
}
