package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ErrorHandler creates a middleware for centralized error handling
func ErrorHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Forward to next handler
		err := c.Next()

		// Check if response was written
		if len(c.Response().Body()) == 0 {
			if err == nil {
				// No error but no response sent - this is a handler bug
				log.Printf("Warning: Handler didn't send any response for %s %s\n",
					c.Method(), c.Path())

				return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
					Error: "Internal server error: no response sent",
					Code:  "INTERNAL_ERROR",
				})
			}
		} else if err == nil {
			// Response was sent and no error - all good
			return nil
		}

		// Handle different types of errors
		return handleError(c, err)
	}
}

// handleError processes different error types and returns appropriate HTTP responses
func handleError(c *fiber.Ctx, err error) error {
	// Log the original error for debugging
	log.Printf("Error handling request %s %s: %v", c.Method(), c.Path(), err)

	// Handle Fiber errors first (highest priority)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(ErrorResponse{
			Error: fiberErr.Message,
			Code:  getFiberErrorCode(fiberErr.Code),
		})
	}

	// Handle gRPC status errors
	if grpcStatus, ok := status.FromError(err); ok {
		return handleGRPCError(c, grpcStatus)
	}

	// Handle context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return c.Status(fiber.StatusGatewayTimeout).JSON(ErrorResponse{
			Error: "Request timeout",
			Code:  "TIMEOUT_ERROR",
		})
	}

	if errors.Is(err, context.Canceled) {
		return c.Status(fiber.StatusRequestTimeout).JSON(ErrorResponse{
			Error: "Request was cancelled",
			Code:  "CANCELLED_ERROR",
		})
	}

	// Handle PostgreSQL database errors
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return handleDatabaseError(c, pqErr)
	}

	// Handle business logic errors based on error message patterns
	errorMsg := err.Error()
	if businessErr := handleBusinessLogicError(errorMsg); businessErr != nil {
		return c.Status(businessErr.StatusCode).JSON(ErrorResponse{
			Error:   businessErr.Message,
			Code:    businessErr.Code,
			Details: extractErrorDetails(errorMsg),
		})
	}

	// Default error response
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Error: "Internal server error",
		Code:  "INTERNAL_ERROR",
	})
}

// BusinessError represents a structured business logic error
type BusinessError struct {
	StatusCode int
	Message    string
	Code       string
}

// handleBusinessLogicError maps business logic error messages to appropriate HTTP responses
func handleBusinessLogicError(errorMsg string) *BusinessError {
	errorMsg = strings.ToLower(errorMsg)

	// Parameter validation errors
	if strings.Contains(errorMsg, "invalid parameters") {
		return &BusinessError{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Invalid request parameters",
			Code:       "INVALID_PARAMETERS",
		}
	}

	// Account-related errors
	if strings.Contains(errorMsg, "account not found") {
		return &BusinessError{
			StatusCode: fiber.StatusNotFound,
			Message:    "Account not found",
			Code:       "ACCOUNT_NOT_FOUND",
		}
	}

	if strings.Contains(errorMsg, "account validation failed") {
		return &BusinessError{
			StatusCode: fiber.StatusUnprocessableEntity,
			Message:    "Account validation failed",
			Code:       "ACCOUNT_VALIDATION_FAILED",
		}
	}

	// Balance-related errors
	if strings.Contains(errorMsg, "insufficient funds") || strings.Contains(errorMsg, "insufficient balance") {
		return &BusinessError{
			StatusCode: fiber.StatusUnprocessableEntity,
			Message:    "Insufficient funds",
			Code:       "INSUFFICIENT_FUNDS",
		}
	}

	// Currency-related errors
	if strings.Contains(errorMsg, "invalid currency") || strings.Contains(errorMsg, "unsupported currency") {
		return &BusinessError{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Invalid or unsupported currency",
			Code:       "INVALID_CURRENCY",
		}
	}

	// Transaction-related errors
	if strings.Contains(errorMsg, "transaction failed") {
		return &BusinessError{
			StatusCode: fiber.StatusUnprocessableEntity,
			Message:    "Transaction processing failed",
			Code:       "TRANSACTION_FAILED",
		}
	}

	if strings.Contains(errorMsg, "duplicate transaction") || strings.Contains(errorMsg, "already exists") {
		return &BusinessError{
			StatusCode: fiber.StatusConflict,
			Message:    "Duplicate transaction",
			Code:       "DUPLICATE_TRANSACTION",
		}
	}

	// Workflow-related errors
	if strings.Contains(errorMsg, "workflow not found") {
		return &BusinessError{
			StatusCode: fiber.StatusNotFound,
			Message:    "Transfer not found",
			Code:       "TRANSFER_NOT_FOUND",
		}
	}

	if strings.Contains(errorMsg, "workflow failed") {
		return &BusinessError{
			StatusCode: fiber.StatusUnprocessableEntity,
			Message:    "Transfer processing failed",
			Code:       "TRANSFER_FAILED",
		}
	}

	// Authorization/Permission errors
	if strings.Contains(errorMsg, "unauthorized") || strings.Contains(errorMsg, "access denied") {
		return &BusinessError{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Unauthorized access",
			Code:       "UNAUTHORIZED",
		}
	}

	if strings.Contains(errorMsg, "forbidden") {
		return &BusinessError{
			StatusCode: fiber.StatusForbidden,
			Message:    "Access forbidden",
			Code:       "FORBIDDEN",
		}
	}

	return nil
}

