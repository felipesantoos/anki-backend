package secondary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/events"
)

// EventHandler defines the interface for handling domain events
type EventHandler interface {
	// Handle processes the given domain event
	Handle(ctx context.Context, event events.DomainEvent) error

	// EventType returns the type of event this handler processes
	EventType() string

	// HandlerID returns a unique identifier for this handler instance
	HandlerID() string
}

// IEventBus defines the interface for publishing and subscribing to domain events
type IEventBus interface {
	// Publish publishes a domain event to all subscribed handlers
	Publish(ctx context.Context, event events.DomainEvent) error

	// Subscribe subscribes a handler to a specific event type
	Subscribe(eventType string, handler EventHandler) error

	// Unsubscribe removes a handler subscription for a specific event type
	Unsubscribe(eventType string, handlerID string) error

	// Start starts the event bus processing (if async processing is needed)
	Start() error

	// Stop stops the event bus gracefully, waiting for pending events to be processed
	Stop() error
}
