package events

import (
	"context"
	"fmt"

	"github.com/felipesantos/anki-backend/core/domain/events"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
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
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eventType := event.EventType()
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	if event.AggregateID() == "" {
		return fmt.Errorf("event aggregate ID cannot be empty")
	}

	return s.bus.Publish(ctx, event)
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
