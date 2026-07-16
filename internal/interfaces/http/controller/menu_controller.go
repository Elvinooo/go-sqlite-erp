package controller

import (
	authapp "erp/internal/application/auth"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type MenuController struct {
	service *authapp.Service
}

func NewMenuController(service *authapp.Service) *MenuController {
	return &MenuController{service: service}
}

func (ctl *MenuController) List(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.ListMenus(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *MenuController) Tree(c *gin.Context) {
	tree, err := ctl.service.MenuTree(c.Request.Context(), utils.GetTenantID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, tree)
}

func (ctl *MenuController) Create(c *gin.Context) {
	var req authapp.MenuRequest
	if !bind(c, &req) {
		return
	}
	menu, err := ctl.service.CreateMenu(c.Request.Context(), utils.GetTenantID(c), req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, menu)
}

func (ctl *MenuController) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req authapp.MenuRequest
	if !bind(c, &req) {
		return
	}
	menu, err := ctl.service.UpdateMenu(c.Request.Context(), utils.GetTenantID(c), id, req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, menu)
}

func (ctl *MenuController) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctl.service.DeleteMenu(c.Request.Context(), utils.GetTenantID(c), id); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, gin.H{"id": id})
}
