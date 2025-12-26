package integration

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/routes"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/services/health"
	"github.com/felipesantos/anki-backend/infra/postgres"
	"github.com/felipesantos/anki-backend/infra/redis"
)

// setupIntegrationHealthServer creates an Echo server with health check route using real services
func setupIntegrationHealthServer(t *testing.T) (*echo.Echo, *httptest.Server, func()) {
	// Skip if services are not available
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping integration tests (SKIP_DB_TESTS=true)")
	}

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	// Create database connection
	db, err := postgres.NewPostgresRepository(cfg.Database, logger)
	if err != nil {
		t.Skipf("Database not available: %v", err)
		return nil, nil, func() {}
	}

	// Create Redis connection
	rdb, err := redis.NewRedisRepository(cfg.Redis, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		db.Close()
		return nil, nil, func() {}
	}

	// Create health service
	healthService := health.NewHealthService(db, rdb)

	// Create Echo server
	e := echo.New()
	e.HideBanner = true

	// Register routes
	routes.RegisterHealthRoutes(e, healthService)

	// Create test server
	server := httptest.NewServer(e)

	// Cleanup function
	cleanup := func() {
		server.Close()
		rdb.Close()
		db.Close()
	}

	return e, server, cleanup
}

func TestHealth_Integration_WithRealServices(t *testing.T) {
	e, server, cleanup := setupIntegrationHealthServer(t)
	if e == nil {
		return // Skipped
	}
	defer cleanup()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should return 200 if all services are healthy, or 503 if degraded/unhealthy
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("GET /health status code = %d, want %d or %d", resp.StatusCode, http.StatusOK, http.StatusServiceUnavailable)
	}

	var result response.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if result.Status == "" {
		t.Error("Response missing 'status' field")
	}

	// Verify both components are checked
	if len(result.Components) != 2 {
		t.Errorf("Components count = %d, want 2", len(result.Components))
	}

	// Verify database component
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
		// With real services, database should be healthy (or we skipped the test)
		if dbHealth.Status != "healthy" && dbHealth.Status != "unhealthy" {
			t.Errorf("Database status = %v, want 'healthy' or 'unhealthy'", dbHealth.Status)
		}
	}

	// Verify redis component
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
		// With real services, redis should be healthy (or we skipped the test)
		if redisHealth.Status != "healthy" && redisHealth.Status != "unhealthy" {
			t.Errorf("Redis status = %v, want 'healthy' or 'unhealthy'", redisHealth.Status)
		}
	}
}

func TestHealth_Integration_StatusCodes(t *testing.T) {
	e, server, cleanup := setupIntegrationHealthServer(t)
	if e == nil {
		return // Skipped
	}
	defer cleanup()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var result response.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify status code matches health status
	if result.Status == "healthy" {
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status 'healthy' should return %d, got %d", http.StatusOK, resp.StatusCode)
		}
	} else {
		// degraded or unhealthy should return 503
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("Status '%s' should return %d, got %d", result.Status, http.StatusServiceUnavailable, resp.StatusCode)
		}
	}
}

func TestHealth_Integration_ResponseFields(t *testing.T) {
	e, server, cleanup := setupIntegrationHealthServer(t)
	if e == nil {
		return // Skipped
	}
	defer cleanup()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var result response.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify all required fields are present
	if result.Status == "" {
		t.Error("Response missing 'status' field")
	}

	if result.Timestamp == "" {
		t.Error("Response missing 'timestamp' field")
	}

	if result.Components == nil {
		t.Error("Response missing 'components' field")
	}

	// Verify timestamp is present (format validation is done in unit tests)
	if len(result.Timestamp) == 0 {
		t.Error("Timestamp is empty")
	}
}

