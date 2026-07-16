package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	customerapp "erp/internal/application/customer"
	apperrors "erp/internal/shared/errors"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type CustomerController struct {
	service *customerapp.Service
}

func NewCustomerController(service *customerapp.Service) *CustomerController {
	return &CustomerController{service: service}
}

// List godoc
// @Summary 客户分页列表
// @Tags 客户管理
// @Security BearerAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "模糊搜索：编码/名称/电话/税号/地址"
// @Success 200 {object} response.Body
// @Router /api/v1/customers [get]
func (ctl *CustomerController) List(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.List(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

// Get godoc
// @Summary 客户详情
// @Tags 客户管理
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Success 200 {object} response.Body
// @Router /api/v1/customers/{id} [get]
func (ctl *CustomerController) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	customer, err := ctl.service.Get(c.Request.Context(), utils.GetTenantID(c), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, customer)
}

// Create godoc
// @Summary 创建客户
// @Tags 客户管理
// @Security BearerAuth
// @Accept json
// @Param body body customer.CustomerRequest true "客户"
// @Success 201 {object} response.Body
// @Router /api/v1/customers [post]
func (ctl *CustomerController) Create(c *gin.Context) {
	var req customerapp.CustomerRequest
	if !bind(c, &req) {
		return
	}
	customer, err := ctl.service.Create(c.Request.Context(), utils.GetTenantID(c), req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, customer)
}

// Update godoc
// @Summary 更新客户
// @Tags 客户管理
// @Security BearerAuth
// @Accept json
// @Param id path int true "客户ID"
// @Param body body customer.CustomerRequest true "客户"
// @Success 200 {object} response.Body
// @Router /api/v1/customers/{id} [put]
func (ctl *CustomerController) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req customerapp.CustomerRequest
	if !bind(c, &req) {
		return
	}
	customer, err := ctl.service.Update(c.Request.Context(), utils.GetTenantID(c), id, req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, customer)
}

// Delete godoc
// @Summary 删除客户
// @Tags 客户管理
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Success 200 {object} response.Body
// @Router /api/v1/customers/{id} [delete]
func (ctl *CustomerController) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctl.service.Delete(c.Request.Context(), utils.GetTenantID(c), id); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ImportExcel godoc
// @Summary 导入客户Excel
// @Tags 客户管理
// @Security BearerAuth
// @Accept multipart/form-data
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Body
// @Router /api/v1/customers/import [post]
func (ctl *CustomerController) ImportExcel(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest)
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		_ = c.Error(err)
		return
	}
	defer file.Close()
	result, err := ctl.service.ImportExcel(c.Request.Context(), utils.GetTenantID(c), utils.GetUserID(c), file)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, result)
}

// ExportExcel godoc
// @Summary 导出客户Excel
// @Tags 客户管理
// @Security BearerAuth
// @Param keyword query string false "模糊搜索"
// @Success 200 {file} file
// @Router /api/v1/customers/export [get]
func (ctl *CustomerController) ExportExcel(c *gin.Context) {
	q := query.FromGin(c)
	body, err := ctl.service.ExportExcel(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	fileName := fmt.Sprintf("customers_%s.xlsx", time.Now().Format("20060102150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", body)
}

// Debt godoc
// @Summary 客户欠款
// @Tags 客户管理
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Success 200 {object} response.Body
// @Router /api/v1/customers/{id}/debt [get]
func (ctl *CustomerController) Debt(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	debt, err := ctl.service.Debt(c.Request.Context(), utils.GetTenantID(c), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, debt)
}

// OrderHistory godoc
// @Summary 客户历史订单
// @Tags 客户管理
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Body
// @Router /api/v1/customers/{id}/orders [get]
func (ctl *CustomerController) OrderHistory(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	q := query.FromGin(c)
	list, total, err := ctl.service.OrderHistory(c.Request.Context(), utils.GetTenantID(c), id, q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}