// handleGRPCError maps gRPC status codes to HTTP responses
func handleGRPCError(c *fiber.Ctx, grpcStatus *status.Status) error {
	var statusCode int
	var errorCode string
	var message string

	switch grpcStatus.Code() {
	case codes.InvalidArgument:
		statusCode = fiber.StatusBadRequest
		errorCode = "INVALID_ARGUMENT"
		message = "Invalid request parameters"
	case codes.NotFound:
		statusCode = fiber.StatusNotFound
		errorCode = "NOT_FOUND"
		message = "Resource not found"
	case codes.AlreadyExists:
		statusCode = fiber.StatusConflict
		errorCode = "ALREADY_EXISTS"
		message = "Resource already exists"
	case codes.PermissionDenied:
		statusCode = fiber.StatusForbidden
		errorCode = "PERMISSION_DENIED"
		message = "Permission denied"
	case codes.Unauthenticated:
		statusCode = fiber.StatusUnauthorized
		errorCode = "UNAUTHENTICATED"
		message = "Authentication required"
	case codes.ResourceExhausted:
		statusCode = fiber.StatusTooManyRequests
		errorCode = "RATE_LIMITED"
		message = "Rate limit exceeded"
	case codes.FailedPrecondition:
		statusCode = fiber.StatusPreconditionFailed
		errorCode = "PRECONDITION_FAILED"
		message = "Precondition failed"
	case codes.OutOfRange:
		statusCode = fiber.StatusBadRequest
		errorCode = "OUT_OF_RANGE"
		message = "Parameter out of range"
	case codes.Unimplemented:
		statusCode = fiber.StatusNotImplemented
		errorCode = "NOT_IMPLEMENTED"
		message = "Feature not implemented"
	case codes.Unavailable:
		statusCode = fiber.StatusServiceUnavailable
		errorCode = "SERVICE_UNAVAILABLE"
		message = "Service temporarily unavailable"
	case codes.DeadlineExceeded:
		statusCode = fiber.StatusGatewayTimeout
		errorCode = "TIMEOUT"
		message = "Request timeout"
	case codes.Canceled:
		statusCode = fiber.StatusRequestTimeout
		errorCode = "CANCELLED"
		message = "Request was cancelled"
	case codes.DataLoss:
		statusCode = fiber.StatusInternalServerError
		errorCode = "DATA_LOSS"
		message = "Data loss detected"
	case codes.Unknown:
		statusCode = fiber.StatusInternalServerError
		errorCode = "UNKNOWN_ERROR"
		message = "Unknown error occurred"
	case codes.Internal:
		fallthrough
	default:
		statusCode = fiber.StatusInternalServerError
		errorCode = "INTERNAL_ERROR"
		message = "Internal server error"
	}

	// Use gRPC message if it's more descriptive than our default
	if grpcStatus.Message() != "" && len(grpcStatus.Message()) > len(message) {
		return c.Status(statusCode).JSON(ErrorResponse{
			Error:   message,
			Code:    errorCode,
			Details: grpcStatus.Message(),
		})
	}

	return c.Status(statusCode).JSON(ErrorResponse{
		Error: message,
		Code:  errorCode,
	})
}

