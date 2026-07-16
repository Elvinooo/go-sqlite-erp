package controller

import (
	authapp "erp/internal/application/auth"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type AuditController struct {
	service *authapp.Service
}

func NewAuditController(service *authapp.Service) *AuditController {
	return &AuditController{service: service}
}

func (ctl *AuditController) LoginLogs(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.ListLoginLogs(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *AuditController) OperationLogs(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.ListOperationLogs(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}
