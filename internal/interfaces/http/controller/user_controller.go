package controller

import (
	authapp "erp/internal/application/auth"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	service *authapp.Service
}

func NewUserController(service *authapp.Service) *UserController {
	return &UserController{service: service}
}

func (ctl *UserController) List(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.ListUsers(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *UserController) Create(c *gin.Context) {
	var req authapp.UserCreateRequest
	if !bind(c, &req) {
		return
	}
	user, err := ctl.service.CreateUser(c.Request.Context(), utils.GetTenantID(c), req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, user)
}

func (ctl *UserController) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req authapp.UserUpdateRequest
	if !bind(c, &req) {
		return
	}
	user, err := ctl.service.UpdateUser(c.Request.Context(), utils.GetTenantID(c), id, req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, user)
}

func (ctl *UserController) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctl.service.DeleteUser(c.Request.Context(), utils.GetTenantID(c), id); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (ctl *UserController) ResetPassword(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req authapp.ResetPasswordRequest
	if !bind(c, &req) {
		return
	}
	if err := ctl.service.ResetPassword(c.Request.Context(), utils.GetTenantID(c), id, req); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, gin.H{"id": id})
}
