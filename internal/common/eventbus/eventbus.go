package eventbus

import (
	"context"
	"sync"

	"github.com/hanjo-erp/hanjo/internal/common/domain"
)

// EventHandler is a function that handles a domain event.
type EventHandler func(ctx context.Context, event domain.DomainEvent) error

// EventBus is the interface for publishing and subscribing to domain events.
type EventBus interface {
	// Publish dispatches a domain event to all registered handlers.
	Publish(ctx context.Context, event domain.DomainEvent) error

	// PublishAll dispatches multiple domain events.
	PublishAll(ctx context.Context, events []domain.DomainEvent) error

	// Subscribe registers a handler for a specific event type.
	Subscribe(eventType string, handler EventHandler)
}

// InMemoryEventBus is a synchronous in-memory implementation of EventBus.
// Suitable for modular monolith (MVP). Can be replaced with NATS/RabbitMQ later.
type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

// NewInMemoryEventBus creates a new in-memory event bus.
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish dispatches a single domain event to all registered handlers synchronously.
func (bus *InMemoryEventBus) Publish(ctx context.Context, event domain.DomainEvent) error {
	bus.mu.RLock()
	handlers, ok := bus.handlers[event.EventType()]
	bus.mu.RUnlock()

	if !ok {
		return nil
	}

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// PublishAll dispatches multiple domain events in order.
func (bus *InMemoryEventBus) PublishAll(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := bus.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe registers a handler for the given event type.
func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
}
