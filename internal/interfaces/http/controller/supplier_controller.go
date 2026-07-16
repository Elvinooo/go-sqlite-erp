package controller

import (
	supplierapp "erp/internal/application/supplier"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type SupplierController struct {
	service *supplierapp.Service
}

func NewSupplierController(service *supplierapp.Service) *SupplierController {
	return &SupplierController{service: service}
}

func (ctl *SupplierController) List(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.List(c.Request.Context(), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *SupplierController) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	item, err := ctl.service.Get(c.Request.Context(), utils.GetTenantID(c), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, item)
}

func (ctl *SupplierController) Create(c *gin.Context) {
	var req supplierapp.SupplierRequest
	if !bind(c, &req) {
		return
	}
	item, err := ctl.service.Create(c.Request.Context(), utils.GetTenantID(c), req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, item)
}

func (ctl *SupplierController) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req supplierapp.SupplierRequest
	if !bind(c, &req) {
		return
	}
	item, err := ctl.service.Update(c.Request.Context(), utils.GetTenantID(c), id, req, utils.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, item)
}

func (ctl *SupplierController) Delete(c *gin.Context) {
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
