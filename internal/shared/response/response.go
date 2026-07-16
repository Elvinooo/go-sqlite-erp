package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const SuccessCode = 0

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type PageData struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: SuccessCode, Message: "success", Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Body{Code: SuccessCode, Message: "success", Data: data})
}

func Fail(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Body{Code: code, Message: message, Data: nil})
}

func Page(c *gin.Context, list any, total int64, page int, pageSize int) {
	OK(c, PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}
