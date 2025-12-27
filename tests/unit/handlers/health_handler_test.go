package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/handlers"
)

// mockHealthService is a mock implementation of IHealthService
type mockHealthService struct {
	healthResp *response.HealthResponse
	err        error
}

func (m *mockHealthService) CheckHealth(ctx context.Context) (*response.HealthResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.healthResp, nil
}

func TestHealthHandler_HealthCheck_AllHealthy(t *testing.T) {
	healthResp := response.NewHealthResponse()
	healthResp.SetComponent("database", "healthy", "Connection successful")
	healthResp.SetComponent("redis", "healthy", "Connection successful")
	healthResp.CalculateOverallStatus()

	mockService := &mockHealthService{healthResp: healthResp}
	handler := handlers.NewHealthHandler(mockService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HealthCheck(c)

	if err != nil {
		t.Fatalf("HealthCheck() error = %v, want nil", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("HealthCheck() status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var result response.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("HealthCheck() failed to unmarshal response: %v", err)
	}

	if result.Status != "healthy" {
		t.Errorf("HealthCheck() status = %v, want 'healthy'", result.Status)
	}
}

func TestHealthHandler_HealthCheck_Degraded(t *testing.T) {
	healthResp := response.NewHealthResponse()
	healthResp.SetComponent("database", "unhealthy", "Connection failed")
	healthResp.SetComponent("redis", "healthy", "Connection successful")
	healthResp.CalculateOverallStatus()

	mockService := &mockHealthService{healthResp: healthResp}
	handler := handlers.NewHealthHandler(mockService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HealthCheck(c)

	if err != nil {
		t.Fatalf("HealthCheck() error = %v, want nil", err)
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("HealthCheck() status code = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var result response.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("HealthCheck() failed to unmarshal response: %v", err)
	}

	if result.Status != "degraded" {
		t.Errorf("HealthCheck() status = %v, want 'degraded'", result.Status)
	}
}

func TestHealthHandler_HealthCheck_Unhealthy(t *testing.T) {
	healthResp := response.NewHealthResponse()
	healthResp.SetComponent("database", "unhealthy", "Connection failed")
	healthResp.SetComponent("redis", "unhealthy", "Connection failed")
	healthResp.CalculateOverallStatus()

	mockService := &mockHealthService{healthResp: healthResp}
	handler := handlers.NewHealthHandler(mockService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HealthCheck(c)

	if err != nil {
		t.Fatalf("HealthCheck() error = %v, want nil", err)
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("HealthCheck() status code = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var result response.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("HealthCheck() failed to unmarshal response: %v", err)
	}

	if result.Status != "unhealthy" {
		t.Errorf("HealthCheck() status = %v, want 'unhealthy'", result.Status)
	}
}

func TestHealthHandler_HealthCheck_ServiceError(t *testing.T) {
	mockService := &mockHealthService{err: errors.New("service error")}
	handler := handlers.NewHealthHandler(mockService)

	e := echo.New()
	e.GET("/health", handler.HealthCheck)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("HealthCheck() status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("HealthCheck() failed to unmarshal response: %v", err)
	}

	if result["message"] != "Failed to check health" {
		t.Errorf("HealthCheck() error message = %v, want 'Failed to check health'", result["message"])
	}
}

func TestHealthHandler_HealthCheck_JSONStructure(t *testing.T) {
	healthResp := response.NewHealthResponse()
	healthResp.SetComponent("database", "healthy", "Connection successful")
	healthResp.SetComponent("redis", "healthy", "Connection successful")
	healthResp.CalculateOverallStatus()

	mockService := &mockHealthService{healthResp: healthResp}
	handler := handlers.NewHealthHandler(mockService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HealthCheck(c)

	if err != nil {
		t.Fatalf("HealthCheck() error = %v, want nil", err)
	}

	var result response.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("HealthCheck() failed to unmarshal response: %v", err)
	}

	// Verify JSON structure
	if result.Status == "" {
		t.Error("HealthCheck() response missing 'status' field")
	}

	if result.Timestamp == "" {
		t.Error("HealthCheck() response missing 'timestamp' field")
	}

	if len(result.Components) != 2 {
		t.Errorf("HealthCheck() components count = %d, want 2", len(result.Components))
	}

	dbHealth, exists := result.Components["database"]
	if !exists {
		t.Error("HealthCheck() response missing 'database' component")
	} else {
		if dbHealth.Status == "" {
			t.Error("HealthCheck() database component missing 'status' field")
		}
		if dbHealth.Message == "" {
			t.Error("HealthCheck() database component missing 'message' field")
		}
	}

	redisHealth, exists := result.Components["redis"]
	if !exists {
		t.Error("HealthCheck() response missing 'redis' component")
	} else {
		if redisHealth.Status == "" {
			t.Error("HealthCheck() redis component missing 'status' field")
		}
		if redisHealth.Message == "" {
			t.Error("HealthCheck() redis component missing 'message' field")
		}
	}
}

