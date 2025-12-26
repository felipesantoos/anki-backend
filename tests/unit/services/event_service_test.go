package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/events"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	eventServices "github.com/felipesantos/anki-backend/core/services/events"
)

// mockEventBus is a mock implementation of IEventBus for testing
type mockEventBus struct {
	publishError    error
	subscribeError  error
	unsubscribeError error
	publishedEvents []events.DomainEvent
	subscribedHandlers map[string][]secondary.EventHandler
}

func newMockEventBus() *mockEventBus {
	return &mockEventBus{
		subscribedHandlers: make(map[string][]secondary.EventHandler),
		publishedEvents:    make([]events.DomainEvent, 0),
	}
}

func (m *mockEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	if m.publishError != nil {
		return m.publishError
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventBus) Subscribe(eventType string, handler secondary.EventHandler) error {
	if m.subscribeError != nil {
		return m.subscribeError
	}
	m.subscribedHandlers[eventType] = append(m.subscribedHandlers[eventType], handler)
	return nil
}

func (m *mockEventBus) Unsubscribe(eventType string, handlerID string) error {
	if m.unsubscribeError != nil {
		return m.unsubscribeError
	}
	handlers := m.subscribedHandlers[eventType]
	newHandlers := make([]secondary.EventHandler, 0)
	for _, h := range handlers {
		if h.HandlerID() != handlerID {
			newHandlers = append(newHandlers, h)
		}
	}
	m.subscribedHandlers[eventType] = newHandlers
	return nil
}

func (m *mockEventBus) Start() error {
	return nil
}

func (m *mockEventBus) Stop() error {
	return nil
}

// mockEventHandler is a mock implementation of EventHandler for testing
type mockEventHandler struct {
	handlerID string
	eventType string
	handleError error
	handledEvents []events.DomainEvent
}

func newMockEventHandler(id, eventType string) *mockEventHandler {
	return &mockEventHandler{
		handlerID:    id,
		eventType:    eventType,
		handledEvents: make([]events.DomainEvent, 0),
	}
}

func (m *mockEventHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	if m.handleError != nil {
		return m.handleError
	}
	m.handledEvents = append(m.handledEvents, event)
	return nil
}

func (m *mockEventHandler) EventType() string {
	return m.eventType
}

func (m *mockEventHandler) HandlerID() string {
	return m.handlerID
}

// mockDomainEvent is a simple mock domain event for testing
type mockDomainEvent struct {
	eventType   string
	aggregateID string
	occurredAt  time.Time
	metadata    map[string]interface{}
}

func newMockDomainEvent(eventType, aggregateID string) *mockDomainEvent {
	return &mockDomainEvent{
		eventType:   eventType,
		aggregateID: aggregateID,
		occurredAt:  time.Now(),
		metadata:    make(map[string]interface{}),
	}
}

func (m *mockDomainEvent) EventType() string {
	return m.eventType
}

func (m *mockDomainEvent) AggregateID() string {
	return m.aggregateID
}

func (m *mockDomainEvent) OccurredAt() time.Time {
	return m.occurredAt
}

func (m *mockDomainEvent) Metadata() map[string]interface{} {
	return m.metadata
}

func TestEventService_Publish_Success(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	event := newMockDomainEvent("test.event", "123")
	
	err := service.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(bus.publishedEvents) != 1 {
		t.Fatalf("Expected 1 published event, got: %d", len(bus.publishedEvents))
	}

	if bus.publishedEvents[0] != event {
		t.Error("Published event does not match")
	}
}

func TestEventService_Publish_NilEvent(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	err := service.Publish(context.Background(), nil)
	if err == nil {
		t.Fatal("Expected error for nil event, got nil")
	}
}

func TestEventService_Publish_EmptyEventType(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	event := &mockDomainEvent{
		eventType:   "",
		aggregateID: "123",
		occurredAt:  time.Now(),
		metadata:    make(map[string]interface{}),
	}

	err := service.Publish(context.Background(), event)
	if err == nil {
		t.Fatal("Expected error for empty event type, got nil")
	}
}

func TestEventService_Publish_EmptyAggregateID(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	event := &mockDomainEvent{
		eventType:   "test.event",
		aggregateID: "",
		occurredAt:  time.Now(),
		metadata:    make(map[string]interface{}),
	}

	err := service.Publish(context.Background(), event)
	if err == nil {
		t.Fatal("Expected error for empty aggregate ID, got nil")
	}
}

func TestEventService_Publish_BusError(t *testing.T) {
	bus := newMockEventBus()
	bus.publishError = errors.New("bus error")
	service := eventServices.NewEventService(bus)

	event := newMockDomainEvent("test.event", "123")
	
	err := service.Publish(context.Background(), event)
	if err == nil {
		t.Fatal("Expected error from bus, got nil")
	}
}

func TestEventService_Subscribe_Success(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	handler := newMockEventHandler("handler1", "test.event")
	
	err := service.Subscribe("test.event", handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	handlers := bus.subscribedHandlers["test.event"]
	if len(handlers) != 1 {
		t.Fatalf("Expected 1 subscribed handler, got: %d", len(handlers))
	}

	if handlers[0] != handler {
		t.Error("Subscribed handler does not match")
	}
}

func TestEventService_Subscribe_EmptyEventType(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	handler := newMockEventHandler("handler1", "test.event")
	
	err := service.Subscribe("", handler)
	if err == nil {
		t.Fatal("Expected error for empty event type, got nil")
	}
}

func TestEventService_Subscribe_NilHandler(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	err := service.Subscribe("test.event", nil)
	if err == nil {
		t.Fatal("Expected error for nil handler, got nil")
	}
}

func TestEventService_Subscribe_BusError(t *testing.T) {
	bus := newMockEventBus()
	bus.subscribeError = errors.New("subscribe error")
	service := eventServices.NewEventService(bus)

	handler := newMockEventHandler("handler1", "test.event")
	
	err := service.Subscribe("test.event", handler)
	if err == nil {
		t.Fatal("Expected error from bus, got nil")
	}
}

func TestEventService_Unsubscribe_Success(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	handler := newMockEventHandler("handler1", "test.event")
	
	// Subscribe first
	err := service.Subscribe("test.event", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Unsubscribe
	err = service.Unsubscribe("test.event", "handler1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	handlers := bus.subscribedHandlers["test.event"]
	if len(handlers) != 0 {
		t.Fatalf("Expected 0 handlers after unsubscribe, got: %d", len(handlers))
	}
}

func TestEventService_Unsubscribe_EmptyEventType(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	err := service.Unsubscribe("", "handler1")
	if err == nil {
		t.Fatal("Expected error for empty event type, got nil")
	}
}

func TestEventService_Unsubscribe_EmptyHandlerID(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	err := service.Unsubscribe("test.event", "")
	if err == nil {
		t.Fatal("Expected error for empty handler ID, got nil")
	}
}

func TestEventService_Bus(t *testing.T) {
	bus := newMockEventBus()
	service := eventServices.NewEventService(bus)

	returnedBus := service.Bus()
	if returnedBus != bus {
		t.Error("Bus() should return the same bus instance")
	}
}
