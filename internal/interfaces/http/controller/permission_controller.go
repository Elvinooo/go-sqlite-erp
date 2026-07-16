package controller

import (
	authapp "erp/internal/application/auth"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	service *authapp.Service
}

func NewPermissionController(service *authapp.Service) *PermissionController {
	return &PermissionController{service: service}
}

func (ctl *PermissionController) List(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.ListPermissions(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *PermissionController) Create(c *gin.Context) {
	var req authapp.PermissionRequest
	if !bind(c, &req) {
		return
	}
	permission, err := ctl.service.CreatePermission(c.Request.Context(), utils.GetTenantID(c), req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, permission)
}

func (ctl *PermissionController) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req authapp.PermissionRequest
	if !bind(c, &req) {
		return
	}
	permission, err := ctl.service.UpdatePermission(c.Request.Context(), utils.GetTenantID(c), id, req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, permission)
}

func (ctl *PermissionController) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctl.service.DeletePermission(c.Request.Context(), utils.GetTenantID(c), id); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, gin.H{"id": id})
}
