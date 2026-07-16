package controller

import (
	"net/http"
	"strconv"

	dashboardapp "erp/internal/application/dashboard"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service *dashboardapp.Service
}

type dashboardSignInRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Address   string  `json:"address"`
	Device    string  `json:"device"`
}

func NewDashboardController(service *dashboardapp.Service) *DashboardController {
	return &DashboardController{service: service}
}

// Boss godoc
// @Summary 老板驾驶舱
// @Tags 老板驾驶舱
// @Security BearerAuth
// @Success 200 {object} response.Body
// @Router /api/v1/dashboard/boss [get]
func (ctl *DashboardController) Boss(c *gin.Context) {
	data, err := ctl.service.Boss(c.Request.Context(), utils.GetTenantID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, data)
}

func (ctl *DashboardController) SignInHistory(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.SignInHistory(c.Request.Context(), utils.GetTenantID(c), utils.GetUserID(c), q.Page, q.PageSize)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *DashboardController) PriceProfit(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := ctl.service.PriceProfit(c.Request.Context(), utils.GetTenantID(c), days)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, data)
}

func (ctl *DashboardController) SignIn(c *gin.Context) {
	var req dashboardSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 400, "签到位置不能为空")
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		response.Fail(c, http.StatusBadRequest, 400, "签到位置无效")
		return
	}
	if req.Device == "" {
		req.Device = c.GetHeader("User-Agent")
	}
	data, err := ctl.service.SignIn(c.Request.Context(), utils.GetTenantID(c), utils.GetUserID(c), req.Latitude, req.Longitude, req.Address, req.Device)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, data)
}
