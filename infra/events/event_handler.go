package events

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// BaseEventHandler is a base struct that can be embedded in custom event handlers
// to provide default implementation of HandlerID
type BaseEventHandler struct {
	handlerID string
	eventType string
}

// NewBaseEventHandler creates a new base event handler
func NewBaseEventHandler(eventType string) *BaseEventHandler {
	return &BaseEventHandler{
		handlerID: generateHandlerID(),
		eventType: eventType,
	}
}

// HandlerID returns the unique identifier for this handler
func (h *BaseEventHandler) HandlerID() string {
	return h.handlerID
}

// EventType returns the event type this handler processes
func (h *BaseEventHandler) EventType() string {
	return h.eventType
}

// generateHandlerID generates a unique handler ID
func generateHandlerID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("handler_%d", len(bytes))
	}
	return hex.EncodeToString(bytes)
}

// Note: BaseEventHandler doesn't implement Handle() method, so it doesn't fully
// implement the EventHandler interface. Custom handlers should embed BaseEventHandler
// and implement Handle() to satisfy the interface.

// Example usage:
//
// type MyEventHandler struct {
//     *BaseEventHandler
//     // additional fields
// }
//
// func (h *MyEventHandler) Handle(ctx context.Context, event events.DomainEvent) error {
//     // implementation
//     return nil
// }
//
// func NewMyEventHandler() *MyEventHandler {
//     return &MyEventHandler{
//         BaseEventHandler: NewBaseEventHandler("card.reviewed"),
//     }
// }
