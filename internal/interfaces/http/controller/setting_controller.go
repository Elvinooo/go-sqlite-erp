package controller

import (
	systemapp "erp/internal/application/system"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type SettingController struct {
	service *systemapp.Service
}

func NewSettingController(service *systemapp.Service) *SettingController {
	return &SettingController{service: service}
}

func (ctl *SettingController) List(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.ListSettings(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *SettingController) Create(c *gin.Context) {
	var req systemapp.SettingRequest
	if !bind(c, &req) {
		return
	}
	setting, err := ctl.service.CreateSetting(c.Request.Context(), utils.GetTenantID(c), req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, setting)
}

func (ctl *SettingController) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req systemapp.SettingRequest
	if !bind(c, &req) {
		return
	}
	setting, err := ctl.service.UpdateSetting(c.Request.Context(), utils.GetTenantID(c), id, req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, setting)
}

func (ctl *SettingController) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctl.service.DeleteSetting(c.Request.Context(), utils.GetTenantID(c), id); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (ctl *SettingController) RestoreTestData(c *gin.Context) {
	c.Set("skipOperationAudit", true)
	result, err := ctl.service.RestoreTestData(c.Request.Context(), utils.GetTenantID(c), utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, result)
}

func (ctl *SettingController) MerchantInfo(c *gin.Context) {
	result, err := ctl.service.MerchantInfo(c.Request.Context(), utils.GetTenantID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, result)
}

func (ctl *SettingController) SaveMerchantInfo(c *gin.Context) {
	var req systemapp.MerchantInfoRequest
	if !bind(c, &req) {
		return
	}
	result, err := ctl.service.SaveMerchantInfo(c.Request.Context(), utils.GetTenantID(c), utils.GetUserID(c), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, result)
}
