package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents something that happened in the domain.
// Events are immutable records of business-significant occurrences.
type DomainEvent interface {
	// EventID returns the unique identifier for this event occurrence.
	EventID() uuid.UUID

	// EventType returns the type name of the event (e.g., "TransactionCompleted").
	EventType() string

	// OccurredAt returns when the event occurred.
	OccurredAt() time.Time

	// AggregateID returns the ID of the aggregate that raised this event.
	AggregateID() uuid.UUID

	// AggregateType returns the type of aggregate (e.g., "Transaction").
	AggregateType() string
}

// BaseEvent provides a default implementation of common event fields.
type BaseEvent struct {
	eventID       uuid.UUID
	eventType     string
	occurredAt    time.Time
	aggregateID   uuid.UUID
	aggregateType string
}

// NewBaseEvent creates a new base event.
func NewBaseEvent(eventType string, aggregateID uuid.UUID, aggregateType string) BaseEvent {
	return BaseEvent{
		eventID:       uuid.New(),
		eventType:     eventType,
		occurredAt:    time.Now().UTC(),
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
	}
}

func (e BaseEvent) EventID() uuid.UUID       { return e.eventID }
func (e BaseEvent) EventType() string         { return e.eventType }
func (e BaseEvent) OccurredAt() time.Time     { return e.occurredAt }
func (e BaseEvent) AggregateID() uuid.UUID    { return e.aggregateID }
func (e BaseEvent) AggregateType() string     { return e.aggregateType }
