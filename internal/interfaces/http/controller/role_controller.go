package controller

import (
	authapp "erp/internal/application/auth"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type RoleController struct {
	service *authapp.Service
}

func NewRoleController(service *authapp.Service) *RoleController {
	return &RoleController{service: service}
}

func (ctl *RoleController) List(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.ListRoles(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *RoleController) Create(c *gin.Context) {
	var req authapp.RoleRequest
	if !bind(c, &req) {
		return
	}
	role, err := ctl.service.CreateRole(c.Request.Context(), utils.GetTenantID(c), req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, role)
}

func (ctl *RoleController) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req authapp.RoleRequest
	if !bind(c, &req) {
		return
	}
	role, err := ctl.service.UpdateRole(c.Request.Context(), utils.GetTenantID(c), id, req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, role)
}

func (ctl *RoleController) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctl.service.DeleteRole(c.Request.Context(), utils.GetTenantID(c), id); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, gin.H{"id": id})
}
