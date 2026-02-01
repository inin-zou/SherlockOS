package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "hello"}

	JSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("JSON() status = %v, want %v", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("JSON() Content-Type = %v, want application/json", ct)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("JSON() body decode error: %v", err)
	}

	if result["message"] != "hello" {
		t.Errorf("JSON() body message = %v, want hello", result["message"])
	}
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123"}
	meta := &Meta{Total: 10}

	Success(w, http.StatusOK, data, meta)

	if w.Code != http.StatusOK {
		t.Errorf("Success() status = %v, want %v", w.Code, http.StatusOK)
	}

	var result Response
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Success() body decode error: %v", err)
	}

	if !result.Success {
		t.Error("Success() should set success=true")
	}

	if result.Meta == nil {
		t.Error("Success() should include meta")
	}

	if result.Meta.Total != 10 {
		t.Errorf("Success() meta.total = %v, want 10", result.Meta.Total)
	}
}

func TestSuccess_WithoutMeta(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123"}

	Success(w, http.StatusCreated, data, nil)

	if w.Code != http.StatusCreated {
		t.Errorf("Success() status = %v, want %v", w.Code, http.StatusCreated)
	}

	var result Response
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Success() body decode error: %v", err)
	}

	if !result.Success {
		t.Error("Success() should set success=true")
	}

	if result.Meta != nil {
		t.Error("Success() meta should be nil when not provided")
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	details := map[string]interface{}{"field": "title"}

	Error(w, http.StatusBadRequest, ErrInvalidRequest, "Title is required", details)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Error() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	var result ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Error() body decode error: %v", err)
	}

	if result.Success {
		t.Error("Error() should set success=false")
	}

	if result.Error.Code != ErrInvalidRequest {
		t.Errorf("Error() code = %v, want %v", result.Error.Code, ErrInvalidRequest)
	}

	if result.Error.Message != "Title is required" {
		t.Errorf("Error() message = %v", result.Error.Message)
	}

	if result.Error.Details["field"] != "title" {
		t.Errorf("Error() details.field = %v, want title", result.Error.Details["field"])
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()

	BadRequest(w, "Invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("BadRequest() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	var result ErrorResponse
	json.NewDecoder(w.Body).Decode(&result)

	if result.Error.Code != ErrInvalidRequest {
		t.Errorf("BadRequest() code = %v, want %v", result.Error.Code, ErrInvalidRequest)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()

	NotFound(w, "Case not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("NotFound() status = %v, want %v", w.Code, http.StatusNotFound)
	}

	var result ErrorResponse
	json.NewDecoder(w.Body).Decode(&result)

	if result.Error.Code != ErrNotFound {
		t.Errorf("NotFound() code = %v, want %v", result.Error.Code, ErrNotFound)
	}
}

func TestConflict(t *testing.T) {
	w := httptest.NewRecorder()
	details := map[string]interface{}{"existing_job_id": "job_123"}

	Conflict(w, "Job already exists", details)

	if w.Code != http.StatusConflict {
		t.Errorf("Conflict() status = %v, want %v", w.Code, http.StatusConflict)
	}

	var result ErrorResponse
	json.NewDecoder(w.Body).Decode(&result)

	if result.Error.Code != ErrConflict {
		t.Errorf("Conflict() code = %v, want %v", result.Error.Code, ErrConflict)
	}

	if result.Error.Details["existing_job_id"] != "job_123" {
		t.Errorf("Conflict() details.existing_job_id = %v", result.Error.Details["existing_job_id"])
	}
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()

	InternalError(w, "Database connection failed")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("InternalError() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}

	var result ErrorResponse
	json.NewDecoder(w.Body).Decode(&result)

	if result.Error.Code != ErrInternalError {
		t.Errorf("InternalError() code = %v, want %v", result.Error.Code, ErrInternalError)
	}
}

func TestErrorCode_Values(t *testing.T) {
	// Verify error code string values match expected format
	codes := map[ErrorCode]string{
		ErrInvalidRequest:   "INVALID_REQUEST",
		ErrUnauthorized:     "UNAUTHORIZED",
		ErrForbidden:        "FORBIDDEN",
		ErrNotFound:         "NOT_FOUND",
		ErrConflict:         "CONFLICT",
		ErrRateLimited:      "RATE_LIMITED",
		ErrJobFailed:        "JOB_FAILED",
		ErrModelUnavailable: "MODEL_UNAVAILABLE",
		ErrInternalError:    "INTERNAL_ERROR",
	}

	for code, expected := range codes {
		if string(code) != expected {
			t.Errorf("ErrorCode %v = %v, want %v", code, string(code), expected)
		}
	}
}
