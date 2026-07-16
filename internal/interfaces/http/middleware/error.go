package middleware

import (
	stderrors "errors"
	"net/http"

	apperrors "erp/internal/shared/errors"
	"erp/internal/shared/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func ErrorHandler(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *apperrors.AppError
		if stderrors.As(err, &appErr) {
			response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, apperrors.ErrNotFound.Code, apperrors.ErrNotFound.Message)
			return
		}
		log.Error("request failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, apperrors.ErrInternal.Code, apperrors.ErrInternal.Message)
	}
}
