package api

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents API error codes
type ErrorCode string

const (
	ErrInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrForbidden         ErrorCode = "FORBIDDEN"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrConflict          ErrorCode = "CONFLICT"
	ErrRateLimited       ErrorCode = "RATE_LIMITED"
	ErrJobFailed         ErrorCode = "JOB_FAILED"
	ErrModelUnavailable  ErrorCode = "MODEL_UNAVAILABLE"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrInternalError     ErrorCode = "INTERNAL_ERROR"
)

// Response represents a successful API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination metadata
type Meta struct {
	Cursor string `json:"cursor,omitempty"`
	Total  int    `json:"total,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Success writes a successful response
func Success(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	JSON(w, status, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Error writes an error response
func Error(w http.ResponseWriter, status int, code ErrorCode, message string, details map[string]interface{}) {
	JSON(w, status, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// BadRequest writes a 400 error
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, ErrInvalidRequest, message, nil)
}

// NotFound writes a 404 error
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, ErrNotFound, message, nil)
}

// Conflict writes a 409 error
func Conflict(w http.ResponseWriter, message string, details map[string]interface{}) {
	Error(w, http.StatusConflict, ErrConflict, message, details)
}

// InternalError writes a 500 error
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, ErrInternalError, message, nil)
}

// ServiceUnavailable writes a 503 error for unavailable services
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, ErrServiceUnavailable, message, nil)
}
