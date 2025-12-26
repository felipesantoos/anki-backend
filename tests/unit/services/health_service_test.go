package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/services/health"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// mockDatabaseRepository is a mock implementation of IDatabaseRepository
type mockDatabaseRepository struct {
	pingErr error
}

func (m *mockDatabaseRepository) Ping(ctx context.Context) error {
	return m.pingErr
}

// mockCacheRepository is a mock implementation of ICacheRepository
type mockCacheRepository struct {
	pingErr error
}

func (m *mockCacheRepository) Ping(ctx context.Context) error {
	return m.pingErr
}

func TestHealthService_CheckHealth_AllHealthy(t *testing.T) {
	mockDB := &mockDatabaseRepository{pingErr: nil}
	mockCache := &mockCacheRepository{pingErr: nil}

	service := health.NewHealthService(mockDB, mockCache)

	ctx := context.Background()
	healthResp, err := service.CheckHealth(ctx)

	if err != nil {
		t.Fatalf("CheckHealth() error = %v, want nil", err)
	}

	if healthResp.Status != "healthy" {
		t.Errorf("CheckHealth() status = %v, want 'healthy'", healthResp.Status)
	}

	dbHealth, exists := healthResp.Components["database"]
	if !exists {
		t.Error("CheckHealth() missing 'database' component")
	} else if dbHealth.Status != "healthy" {
		t.Errorf("CheckHealth() database status = %v, want 'healthy'", dbHealth.Status)
	}

	redisHealth, exists := healthResp.Components["redis"]
	if !exists {
		t.Error("CheckHealth() missing 'redis' component")
	} else if redisHealth.Status != "healthy" {
		t.Errorf("CheckHealth() redis status = %v, want 'healthy'", redisHealth.Status)
	}
}

func TestHealthService_CheckHealth_DatabaseDown(t *testing.T) {
	mockDB := &mockDatabaseRepository{pingErr: errors.New("connection refused")}
	mockCache := &mockCacheRepository{pingErr: nil}

	service := health.NewHealthService(mockDB, mockCache)

	ctx := context.Background()
	healthResp, err := service.CheckHealth(ctx)

	if err != nil {
		t.Fatalf("CheckHealth() error = %v, want nil", err)
	}

	if healthResp.Status != "degraded" {
		t.Errorf("CheckHealth() status = %v, want 'degraded'", healthResp.Status)
	}

	dbHealth, exists := healthResp.Components["database"]
	if !exists {
		t.Error("CheckHealth() missing 'database' component")
	} else if dbHealth.Status != "unhealthy" {
		t.Errorf("CheckHealth() database status = %v, want 'unhealthy'", dbHealth.Status)
	}

	redisHealth, exists := healthResp.Components["redis"]
	if !exists {
		t.Error("CheckHealth() missing 'redis' component")
	} else if redisHealth.Status != "healthy" {
		t.Errorf("CheckHealth() redis status = %v, want 'healthy'", redisHealth.Status)
	}
}

func TestHealthService_CheckHealth_CacheDown(t *testing.T) {
	mockDB := &mockDatabaseRepository{pingErr: nil}
	mockCache := &mockCacheRepository{pingErr: errors.New("connection refused")}

	service := health.NewHealthService(mockDB, mockCache)

	ctx := context.Background()
	healthResp, err := service.CheckHealth(ctx)

	if err != nil {
		t.Fatalf("CheckHealth() error = %v, want nil", err)
	}

	if healthResp.Status != "degraded" {
		t.Errorf("CheckHealth() status = %v, want 'degraded'", healthResp.Status)
	}

	dbHealth, exists := healthResp.Components["database"]
	if !exists {
		t.Error("CheckHealth() missing 'database' component")
	} else if dbHealth.Status != "healthy" {
		t.Errorf("CheckHealth() database status = %v, want 'healthy'", dbHealth.Status)
	}

	redisHealth, exists := healthResp.Components["redis"]
	if !exists {
		t.Error("CheckHealth() missing 'redis' component")
	} else if redisHealth.Status != "unhealthy" {
		t.Errorf("CheckHealth() redis status = %v, want 'unhealthy'", redisHealth.Status)
	}
}

