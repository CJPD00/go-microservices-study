package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error codes
const (
	CodeValidation   = "VALIDATION_ERROR"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeInternal     = "INTERNAL_ERROR"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
)

// AppError represents an application error
type AppError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Err     error       `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// ErrorResponse is the JSON response structure for errors
type ErrorResponse struct {
	Error   ErrorBody `json:"error"`
	TraceID string    `json:"trace_id,omitempty"`
}

// ErrorBody contains error details
type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ToJSON converts an error to the standard JSON response
func ToJSON(err error, traceID string) (int, []byte) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = &AppError{
			Code:    CodeInternal,
			Message: "An internal error occurred",
		}
	}

	response := ErrorResponse{
		Error: ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		TraceID: traceID,
	}

	data, _ := json.Marshal(response)
	return HTTPStatus(appErr), data
}

// HTTPStatus returns the HTTP status code for an error
func HTTPStatus(err error) int {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return http.StatusInternalServerError
	}

	switch appErr.Code {
	case CodeValidation:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// GRPCStatus converts an error to a gRPC status
func GRPCStatus(err error) error {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return status.Error(codes.Internal, "internal error")
	}

	var code codes.Code
	switch appErr.Code {
	case CodeValidation:
		code = codes.InvalidArgument
	case CodeNotFound:
		code = codes.NotFound
	case CodeConflict:
		code = codes.AlreadyExists
	case CodeUnauthorized:
		code = codes.Unauthenticated
	case CodeForbidden:
		code = codes.PermissionDenied
	default:
		code = codes.Internal
	}

	return status.Error(code, appErr.Message)
}

// FromGRPCStatus converts a gRPC status to an AppError
func FromGRPCStatus(err error) *AppError {
	st, ok := status.FromError(err)
	if !ok {
		return NewInternal("unknown error", err)
	}

	var code string
	switch st.Code() {
	case codes.InvalidArgument:
		code = CodeValidation
	case codes.NotFound:
		code = CodeNotFound
	case codes.AlreadyExists:
		code = CodeConflict
	case codes.Unauthenticated:
		code = CodeUnauthorized
	case codes.PermissionDenied:
		code = CodeForbidden
	default:
		code = CodeInternal
	}

	return &AppError{
		Code:    code,
		Message: st.Message(),
		Err:     err,
	}
}

// Constructor functions

// NewValidation creates a validation error
func NewValidation(message string, details interface{}) *AppError {
	return &AppError{
		Code:    CodeValidation,
		Message: message,
		Details: details,
	}
}

// NewNotFound creates a not found error
func NewNotFound(resource string, id interface{}) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s with id '%v' not found", resource, id),
	}
}

// NewConflict creates a conflict error
func NewConflict(message string) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: message,
	}
}

// NewInternal creates an internal error
func NewInternal(message string, err error) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: message,
		Err:     err,
	}
}

// NewUnauthorized creates an unauthorized error
func NewUnauthorized(message string) *AppError {
	return &AppError{
		Code:    CodeUnauthorized,
		Message: message,
	}
}

// Is checks if an error matches a specific code
func Is(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return &AppError{
			Code:    appErr.Code,
			Message: message + ": " + appErr.Message,
			Details: appErr.Details,
			Err:     err,
		}
	}
	return NewInternal(message, err)
}
