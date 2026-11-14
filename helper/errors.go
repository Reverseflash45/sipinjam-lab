package helper

import (
	"fmt"
	"net/http"
)

// AppError adalah kegagalan yang sudah direncanakan: punya status HTTP,
// kode yang stabil untuk frontend, dan pesan untuk manusia. cause hanya
// dicatat di log dan tidak pernah dikirim ke client.
type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	cause   error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.cause)
	}
	return e.Code + ": " + e.Message
}

func (e *AppError) Unwrap() error { return e.cause }

func NewError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func BadRequest(msg string) *AppError { return NewError(http.StatusBadRequest, "BAD_REQUEST", msg) }
func Unauthorized(msg string) *AppError {
	return NewError(http.StatusUnauthorized, "UNAUTHORIZED", msg)
}
func Forbidden(msg string) *AppError { return NewError(http.StatusForbidden, "FORBIDDEN", msg) }
func NotFound(msg string) *AppError  { return NewError(http.StatusNotFound, "NOT_FOUND", msg) }
func Conflict(code, msg string) *AppError {
	return NewError(http.StatusConflict, code, msg)
}

func Validation(fields map[string]string) *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Code: "VALIDATION_ERROR",
		Message: "data tidak valid", Fields: fields}
}

func Internal(cause error) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR",
		Message: "terjadi kesalahan pada server", cause: cause}
}