func TestHealthService_CheckHealth_BothDown(t *testing.T) {
	mockDB := &mockDatabaseRepository{pingErr: errors.New("connection refused")}
	mockCache := &mockCacheRepository{pingErr: errors.New("connection refused")}

	service := health.NewHealthService(mockDB, mockCache)

	ctx := context.Background()
	healthResp, err := service.CheckHealth(ctx)

	if err != nil {
		t.Fatalf("CheckHealth() error = %v, want nil", err)
	}

	if healthResp.Status != "unhealthy" {
		t.Errorf("CheckHealth() status = %v, want 'unhealthy'", healthResp.Status)
	}

	dbHealth, exists := healthResp.Components["database"]
	if !exists {
		t.Error("CheckHealth() missing 'database' component")
	} else if dbHealth.Status != "unhealthy" {
		t.Errorf("CheckHealth() database status = %v, want 'unhealthy'", dbHealth.Status)
	}

	redisHealth, exists := healthResp.Components["redis"]
	if !exists {
		t.Error("CheckHealth() missing 'redis' component")
	} else if redisHealth.Status != "unhealthy" {
		t.Errorf("CheckHealth() redis status = %v, want 'unhealthy'", redisHealth.Status)
	}
}

func TestHealthService_CheckHealth_NilDatabase(t *testing.T) {
	var nilDB secondary.IDatabaseRepository = nil
	mockCache := &mockCacheRepository{pingErr: nil}

	service := health.NewHealthService(nilDB, mockCache)

	ctx := context.Background()
	healthResp, err := service.CheckHealth(ctx)

	if err != nil {
		t.Fatalf("CheckHealth() error = %v, want nil", err)
	}

	if healthResp.Status != "degraded" {
		t.Errorf("CheckHealth() status = %v, want 'degraded'", healthResp.Status)
	}

	dbHealth, exists := healthResp.Components["database"]
	if !exists {
		t.Error("CheckHealth() missing 'database' component")
	} else if dbHealth.Status != "unhealthy" {
		t.Errorf("CheckHealth() database status = %v, want 'unhealthy'", dbHealth.Status)
	}
}

func TestHealthService_CheckHealth_NilCache(t *testing.T) {
	mockDB := &mockDatabaseRepository{pingErr: nil}
	var nilCache secondary.ICacheRepository = nil

	service := health.NewHealthService(mockDB, nilCache)

	ctx := context.Background()
	healthResp, err := service.CheckHealth(ctx)

	if err != nil {
		t.Fatalf("CheckHealth() error = %v, want nil", err)
	}

	if healthResp.Status != "degraded" {
		t.Errorf("CheckHealth() status = %v, want 'degraded'", healthResp.Status)
	}

	redisHealth, exists := healthResp.Components["redis"]
	if !exists {
		t.Error("CheckHealth() missing 'redis' component")
	} else if redisHealth.Status != "unhealthy" {
		t.Errorf("CheckHealth() redis status = %v, want 'unhealthy'", redisHealth.Status)
	}
}

func TestHealthService_CheckHealth_Timeout(t *testing.T) {
	mockDB := &mockDatabaseRepository{pingErr: context.DeadlineExceeded}
	mockCache := &mockCacheRepository{pingErr: nil}

	service := health.NewHealthService(mockDB, mockCache)

	ctx := context.Background()
	healthResp, err := service.CheckHealth(ctx)

	if err != nil {
		t.Fatalf("CheckHealth() error = %v, want nil", err)
	}

	if healthResp.Status != "degraded" {
		t.Errorf("CheckHealth() status = %v, want 'degraded'", healthResp.Status)
	}

	dbHealth, exists := healthResp.Components["database"]
	if !exists {
		t.Error("CheckHealth() missing 'database' component")
	} else if dbHealth.Status != "unhealthy" {
		t.Errorf("CheckHealth() database status = %v, want 'unhealthy'", dbHealth.Status)
	}
}

func TestHealthService_CheckHealth_ResponseStructure(t *testing.T) {
	mockDB := &mockDatabaseRepository{pingErr: nil}
	mockCache := &mockCacheRepository{pingErr: nil}

	service := health.NewHealthService(mockDB, mockCache)

	ctx := context.Background()
	healthResp, err := service.CheckHealth(ctx)

	if err != nil {
		t.Fatalf("CheckHealth() error = %v, want nil", err)
	}

	// Verify response structure
	if healthResp.Timestamp == "" {
		t.Error("CheckHealth() timestamp is empty")
	}

	// Verify timestamp is valid RFC3339
	_, err = time.Parse(time.RFC3339, healthResp.Timestamp)
	if err != nil {
		t.Errorf("CheckHealth() timestamp is not valid RFC3339: %v", err)
	}

	if len(healthResp.Components) != 2 {
		t.Errorf("CheckHealth() components count = %d, want 2", len(healthResp.Components))
	}
}

