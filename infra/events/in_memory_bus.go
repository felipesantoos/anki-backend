package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/events"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// InMemoryEventBus implements IEventBus using in-memory processing
type InMemoryEventBus struct {
	handlers map[string][]secondary.EventHandler // eventType -> []handlers
	mu       sync.RWMutex
	logger   *slog.Logger
	queue    chan eventMessage
	workers  int
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	started  bool
	startMu  sync.Mutex
}

type eventMessage struct {
	ctx   context.Context
	event events.DomainEvent
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus(workerCount int, queueSize int, logger *slog.Logger) *InMemoryEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &InMemoryEventBus{
		handlers: make(map[string][]secondary.EventHandler),
		logger:   logger,
		queue:    make(chan eventMessage, queueSize),
		workers:  workerCount,
		ctx:      ctx,
		cancel:   cancel,
		started:  false,
	}
}

// Start starts the event bus workers for async processing
func (b *InMemoryEventBus) Start() error {
	b.startMu.Lock()
	defer b.startMu.Unlock()

	if b.started {
		return fmt.Errorf("event bus is already started")
	}

	// Start worker goroutines
	for i := 0; i < b.workers; i++ {
		b.wg.Add(1)
		go b.worker(i)
	}

	b.started = true
	b.logger.Info("Event bus started", "workers", b.workers)
	return nil
}

// Stop stops the event bus gracefully
func (b *InMemoryEventBus) Stop() error {
	b.startMu.Lock()
	defer b.startMu.Unlock()

	if !b.started {
		return nil
	}

	// Signal workers to stop
	b.cancel()

	// Close queue channel to stop accepting new events
	close(b.queue)

	// Wait for workers to finish processing pending events
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		b.logger.Info("Event bus stopped gracefully")
	case <-time.After(10 * time.Second):
		b.logger.Warn("Event bus stop timeout exceeded, forcing shutdown")
	}

	b.started = false
	return nil
}

// worker processes events from the queue
func (b *InMemoryEventBus) worker(id int) {
	defer b.wg.Done()
	b.logger.Debug("Event bus worker started", "worker_id", id)

	stopRequested := false

	for {
		if stopRequested {
			// After stop signal, drain remaining events from queue
			select {
			case msg, ok := <-b.queue:
				if !ok {
					// Channel closed and drained
					b.logger.Debug("Event bus worker exiting (channel closed)", "worker_id", id)
					return
				}
				b.processEvent(msg.ctx, msg.event, id)
			default:
				// No more events immediately available, but channel might not be closed yet
				// Try one more time to read from channel (blocks briefly)
				select {
				case msg, ok := <-b.queue:
					if !ok {
						// Channel closed and fully drained
						b.logger.Debug("Event bus worker exiting (channel closed)", "worker_id", id)
						return
					}
					b.processEvent(msg.ctx, msg.event, id)
				case <-time.After(50 * time.Millisecond):
					// No events for a short time, consider queue drained
					b.logger.Debug("Event bus worker stopping (queue drained)", "worker_id", id)
					return
				}
			}
			continue
		}

		select {
		case <-b.ctx.Done():
			// Context cancelled, start draining queue
			b.logger.Debug("Event bus worker received stop signal, draining queue", "worker_id", id)
			stopRequested = true
		case msg, ok := <-b.queue:
			if !ok {
				// Channel closed, exit
				b.logger.Debug("Event bus worker exiting (channel closed)", "worker_id", id)
				return
			}
			b.processEvent(msg.ctx, msg.event, id)
		}
	}
}

