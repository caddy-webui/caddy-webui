package auth

import (
	"fmt"
	"time"

	"github.com/caddy-webui/caddy-webui/internal/database"
)

func CheckAccountLock() (bool, int, error) {
	lockedUntil, err := database.GetLockedUntil()
	if err != nil {
		return false, 0, err
	}
	if lockedUntil == "" {
		return false, 0, nil
	}

	t, err := time.Parse(time.RFC3339, lockedUntil)
	if err != nil {
		return false, 0, nil
	}

	if time.Now().Before(t) {
		remaining := int(time.Until(t).Minutes())
		if remaining < 1 {
			remaining = 1
		}
		return true, remaining, nil
	}

	return false, 0, nil
}

func RecordLoginFailure() error {
	count, err := database.GetLoginFailCount()
	if err != nil {
		return err
	}

	count++
	if err := database.SetLoginFailCount(count); err != nil {
		return err
	}

	if count >= 5 {
		lockedUntil := time.Now().Add(15 * time.Minute).Format(time.RFC3339)
		if err := database.SetLockedUntil(lockedUntil); err != nil {
			return err
		}
		return fmt.Errorf("账号已锁定，请 15 分钟后重试")
	}

	return nil
}

func ResetLoginFailure() error {
	if err := database.SetLoginFailCount(0); err != nil {
		return err
	}
	return database.SetLockedUntil("")
}
