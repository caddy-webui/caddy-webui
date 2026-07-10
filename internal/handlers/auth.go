package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/caddy-webui/caddy-webui/internal/auth"
	"github.com/caddy-webui/caddy-webui/internal/database"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func HandleSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	initialized, err := database.IsInitialized()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "检查初始化状态失败")
		return
	}
	if initialized {
		ErrorResponse(w, http.StatusConflict, 40901, "系统已初始化，禁止重复初始化")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	if !usernameRegex.MatchString(req.Username) {
		ErrorResponse(w, http.StatusBadRequest, 40001, "用户名格式不正确，仅允许3-32位字母、数字和下划线")
		return
	}
	if len(req.Password) < 6 || len(req.Password) > 64 {
		ErrorResponse(w, http.StatusBadRequest, 40001, "密码长度必须在6-64个字符之间")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "密码加密失败")
		return
	}

	if err := database.SetSetting("admin_username", req.Username); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "保存管理员账号失败")
		return
	}
	if err := database.SetSetting("admin_password_hash", hash); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "保存管理员密码失败")
		return
	}
	if err := database.SetSetting("initialized", "true"); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "保存初始化标记失败")
		return
	}

	SuccessResponse(w, "初始化成功", nil)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	locked, remaining, err := auth.CheckAccountLock()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "检查锁定状态失败")
		return
	}
	if locked {
		ErrorResponse(w, http.StatusForbidden, 40301, fmt.Sprintf("账号已锁定，请 %d 分钟后重试", remaining))
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	storedUsername, _ := database.GetAdminUsername()
	storedHash, _ := database.GetAdminPasswordHash()

	if req.Username != storedUsername || !auth.CheckPassword(req.Password, storedHash) {
		auth.RecordLoginFailure()
		ErrorResponse(w, http.StatusUnauthorized, 40101, "用户名或密码错误")
		return
	}

	auth.ResetLoginFailure()

	token, expiresAt, err := auth.GenerateToken(req.Username)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "生成令牌失败")
		return
	}

	SuccessResponse(w, "登录成功", map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt.Format("2006-01-02T15:04:05Z"),
	})
}

func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		ErrorResponse(w, http.StatusMethodNotAllowed, 40001, "方法不允许")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, 40001, "请求格式不正确")
		return
	}

	storedHash, _ := database.GetAdminPasswordHash()
	if !auth.CheckPassword(req.OldPassword, storedHash) {
		ErrorResponse(w, http.StatusUnauthorized, 40101, "旧密码不正确")
		return
	}

	if len(req.NewPassword) < 6 {
		ErrorResponse(w, http.StatusBadRequest, 40001, "新密码长度不能少于6个字符")
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "密码加密失败")
		return
	}

	if err := database.SetSetting("admin_password_hash", hash); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, 50001, "更新密码失败")
		return
	}

	SuccessResponse(w, "密码修改成功", nil)
}

func HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	initialized, _ := database.IsInitialized()
	username, _ := database.GetAdminUsername()

	SuccessResponse(w, "success", map[string]interface{}{
		"initialized": initialized,
		"username":    username,
	})
}