// processEvent processes a single event by notifying all subscribed handlers
func (b *InMemoryEventBus) processEvent(ctx context.Context, event events.DomainEvent, workerID int) {
	eventType := event.EventType()

	b.mu.RLock()
	handlers := make([]secondary.EventHandler, len(b.handlers[eventType]))
	copy(handlers, b.handlers[eventType])
	b.mu.RUnlock()

	if len(handlers) == 0 {
		b.logger.Debug("No handlers registered for event type",
			"worker_id", workerID,
			"event_type", eventType,
			"aggregate_id", event.AggregateID(),
		)
		return
	}

	b.logger.Info("Processing event",
		"worker_id", workerID,
		"event_type", eventType,
		"aggregate_id", event.AggregateID(),
		"handler_count", len(handlers),
	)

	// Process all handlers concurrently using goroutines
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h secondary.EventHandler) {
			defer wg.Done()
			b.executeHandler(ctx, h, event, workerID)
		}(handler)
	}

	wg.Wait()
}

// executeHandler executes a single handler and logs errors without interrupting others
func (b *InMemoryEventBus) executeHandler(ctx context.Context, handler secondary.EventHandler, event events.DomainEvent, workerID int) {
	handlerID := handler.HandlerID()
	
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("Handler panicked",
				"worker_id", workerID,
				"handler_id", handlerID,
				"event_type", event.EventType(),
				"panic", r,
			)
		}
	}()

	err := handler.Handle(ctx, event)
	if err != nil {
		b.logger.Error("Handler execution failed",
			"worker_id", workerID,
			"handler_id", handlerID,
			"event_type", event.EventType(),
			"aggregate_id", event.AggregateID(),
			"error", err,
		)
	} else {
		b.logger.Debug("Handler executed successfully",
			"worker_id", workerID,
			"handler_id", handlerID,
			"event_type", event.EventType(),
		)
	}
}

// Publish publishes a domain event to all subscribed handlers
func (b *InMemoryEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eventType := event.EventType()
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	// If event bus is started (async mode), queue the event
	b.startMu.Lock()
	started := b.started
	b.startMu.Unlock()

	if started {
		select {
		case b.queue <- eventMessage{ctx: ctx, event: event}:
			// Event queued successfully
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Queue is full, log warning but still try to process synchronously
			b.logger.Warn("Event queue is full, processing synchronously",
				"event_type", eventType,
				"aggregate_id", event.AggregateID(),
			)
			// Process synchronously as fallback
			b.processEvent(ctx, event, -1)
		}
	} else {
		// Process synchronously if bus is not started
		b.processEvent(ctx, event, -1)
	}

	return nil
}

// Subscribe subscribes a handler to a specific event type
func (b *InMemoryEventBus) Subscribe(eventType string, handler secondary.EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	handlerID := handler.HandlerID()
	if handlerID == "" {
		return fmt.Errorf("handler ID cannot be empty")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if handler is already subscribed
	handlers := b.handlers[eventType]
	for _, h := range handlers {
		if h.HandlerID() == handlerID {
			return fmt.Errorf("handler with ID '%s' is already subscribed to event type '%s'", handlerID, eventType)
		}
	}

	// Add handler
	b.handlers[eventType] = append(b.handlers[eventType], handler)

	b.logger.Info("Handler subscribed",
		"event_type", eventType,
		"handler_id", handlerID,
		"handler_count", len(b.handlers[eventType]),
	)

	return nil
}

// Unsubscribe removes a handler subscription for a specific event type
func (b *InMemoryEventBus) Unsubscribe(eventType string, handlerID string) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handlerID == "" {
		return fmt.Errorf("handler ID cannot be empty")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	handlers, exists := b.handlers[eventType]
	if !exists {
		return fmt.Errorf("no handlers registered for event type '%s'", eventType)
	}

	// Find and remove handler
	newHandlers := make([]secondary.EventHandler, 0, len(handlers))
	found := false
	for _, h := range handlers {
		if h.HandlerID() != handlerID {
			newHandlers = append(newHandlers, h)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("handler with ID '%s' is not subscribed to event type '%s'", handlerID, eventType)
	}

	b.handlers[eventType] = newHandlers

	b.logger.Info("Handler unsubscribed",
		"event_type", eventType,
		"handler_id", handlerID,
		"remaining_handlers", len(newHandlers),
	)

	return nil
}