// handleDatabaseError maps PostgreSQL errors to HTTP responses
func handleDatabaseError(c *fiber.Ctx, pqErr *pq.Error) error {
	var statusCode int
	var message string
	var errorCode string

	switch pqErr.Code {
	case "23505": // unique_violation
		statusCode = fiber.StatusConflict
		message = "Resource already exists"
		errorCode = "DUPLICATE_RESOURCE"
	case "23503": // foreign_key_violation
		statusCode = fiber.StatusBadRequest
		message = "Referenced resource does not exist"
		errorCode = "INVALID_REFERENCE"
	case "23502": // not_null_violation
		statusCode = fiber.StatusBadRequest
		message = "Required field is missing"
		errorCode = "MISSING_REQUIRED_FIELD"
	case "23514": // check_violation
		statusCode = fiber.StatusBadRequest
		message = "Data constraint violation"
		errorCode = "CONSTRAINT_VIOLATION"
	case "42P01": // undefined_table
		statusCode = fiber.StatusInternalServerError
		message = "Database configuration error"
		errorCode = "DATABASE_ERROR"
	case "42703": // undefined_column
		statusCode = fiber.StatusInternalServerError
		message = "Database schema error"
		errorCode = "DATABASE_ERROR"
	case "08006": // connection_failure
		statusCode = fiber.StatusServiceUnavailable
		message = "Database connection failed"
		errorCode = "DATABASE_UNAVAILABLE"
	case "57014": // query_canceled
		statusCode = fiber.StatusRequestTimeout
		message = "Database query timeout"
		errorCode = "QUERY_TIMEOUT"
	default:
		statusCode = fiber.StatusInternalServerError
		message = "Database error"
		errorCode = "DATABASE_ERROR"
	}

	return c.Status(statusCode).JSON(ErrorResponse{
		Error:   message,
		Code:    errorCode,
		Details: fmt.Sprintf("Database error: %s", pqErr.Code),
	})
}

// getFiberErrorCode maps Fiber HTTP status codes to error codes
func getFiberErrorCode(statusCode int) string {
	switch statusCode {
	case fiber.StatusBadRequest:
		return "BAD_REQUEST"
	case fiber.StatusUnauthorized:
		return "UNAUTHORIZED"
	case fiber.StatusForbidden:
		return "FORBIDDEN"
	case fiber.StatusNotFound:
		return "NOT_FOUND"
	case fiber.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case fiber.StatusRequestTimeout:
		return "REQUEST_TIMEOUT"
	case fiber.StatusConflict:
		return "CONFLICT"
	case fiber.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case fiber.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case fiber.StatusInternalServerError:
		return "INTERNAL_ERROR"
	case fiber.StatusNotImplemented:
		return "NOT_IMPLEMENTED"
	case fiber.StatusBadGateway:
		return "BAD_GATEWAY"
	case fiber.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case fiber.StatusGatewayTimeout:
		return "GATEWAY_TIMEOUT"
	default:
		return "UNKNOWN_ERROR"
	}
}

// extractErrorDetails extracts technical details from error messages while keeping them safe for client consumption
func extractErrorDetails(errorMsg string) string {
	// Remove sensitive information but keep useful debugging info
	details := errorMsg

	// Remove file paths and line numbers
	if idx := strings.Index(details, ".go:"); idx != -1 {
		if start := strings.LastIndex(details[:idx], "/"); start != -1 {
			details = details[:start] + details[idx:]
		}
	}

	// Truncate very long error messages
	if len(details) > 200 {
		details = details[:200] + "..."
	}

	return details
}
