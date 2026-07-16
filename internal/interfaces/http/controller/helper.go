package controller

import (
	"strconv"

	apperrors "erp/internal/shared/errors"
	"github.com/gin-gonic/gin"
)

func bind(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest)
		return false
	}
	return true
}

func idParam(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		_ = c.Error(apperrors.ErrBadRequest)
		return 0, false
	}
	return id, true
}
