package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRequestIDMiddleware_GeneratesRequestID(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify that a request ID was generated and added to response header
	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected request ID to be generated")
	}
	if len(requestID) != 32 {
		t.Errorf("Expected request ID to be 32 chars (16 bytes = 32 hex chars), got %d", len(requestID))
	}
}

func TestRequestIDMiddleware_UsesExistingRequestID(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	existingID := "existing-request-id-12345"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify that the existing request ID was used
	requestID := rec.Header().Get("X-Request-ID")
	if requestID != existingID {
		t.Errorf("Expected existing request ID to be used, got %s", requestID)
	}
}

func TestRequestIDMiddleware_AvailableInContext(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware())

	var capturedRequestID string
	e.GET("/test", func(c echo.Context) error {
		// Extract request ID from context using GetRequestID()
		capturedRequestID = GetRequestID(c.Request().Context())
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify that request ID is available in context
	if capturedRequestID == "" {
		t.Error("Expected request ID to be available in context")
	}
	
	// Verify it matches the response header
	headerRequestID := rec.Header().Get("X-Request-ID")
	if headerRequestID != capturedRequestID {
		t.Errorf("Request ID in context should match response header. Context: %s, Header: %s", capturedRequestID, headerRequestID)
	}
}

func TestRequestIDMiddleware_ConsistentAcrossRequest(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware())

	var requestID1, requestID2 string
	e.GET("/test", func(c echo.Context) error {
		// Get request ID twice from context
		requestID1 = GetRequestID(c.Request().Context())
		requestID2 = GetRequestID(c.Request().Context())
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify that request ID is consistent
	if requestID1 != requestID2 {
		t.Errorf("Request ID should be consistent when retrieved multiple times. First: %s, Second: %s", requestID1, requestID2)
	}
	
	// Verify it matches the response header
	headerRequestID := rec.Header().Get("X-Request-ID")
	if headerRequestID != requestID1 {
		t.Errorf("Request ID in context should match response header. Context: %s, Header: %s", requestID1, headerRequestID)
	}
}

func TestRequestIDMiddleware_WorksWithExistingRequestID(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware())

	var capturedRequestID string
	e.GET("/test", func(c echo.Context) error {
		capturedRequestID = GetRequestID(c.Request().Context())
		return c.String(http.StatusOK, "OK")
	})

	existingID := "custom-request-id-from-client"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify that the existing request ID is available in context
	if capturedRequestID != existingID {
		t.Errorf("Existing request ID should be available in context. Expected: %s, Got: %s", existingID, capturedRequestID)
	}
	
	// Verify it's also in the response header
	headerRequestID := rec.Header().Get("X-Request-ID")
	if headerRequestID != existingID {
		t.Errorf("Existing request ID should be in response header. Expected: %s, Got: %s", existingID, headerRequestID)
	}
}

func TestRequestIDMiddleware_MultipleRequestsHaveDifferentIDs(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	// Verify that different requests get different IDs
	requestID1 := rec1.Header().Get("X-Request-ID")
	requestID2 := rec2.Header().Get("X-Request-ID")
	
	if requestID1 == "" {
		t.Error("First request should have a request ID")
	}
	if requestID2 == "" {
		t.Error("Second request should have a request ID")
	}
	if requestID1 == requestID2 {
		t.Errorf("Different requests should have different request IDs. Both got: %s", requestID1)
	}
}

func TestRequestIDMiddleware_ContextWithoutRequestID(t *testing.T) {
	// Test GetRequestID with a context that doesn't have a request ID
	ctx := context.Background()
	requestID := GetRequestID(ctx)
	if requestID != "" {
		t.Errorf("Context without request ID should return empty string, got: %s", requestID)
	}
}
