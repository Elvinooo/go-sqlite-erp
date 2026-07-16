package middleware

import (
	"net/http"
	"strings"

	"erp/internal/shared/response"
	"github.com/gin-gonic/gin"
)

const demoModeMessage = "当前为演示模式，不可修改"

func DemoMode(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enabled && isDemoModeBlocked(c.Request.Method, c.Request.URL.Path) {
			response.Fail(c, http.StatusForbidden, 40301, demoModeMessage)
			c.Abort()
			return
		}
		c.Next()
	}
}

func isDemoModeBlocked(method string, path string) bool {
	if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
		return false
	}
	path = strings.TrimRight(path, "/")
	if path == "/api/v1/auth/password" {
		return true
	}
	for _, prefix := range []string{"/api/v1/users", "/api/v1/roles", "/api/v1/permissions", "/api/v1/menus"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
