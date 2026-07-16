package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type PageQuery struct {
	Page         int
	PageSize     int
	Keyword      string
	SortBy       string
	Order        string
	Module       string
	Action       string
	Method       string
	StatusCode   string
	StartDate    string
	EndDate      string
	ProductCode  string
	ProductName  string
	CustomerID   uint64
	SourceType   string
	OperatorName string
	UserID       uint64
	Username     string
	DeletedOnly  bool
}

func FromGin(c *gin.Context) PageQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	order := c.DefaultQuery("order", "desc")
	if order != "asc" {
		order = "desc"
	}
	deletedOnly, _ := strconv.ParseBool(c.DefaultQuery("deletedOnly", "false"))
	customerID, _ := strconv.ParseUint(c.Query("customerId"), 10, 64)
	userID, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	return PageQuery{
		Page:         page,
		PageSize:     pageSize,
		Keyword:      c.Query("keyword"),
		SortBy:       c.DefaultQuery("sortBy", "id"),
		Order:        order,
		Module:       c.Query("module"),
		Action:       c.Query("action"),
		Method:       c.Query("method"),
		StatusCode:   c.Query("statusCode"),
		StartDate:    c.Query("startDate"),
		EndDate:      c.Query("endDate"),
		ProductCode:  c.Query("productCode"),
		ProductName:  c.Query("productName"),
		CustomerID:   customerID,
		SourceType:   c.Query("sourceType"),
		OperatorName: c.Query("operatorName"),
		UserID:       userID,
		Username:     c.Query("username"),
		DeletedOnly:  deletedOnly,
	}
}

func (q PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}
