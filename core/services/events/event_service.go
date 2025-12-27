package events

import (
	"context"
	"fmt"

	"github.com/felipesantos/anki-backend/core/domain/events"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// EventService provides high-level operations for managing domain events
type EventService struct {
	bus secondary.IEventBus
}

// NewEventService creates a new EventService instance
func NewEventService(bus secondary.IEventBus) *EventService {
	return &EventService{
		bus: bus,
	}
}

// Publish publishes a domain event to all subscribed handlers
func (s *EventService) Publish(ctx context.Context, event events.DomainEvent) error {
	eventType := ""
	aggregateID := ""
	
	if event != nil {
		eventType = event.EventType()
		aggregateID = event.AggregateID()
	}

	ctx, span := tracing.StartSpan(ctx, "event.publish",
		trace.WithAttributes(
			attribute.String("event.type", eventType),
			attribute.String("event.aggregate_id", aggregateID),
		),
	)
	defer span.End()

	if event == nil {
		err := fmt.Errorf("event cannot be nil")
		tracing.RecordError(span, err)
		return err
	}

	if eventType == "" {
		err := fmt.Errorf("event type cannot be empty")
		tracing.RecordError(span, err)
		return err
	}

	if aggregateID == "" {
		err := fmt.Errorf("event aggregate ID cannot be empty")
		tracing.RecordError(span, err)
		return err
	}

	err := s.bus.Publish(ctx, event)
	if err != nil {
		tracing.RecordError(span, err)
		return err
	}
	return nil
}

// Subscribe subscribes a handler to a specific event type
func (s *EventService) Subscribe(eventType string, handler secondary.EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	return s.bus.Subscribe(eventType, handler)
}

// Unsubscribe removes a handler subscription for a specific event type
func (s *EventService) Unsubscribe(eventType string, handlerID string) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handlerID == "" {
		return fmt.Errorf("handler ID cannot be empty")
	}

	return s.bus.Unsubscribe(eventType, handlerID)
}

// Bus returns the underlying event bus (for advanced use cases)
func (s *EventService) Bus() secondary.IEventBus {
	return s.bus
}
