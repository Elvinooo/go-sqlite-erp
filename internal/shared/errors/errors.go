package errors

import "net/http"

type AppError struct {
	Code       int
	Message    string
	HTTPStatus int
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

var (
	ErrBadRequest        = New(40000, "请求参数错误", http.StatusBadRequest)
	ErrUnauthorized      = New(40100, "请先登录", http.StatusUnauthorized)
	ErrInvalidToken      = New(40101, "Token无效或已过期", http.StatusUnauthorized)
	ErrForbidden         = New(40300, "无操作权限", http.StatusForbidden)
	ErrNotFound          = New(40400, "数据不存在", http.StatusNotFound)
	ErrConflict          = New(40900, "数据已存在", http.StatusConflict)
	ErrInternal          = New(50000, "服务器内部错误", http.StatusInternalServerError)
	ErrInvalidCredential = New(40102, "用户名或密码错误", http.StatusUnauthorized)
	ErrDisabledUser      = New(40103, "用户已被禁用", http.StatusUnauthorized)
)
