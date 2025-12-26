package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/routes"
	"github.com/felipesantos/anki-backend/core/services/health"
	"github.com/felipesantos/anki-backend/infra/postgres"
	"github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/config"
	"log/slog"
)

// setupHealthTestServer creates an Echo server with health check route for E2E tests
func setupHealthTestServer(t *testing.T) (*echo.Echo, *httptest.Server) {
	e := echo.New()
	e.HideBanner = true

	// Load configuration (may fail if env vars not set, but that's ok for E2E)
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Skipping E2E test - failed to load config: %v", err)
	}

	logger := slog.Default()

	// Try to create real connections (may fail if services not available)
	db, err := postgres.NewPostgresRepository(cfg.Database, logger)
	if err != nil {
		t.Skipf("Skipping E2E test - database not available: %v", err)
		return nil, nil
	}

	rdb, err := redis.NewRedisRepository(cfg.Redis, logger)
	if err != nil {
		t.Skipf("Skipping E2E test - Redis not available: %v", err)
		db.Close()
		return nil, nil
	}

	// Create health service
	healthService := health.NewHealthService(db, rdb)

	// Register routes
	routes.RegisterHealthRoutes(e, healthService)

	// Create test server
	server := httptest.NewServer(e)
	return e, server
}

func TestHealth_E2E_EndpointExists(t *testing.T) {
	e, server := setupHealthTestServer(t)
	if e == nil {
		return // Skipped
	}
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("GET /health status code = %d, want %d or %d", resp.StatusCode, http.StatusOK, http.StatusServiceUnavailable)
	}
}

func TestHealth_E2E_ResponseStructure(t *testing.T) {
	e, server := setupHealthTestServer(t)
	if e == nil {
		return // Skipped
	}
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var result response.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if result.Status == "" {
		t.Error("Response missing 'status' field")
	}

	if result.Timestamp == "" {
		t.Error("Response missing 'timestamp' field")
	}

	if len(result.Components) == 0 {
		t.Error("Response missing 'components' field")
	}

	// Verify expected components exist
	dbHealth, exists := result.Components["database"]
	if !exists {
		t.Error("Response missing 'database' component")
	} else {
		if dbHealth.Status == "" {
			t.Error("Database component missing 'status' field")
		}
		if dbHealth.Message == "" {
			t.Error("Database component missing 'message' field")
		}
	}

	redisHealth, exists := result.Components["redis"]
	if !exists {
		t.Error("Response missing 'redis' component")
	} else {
		if redisHealth.Status == "" {
			t.Error("Redis component missing 'status' field")
		}
		if redisHealth.Message == "" {
			t.Error("Redis component missing 'message' field")
		}
	}
}

func TestHealth_E2E_ValidStatusValues(t *testing.T) {
	e, server := setupHealthTestServer(t)
	if e == nil {
		return // Skipped
	}
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var result response.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify status is one of the valid values
	validStatuses := map[string]bool{
		"healthy":   true,
		"degraded":  true,
		"unhealthy": true,
	}

	if !validStatuses[result.Status] {
		t.Errorf("Status = %v, want one of: healthy, degraded, unhealthy", result.Status)
	}

	// Verify component statuses are valid
	validComponentStatuses := map[string]bool{
		"healthy":   true,
		"unhealthy": true,
	}

	for componentName, component := range result.Components {
		if !validComponentStatuses[component.Status] {
			t.Errorf("Component %s status = %v, want one of: healthy, unhealthy", componentName, component.Status)
		}
	}
}

