package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/services/cache"
)

// mockCacheRepository is a mock implementation of ICacheRepository for testing
type mockCacheRepository struct {
	getFunc    func(ctx context.Context, key string) (string, error)
	setFunc    func(ctx context.Context, key string, value string, ttl time.Duration) error
	deleteFunc func(ctx context.Context, key string) error
	existsFunc func(ctx context.Context, key string) (bool, error)
	pingFunc   func(ctx context.Context) error
}

func (m *mockCacheRepository) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func (m *mockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	return "", errors.New("not implemented")
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, ttl)
	}
	return errors.New("not implemented")
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return errors.New("not implemented")
}

func (m *mockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, key)
	}
	return false, errors.New("not implemented")
}

func (m *mockCacheRepository) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockCacheRepository) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return errors.New("not implemented")
}

func (m *mockCacheRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, errors.New("not implemented")
}

func TestCacheService_Get(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		mockGet   func(ctx context.Context, key string) (string, error)
		wantValue string
		wantErr   bool
	}{
		{
			name: "successful get",
			key:  "test-key",
			mockGet: func(ctx context.Context, key string) (string, error) {
				return "test-value", nil
			},
			wantValue: "test-value",
			wantErr:   false,
		},
		{
			name: "key not found",
			key:  "nonexistent",
			mockGet: func(ctx context.Context, key string) (string, error) {
				return "", errors.New("key not found: nonexistent")
			},
			wantValue: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCacheRepository{
				getFunc: tt.mockGet,
			}
			service := cache.NewCacheService(repo)

			value, err := service.Get(context.Background(), tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if value != tt.wantValue {
				t.Errorf("Get() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

func TestCacheService_Set(t *testing.T) {
	var capturedKey, capturedValue string
	var capturedTTL time.Duration

	repo := &mockCacheRepository{
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			capturedKey = key
			capturedValue = value
			capturedTTL = ttl
			return nil
		},
	}
	service := cache.NewCacheService(repo)

	key := "test-key"
	value := "test-value"
	ttl := 5 * time.Minute

	err := service.Set(context.Background(), key, value, ttl)
	if err != nil {
		t.Errorf("Set() error = %v, want nil", err)
	}

	if capturedKey != key {
		t.Errorf("Set() captured key = %v, want %v", capturedKey, key)
	}
	if capturedValue != value {
		t.Errorf("Set() captured value = %v, want %v", capturedValue, value)
	}
	if capturedTTL != ttl {
		t.Errorf("Set() captured TTL = %v, want %v", capturedTTL, ttl)
	}
}

func TestCacheService_Delete(t *testing.T) {
	var capturedKey string

	repo := &mockCacheRepository{
		deleteFunc: func(ctx context.Context, key string) error {
			capturedKey = key
			return nil
		},
	}
	service := cache.NewCacheService(repo)

	key := "test-key"
	err := service.Delete(context.Background(), key)
	if err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	if capturedKey != key {
		t.Errorf("Delete() captured key = %v, want %v", capturedKey, key)
	}
}

func TestCacheService_GetOrSet_CacheHit(t *testing.T) {
	repo := &mockCacheRepository{
		getFunc: func(ctx context.Context, key string) (string, error) {
			return "cached-value", nil
		},
	}
	service := cache.NewCacheService(repo)

	fetchCalled := false
	fetchFunc := func() (string, error) {
		fetchCalled = true
		return "fetched-value", nil
	}

	value, err := service.GetOrSet(context.Background(), "test-key", fetchFunc, time.Minute)
	if err != nil {
		t.Errorf("GetOrSet() error = %v, want nil", err)
	}

	if value != "cached-value" {
		t.Errorf("GetOrSet() value = %v, want cached-value", value)
	}

	if fetchCalled {
		t.Error("GetOrSet() fetchFunc was called, but should have returned cached value")
	}
}

func TestCacheService_GetOrSet_CacheMiss(t *testing.T) {
	setCalled := false
	repo := &mockCacheRepository{
		getFunc: func(ctx context.Context, key string) (string, error) {
			return "", errors.New("key not found: test-key")
		},
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			setCalled = true
			return nil
		},
	}
	service := cache.NewCacheService(repo)

	fetchFunc := func() (string, error) {
		return "fetched-value", nil
	}

	value, err := service.GetOrSet(context.Background(), "test-key", fetchFunc, time.Minute)
	if err != nil {
		t.Errorf("GetOrSet() error = %v, want nil", err)
	}

	if value != "fetched-value" {
		t.Errorf("GetOrSet() value = %v, want fetched-value", value)
	}

	if !setCalled {
		t.Error("GetOrSet() setFunc was not called after cache miss")
	}
}

func TestCacheService_Exists(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		mockExists func(ctx context.Context, key string) (bool, error)
		want      bool
		wantErr   bool
	}{
		{
			name: "key exists",
			key:  "test-key",
			mockExists: func(ctx context.Context, key string) (bool, error) {
				return true, nil
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "key does not exist",
			key:  "nonexistent",
			mockExists: func(ctx context.Context, key string) (bool, error) {
				return false, nil
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCacheRepository{
				existsFunc: tt.mockExists,
			}
			service := cache.NewCacheService(repo)

			exists, err := service.Exists(context.Background(), tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if exists != tt.want {
				t.Errorf("Exists() = %v, want %v", exists, tt.want)
			}
		})
	}
}

