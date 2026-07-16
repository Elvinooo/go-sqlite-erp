package middleware

import (
	"strings"

	authapp "erp/internal/application/auth"
	"erp/internal/infrastructure/security"
	apperrors "erp/internal/shared/errors"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

func JWTAuth(jwt *security.JWTManager, authService *authapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			_ = c.Error(apperrors.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil || claims.TokenUse != "access" {
			_ = c.Error(apperrors.ErrInvalidToken)
			c.Abort()
			return
		}
		if err := authService.ValidateAccessToken(c.Request.Context(), claims); err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set(utils.ContextUserID, claims.UserID)
		c.Set(utils.ContextTenantID, claims.TenantID)
		c.Set(utils.ContextUsername, claims.Username)
		c.Next()
	}
}

func RBAC(authService *authapp.Service, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := utils.GetUserID(c)
		if userID == 0 || !authService.HasPermission(c.Request.Context(), userID, permissionCode) {
			_ = c.Error(apperrors.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func DataScope(authService *authapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := utils.GetUserID(c)
		if userID == 0 {
			c.Next()
			return
		}
		scopes, err := authService.DataScopes(c.Request.Context(), userID)
		if err == nil {
			c.Set(utils.ContextDataScopes, scopes)
		}
		c.Next()
	}
}
