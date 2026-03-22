package httpapi

import "net/http"

type AppError struct {
	Status  int
	Code    string
	Message string
	Details interface{}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}

func BadRequest(code, message string, details interface{}) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code, Message: message, Details: details}
}

func Unauthorized(message string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: "unauthorized", Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Status: http.StatusForbidden, Code: "forbidden", Message: message}
}

func NotFound(message string) *AppError {
	return &AppError{Status: http.StatusNotFound, Code: "not_found", Message: message}
}

func Conflict(message string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: "conflict", Message: message}
}

func Internal(message string) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: "internal", Message: message}
}
