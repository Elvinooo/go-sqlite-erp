package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	authmodel "erp/internal/domain/auth/model"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OperationAudit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipOperationAudit(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}
		start := time.Now()
		requestBody := captureRequestBody(c)
		c.Next()
		if c.GetBool("skipOperationAudit") {
			return
		}
		tenantID := utils.GetTenantID(c)
		userID := utils.GetUserID(c)
		var userIDPtr *uint64
		if userID > 0 {
			userIDPtr = &userID
		}
		log := authmodel.OperationLog{
			TenantID: tenantID, UserID: userIDPtr, Username: c.GetString(utils.ContextUsername),
			Module: moduleFromPath(c.Request.URL.Path), Action: actionFromRequest(c.Request.Method, c.Request.URL.Path),
			Method: c.Request.Method, Path: c.Request.URL.Path, IP: c.ClientIP(),
			UserAgent: c.Request.UserAgent(), StatusCode: c.Writer.Status(),
			RequestBody: requestBody, CostMS: time.Since(start).Milliseconds(), CreatedAt: time.Now(),
		}
		_ = db.WithContext(c.Request.Context()).Create(&log).Error
	}
}

func shouldSkipOperationAudit(method string, path string) bool {
	return method == http.MethodGet ||
		strings.HasPrefix(path, "/swagger") ||
		strings.HasSuffix(path, "/settings/restore-test-data")
}

func captureRequestBody(c *gin.Context) string {
	if c.Request.Body == nil || c.Request.Method == http.MethodGet {
		return ""
	}
	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return captureMultipartSummary(c)
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 64*1024))
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	if len(body) == 0 {
		return ""
	}
	summary := sanitizeBody(body, contentType)
	if len(summary) > 4000 {
		return summary[:4000] + "...(truncated)"
	}
	return summary
}

func captureMultipartSummary(c *gin.Context) string {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return `{"type":"multipart"}`
	}
	summary := map[string]any{"type": "multipart", "fields": map[string]any{}, "files": []map[string]any{}}
	fields := summary["fields"].(map[string]any)
	if c.Request.MultipartForm != nil {
		for key, values := range c.Request.MultipartForm.Value {
			if isSensitiveKey(key) {
				fields[key] = "***"
			} else {
				fields[key] = values
			}
		}
		files := summary["files"].([]map[string]any)
		for field, headers := range c.Request.MultipartForm.File {
			for _, header := range headers {
				files = append(files, fileSummary(field, header))
			}
		}
		summary["files"] = files
	}
	data, _ := json.Marshal(summary)
	return string(data)
}

func fileSummary(field string, header *multipart.FileHeader) map[string]any {
	return map[string]any{"field": field, "filename": header.Filename, "size": header.Size}
}

func sanitizeBody(body []byte, contentType string) string {
	if strings.Contains(contentType, "application/json") {
		var value any
		if err := json.Unmarshal(body, &value); err == nil {
			maskSensitive(value)
			data, _ := json.Marshal(value)
			return string(data)
		}
	}
	return strings.TrimSpace(string(body))
}

func maskSensitive(value any) {
	switch item := value.(type) {
	case map[string]any:
		for key, child := range item {
			if isSensitiveKey(key) {
				item[key] = "***"
			} else {
				maskSensitive(child)
			}
		}
	case []any:
		for _, child := range item {
			maskSensitive(child)
		}
	}
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "password") || strings.Contains(key, "token") || strings.Contains(key, "secret")
}

func actionFromRequest(method string, path string) string {
	switch {
	case strings.Contains(path, "/password"):
		return "reset_password"
	case strings.Contains(path, "/auth/login"):
		return "login"
	case strings.Contains(path, "/auth/logout"):
		return "logout"
	case strings.Contains(path, "/actions"):
		return "action"
	case strings.Contains(path, "/photos") || strings.Contains(path, "/import"):
		return "upload"
	case method == http.MethodPost:
		return "create"
	case method == http.MethodPut || method == http.MethodPatch:
		return "update"
	case method == http.MethodDelete:
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func moduleFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
