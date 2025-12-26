package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/events"
	eventHandlers "github.com/felipesantos/anki-backend/infra/events/handlers"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

// testEventHandler is a test handler that tracks processed events
type testEventHandler struct {
	*infraEvents.BaseEventHandler
	processedEvents []events.DomainEvent
	mu              sync.Mutex
}

func newTestEventHandler(eventType string) *testEventHandler {
	return &testEventHandler{
		BaseEventHandler: infraEvents.NewBaseEventHandler(eventType),
		processedEvents:  make([]events.DomainEvent, 0),
	}
}

func (h *testEventHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.processedEvents = append(h.processedEvents, event)
	return nil
}

func (h *testEventHandler) GetProcessedCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.processedEvents)
}

func (h *testEventHandler) GetProcessedEvents() []events.DomainEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]events.DomainEvent, len(h.processedEvents))
	copy(result, h.processedEvents)
	return result
}

func TestEventBus_PublishAndSubscribe(t *testing.T) {
	log := logger.GetLogger()
	bus := infraEvents.NewInMemoryEventBus(2, 100, log)

	// Create test handler - subscribe to the actual event type
	handler := newTestEventHandler(events.CardReviewedEventType)
	
	// Subscribe handler
	err := bus.Subscribe(events.CardReviewedEventType, handler)
	if err != nil {
		t.Fatalf("Failed to subscribe handler: %v", err)
	}

	// Start bus
	err = bus.Start()
	if err != nil {
		t.Fatalf("Failed to start bus: %v", err)
	}
	defer bus.Stop()

	// Create and publish event
	event := &events.CardReviewed{
		CardID:    123,
		UserID:    456,
		Rating:    5,
		NewState:  "reviewed",
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err = bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Wait for event to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify handler received the event
	count := handler.GetProcessedCount()
	if count != 1 {
		t.Fatalf("Expected handler to process 1 event, got: %d", count)
	}
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	log := logger.GetLogger()
	bus := infraEvents.NewInMemoryEventBus(2, 100, log)

	// Create multiple handlers for the same event type - subscribe to the actual event type
	handler1 := newTestEventHandler(events.NoteCreatedEventType)
	handler2 := newTestEventHandler(events.NoteCreatedEventType)
	handler3 := newTestEventHandler(events.NoteCreatedEventType)

	// Subscribe all handlers
	err := bus.Subscribe(events.NoteCreatedEventType, handler1)
	if err != nil {
		t.Fatalf("Failed to subscribe handler1: %v", err)
	}

	err = bus.Subscribe(events.NoteCreatedEventType, handler2)
	if err != nil {
		t.Fatalf("Failed to subscribe handler2: %v", err)
	}

	err = bus.Subscribe(events.NoteCreatedEventType, handler3)
	if err != nil {
		t.Fatalf("Failed to subscribe handler3: %v", err)
	}

	// Start bus
	err = bus.Start()
	if err != nil {
		t.Fatalf("Failed to start bus: %v", err)
	}
	defer bus.Stop()

	// Publish event
	event := &events.NoteCreated{
		NoteID:    789,
		UserID:    456,
		NoteTypeID: 1,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err = bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Wait for all handlers to process
	time.Sleep(200 * time.Millisecond)

	// Verify all handlers received the event
	if handler1.GetProcessedCount() != 1 {
		t.Errorf("Handler1: expected 1 event, got: %d", handler1.GetProcessedCount())
	}
	if handler2.GetProcessedCount() != 1 {
		t.Errorf("Handler2: expected 1 event, got: %d", handler2.GetProcessedCount())
	}
	if handler3.GetProcessedCount() != 1 {
		t.Errorf("Handler3: expected 1 event, got: %d", handler3.GetProcessedCount())
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	log := logger.GetLogger()
	bus := infraEvents.NewInMemoryEventBus(2, 100, log)

	// Subscribe to the actual event type
	handler1 := newTestEventHandler(events.DeckUpdatedEventType)
	handler2 := newTestEventHandler(events.DeckUpdatedEventType)

	// Subscribe both handlers
	err := bus.Subscribe(events.DeckUpdatedEventType, handler1)
	if err != nil {
		t.Fatalf("Failed to subscribe handler1: %v", err)
	}

	err = bus.Subscribe(events.DeckUpdatedEventType, handler2)
	if err != nil {
		t.Fatalf("Failed to subscribe handler2: %v", err)
	}

	// Start bus
	err = bus.Start()
	if err != nil {
		t.Fatalf("Failed to start bus: %v", err)
	}
	defer bus.Stop()

	// Unsubscribe handler1
	err = bus.Unsubscribe(events.DeckUpdatedEventType, handler1.HandlerID())
	if err != nil {
		t.Fatalf("Failed to unsubscribe handler1: %v", err)
	}

	// Publish event
	event := &events.DeckUpdated{
		DeckID:    111,
		UserID:    222,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err = bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify only handler2 received the event
	if handler1.GetProcessedCount() != 0 {
		t.Errorf("Handler1: expected 0 events (unsubscribed), got: %d", handler1.GetProcessedCount())
	}
	if handler2.GetProcessedCount() != 1 {
		t.Errorf("Handler2: expected 1 event, got: %d", handler2.GetProcessedCount())
	}
}

func TestEventBus_AsyncProcessing(t *testing.T) {
	log := logger.GetLogger()
	bus := infraEvents.NewInMemoryEventBus(3, 100, log)

	// Subscribe to the actual event type
	handler := newTestEventHandler(events.CardReviewedEventType)
	
	err := bus.Subscribe(events.CardReviewedEventType, handler)
	if err != nil {
		t.Fatalf("Failed to subscribe handler: %v", err)
	}

	err = bus.Start()
	if err != nil {
		t.Fatalf("Failed to start bus: %v", err)
	}
	defer bus.Stop()

	// Publish multiple events rapidly
	ctx := context.Background()
	numEvents := 10
	for i := 0; i < numEvents; i++ {
		event := &events.CardReviewed{
			CardID:    int64(i),
			UserID:    456,
			Rating:    5,
			NewState:  "reviewed",
			Timestamp: time.Now(),
		}
		err = bus.Publish(ctx, event)
		if err != nil {
			t.Fatalf("Failed to publish event %d: %v", i, err)
		}
	}

	// Wait for all events to be processed
	time.Sleep(500 * time.Millisecond)

	// Verify all events were processed
	count := handler.GetProcessedCount()
	if count != numEvents {
		t.Fatalf("Expected %d events to be processed, got: %d", numEvents, count)
	}
}

func TestEventBus_GracefulShutdown(t *testing.T) {
	log := logger.GetLogger()
	bus := infraEvents.NewInMemoryEventBus(2, 100, log)

	// Subscribe to the actual event type
	handler := newTestEventHandler(events.NoteCreatedEventType)
	
	err := bus.Subscribe(events.NoteCreatedEventType, handler)
	if err != nil {
		t.Fatalf("Failed to subscribe handler: %v", err)
	}

	err = bus.Start()
	if err != nil {
		t.Fatalf("Failed to start bus: %v", err)
	}

	// Publish some events
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		event := &events.NoteCreated{
			NoteID:    int64(i),
			UserID:    456,
			NoteTypeID: 1,
			Timestamp: time.Now(),
		}
		bus.Publish(ctx, event)
	}

	// Stop bus (should wait for pending events)
	err = bus.Stop()
	if err != nil {
		t.Fatalf("Failed to stop bus: %v", err)
	}

	// Verify events were processed
	count := handler.GetProcessedCount()
	if count != 5 {
		t.Errorf("Expected 5 events to be processed before shutdown, got: %d", count)
	}
}

func TestEventBus_ExampleHandler(t *testing.T) {
	log := logger.GetLogger()
	bus := infraEvents.NewInMemoryEventBus(2, 100, log)

	// Use the example handler from infra/events/handlers
	handler := eventHandlers.NewExampleHandler(events.CardReviewedEventType)
	
	err := bus.Subscribe(events.CardReviewedEventType, handler)
	if err != nil {
		t.Fatalf("Failed to subscribe handler: %v", err)
	}

	err = bus.Start()
	if err != nil {
		t.Fatalf("Failed to start bus: %v", err)
	}
	defer bus.Stop()

	// Publish CardReviewed event
	event := &events.CardReviewed{
		CardID:    999,
		UserID:    888,
		Rating:    4,
		NewState:  "mastered",
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err = bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)
	// If no error occurred, the handler processed successfully
}
