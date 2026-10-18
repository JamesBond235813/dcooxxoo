package ginutil

import (
	"encoding/base64"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	guestAuthorizationScheme  = "Guest "
	maxGuestAuthorizationSize = 4096
)

// GetGuestCredentials 从 Authorization: Guest <base64url(email)> 读取游客订单身份。
// 兼容旧客户端发送的 email\npassword 格式，但当前查单授权仅依据邮箱。
// 凭证不再接受 URL 查询参数，避免进入代理访问日志、浏览器历史和 Referer。
func GetGuestCredentials(c *gin.Context) (string, string, bool) {
	if c == nil {
		return "", "", false
	}
	value := strings.TrimSpace(c.GetHeader("Authorization"))
	if len(value) <= len(guestAuthorizationScheme) || len(value) > maxGuestAuthorizationSize ||
		!strings.EqualFold(value[:len(guestAuthorizationScheme)], guestAuthorizationScheme) {
		return "", "", false
	}
	encoded := strings.TrimSpace(value[len(guestAuthorizationScheme):])
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(raw), "\n", 2)
	email := strings.ToLower(strings.TrimSpace(parts[0]))
	password := ""
	if len(parts) == 2 {
		password = strings.TrimSpace(parts[1])
	}
	if email == "" || len(email) > 320 || len(password) > 256 {
		return "", "", false
	}
	return email, password, true
}
