package util

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel 错误：仓储层与 service 层共享，上层用 errors.Is 判断。
var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrValidation      = errors.New("validation error")
	ErrRateLimited     = errors.New("rate limited")
	ErrBadRequest      = errors.New("bad request")
	ErrQuantityExceeds = errors.New("quantity exceeds stock")
)

// AppError 业务错误：携带错误码、HTTP 状态与面向用户的 message。
type AppError struct {
	Code       int
	HTTPStatus int
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// NewAppError 构造业务错误。
func NewAppError(code, status int, message string, err error) *AppError {
	return &AppError{Code: code, HTTPStatus: status, Message: message, Err: err}
}

// BadRequest 400 错误。
func BadRequest(message string, err error) *AppError {
	return NewAppError(1000, http.StatusBadRequest, message, err)
}

// UnauthorizedError 401 错误。
func UnauthorizedError(message string, err error) *AppError {
	return NewAppError(1001, http.StatusUnauthorized, message, err)
}

// ForbiddenError 403 错误。
func ForbiddenError(message string, err error) *AppError {
	return NewAppError(1002, http.StatusForbidden, message, err)
}

// NotFoundError 404 错误。
func NotFoundError(message string, err error) *AppError {
	return NewAppError(1003, http.StatusNotFound, message, err)
}

// ConflictError 409 错误。
func ConflictError(message string, err error) *AppError {
	return NewAppError(1004, http.StatusConflict, message, err)
}

// InternalError 500 错误。
func InternalError(message string, err error) *AppError {
	return NewAppError(1007, http.StatusInternalServerError, message, err)
}
