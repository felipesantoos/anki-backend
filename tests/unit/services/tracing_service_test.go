package services

import (
	"context"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/services/tracing"
	"github.com/stretchr/testify/assert"
)

func TestTracingService(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		cfg := config.TracingConfig{
			Enabled: false,
		}
		service, err := tracing.NewTracingService(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.False(t, service.IsEnabled())
		assert.NotNil(t, service.GetTracer())
	})

	t.Run("Enabled_ConsoleOnly", func(t *testing.T) {
		cfg := config.TracingConfig{
			Enabled:        true,
			ServiceName:    "test-service",
			Environment:    "test",
			SampleRate:     1.0,
			ConsoleEnabled: true,
		}
		service, err := tracing.NewTracingService(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.True(t, service.IsEnabled())
		assert.NotNil(t, service.GetTracer())

		// Test shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err = service.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("Enabled_NoExporters", func(t *testing.T) {
		cfg := config.TracingConfig{
			Enabled: true,
		}
		service, err := tracing.NewTracingService(cfg)

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "no exporters configured")
	})
}

