package middleware

import (
	"net/http"

	apperrors "erp/internal/shared/errors"
	"erp/internal/shared/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("panic recovered", zap.Any("panic", recovered))
		response.Fail(c, http.StatusInternalServerError, apperrors.ErrInternal.Code, apperrors.ErrInternal.Message)
		c.Abort()
	})
}
