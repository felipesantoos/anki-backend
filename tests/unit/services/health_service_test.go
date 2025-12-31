package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

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

func (m *mockDatabaseRepository) GetDB() *sql.DB {
	return nil
}

// mockCacheRepositoryForHealth is a mock implementation of ICacheRepository for health tests
// Note: A separate type name is used to avoid conflict with mockCacheRepository in cache_service_test.go
type mockCacheRepositoryForHealth struct {
	pingErr error
}

func (m *mockCacheRepositoryForHealth) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockCacheRepositoryForHealth) Get(ctx context.Context, key string) (string, error) {
	return "", errors.New("not implemented")
}

func (m *mockCacheRepositoryForHealth) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return errors.New("not implemented")
}

func (m *mockCacheRepositoryForHealth) Delete(ctx context.Context, key string) error {
	return errors.New("not implemented")
}

func (m *mockCacheRepositoryForHealth) Exists(ctx context.Context, key string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockCacheRepositoryForHealth) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockCacheRepositoryForHealth) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return errors.New("not implemented")
}

func (m *mockCacheRepositoryForHealth) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, errors.New("not implemented")
}

func TestHealthService_CheckHealth_AllHealthy(t *testing.T) {
	mockDB := &mockDatabaseRepository{pingErr: nil}
	mockCache := &mockCacheRepositoryForHealth{pingErr: nil}

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
	mockCache := &mockCacheRepositoryForHealth{pingErr: nil}

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
	mockCache := &mockCacheRepositoryForHealth{pingErr: errors.New("connection refused")}

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
	mockCache := &mockCacheRepositoryForHealth{pingErr: errors.New("connection refused")}

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
	mockCache := &mockCacheRepositoryForHealth{pingErr: nil}

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
	mockCache := &mockCacheRepositoryForHealth{pingErr: nil}

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
	mockCache := &mockCacheRepositoryForHealth{pingErr: nil}

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

