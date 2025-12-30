package middlewares

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/pkg/logger"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// CustomHTTPErrorHandler is the custom error handler for Echo that formats errors
// in a standardized way and logs them appropriately
// This function signature matches echo.HTTPErrorHandler
func CustomHTTPErrorHandler(err error, c echo.Context) {
	// Extract request ID from context
	requestID := GetRequestID(c.Request().Context())

	// Determine status code and error message
	var statusCode int
	var message string
	var errorCode string

	// Check if it's an ownership error
	if errors.Is(err, ownership.ErrResourceNotFound) || errors.Is(err, ownership.ErrAccessDenied) {
		statusCode = http.StatusNotFound
		message = "Resource not found"
	} else if httpErr, ok := err.(*echo.HTTPError); ok {
		statusCode = httpErr.Code
		if msg, ok := httpErr.Message.(string); ok {
			message = msg
		} else {
			message = httpErr.Error()
		}
	} else {
		// Generic error, default to 500
		statusCode = http.StatusInternalServerError
		message = err.Error()
	}

	// Map status code to error code
	errorCode = getErrorCode(statusCode)

	// Get request path
	path := c.Request().URL.Path
	if path == "" {
		path = c.Path()
	}

	// Create error response
	errorResp := response.NewErrorResponse(errorCode, message, requestID, path)

	// Log the error
	log := logger.GetLogger()
	if statusCode >= 500 {
		// Server errors (5xx) - log as ERROR
		log.Error("HTTP error occurred",
			"status_code", statusCode,
			"error_code", errorCode,
			"message", message,
			"request_id", requestID,
			"path", path,
			"error", err.Error(),
		)
	} else {
		// Client errors (4xx) - log as WARN
		log.Warn("HTTP error occurred",
			"status_code", statusCode,
			"error_code", errorCode,
			"message", message,
			"request_id", requestID,
			"path", path,
		)
	}

	// Send JSON response
	if !c.Response().Committed {
		if jsonErr := c.JSON(statusCode, errorResp); jsonErr != nil {
			// If JSON encoding fails, log it but don't panic
			log.Error("Failed to encode error response", "error", jsonErr)
		}
	}
}

// getErrorCode maps HTTP status codes to error codes
func getErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "VALIDATION_ERROR"
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return "INTERNAL_SERVER_ERROR"
	default:
		if statusCode >= 500 {
			return "INTERNAL_SERVER_ERROR"
		}
		return "HTTP_ERROR"
	}
}
