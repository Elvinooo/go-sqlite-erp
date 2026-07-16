package controller

import (
	authapp "erp/internal/application/auth"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *authapp.Service
}

func NewAuthController(service *authapp.Service) *AuthController {
	return &AuthController{service: service}
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req authapp.LoginRequest
	if !bind(c, &req) {
		return
	}
	result, err := ctl.service.Login(c.Request.Context(), 1, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, result)
}

func (ctl *AuthController) Logout(c *gin.Context) {
	response.OK(c, gin.H{"ok": true})
}

func (ctl *AuthController) ChangePassword(c *gin.Context) {
	var req authapp.ChangePasswordRequest
	if !bind(c, &req) {
		return
	}
	if err := ctl.service.ChangePassword(c.Request.Context(), utils.GetTenantID(c), utils.GetUserID(c), req); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (ctl *AuthController) UpdateProfile(c *gin.Context) {
	var req authapp.ProfileRequest
	if !bind(c, &req) {
		return
	}
	user, err := ctl.service.UpdateProfile(c.Request.Context(), utils.GetTenantID(c), utils.GetUserID(c), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, user)
}

func (ctl *AuthController) Refresh(c *gin.Context) {
	var req authapp.RefreshTokenRequest
	if !bind(c, &req) {
		return
	}
	result, err := ctl.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, result)
}

func (ctl *AuthController) Me(c *gin.Context) {
	result, err := ctl.service.CurrentUser(c.Request.Context(), utils.GetTenantID(c), utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, result)
}
