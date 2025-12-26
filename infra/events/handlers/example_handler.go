package handlers

import (
	"context"
	"log/slog"

	"github.com/felipesantos/anki-backend/core/domain/events"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

// ExampleHandler is an example event handler that demonstrates how to handle domain events
// This handler processes CardReviewed events
type ExampleHandler struct {
	*infraEvents.BaseEventHandler
	logger *slog.Logger
}

// NewExampleHandler creates a new example event handler
func NewExampleHandler(eventType string) *ExampleHandler {
	return &ExampleHandler{
		BaseEventHandler: infraEvents.NewBaseEventHandler(eventType),
		logger:           logger.GetLogger(),
	}
}

// Handle processes the domain event
func (h *ExampleHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Info("Example handler processing event",
		"handler_id", h.HandlerID(),
		"event_type", event.EventType(),
		"aggregate_id", event.AggregateID(),
		"occurred_at", event.OccurredAt(),
	)

	// Example: Handle specific event types
	switch e := event.(type) {
	case *events.CardReviewed:
		h.logger.Info("Processing CardReviewed event",
			"card_id", e.CardID,
			"user_id", e.UserID,
			"rating", e.Rating,
			"new_state", e.NewState,
		)
		// Add your business logic here
		// Example: update statistics, send notifications, etc.
	case *events.NoteCreated:
		h.logger.Info("Processing NoteCreated event",
			"note_id", e.NoteID,
			"user_id", e.UserID,
			"note_type_id", e.NoteTypeID,
		)
		// Add your business logic here
	case *events.DeckUpdated:
		h.logger.Info("Processing DeckUpdated event",
			"deck_id", e.DeckID,
			"user_id", e.UserID,
		)
		// Add your business logic here
	default:
		h.logger.Debug("Received event of unknown type",
			"event_type", event.EventType(),
		)
	}

	// Access metadata if needed
	if metadata := event.Metadata(); metadata != nil {
		h.logger.Debug("Event metadata", "metadata", metadata)
	}

	return nil
}

// Note: ExampleHandler embeds BaseEventHandler which provides HandlerID() and EventType()
// We only need to implement Handle() method to satisfy the EventHandler interface
