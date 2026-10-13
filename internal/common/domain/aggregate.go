package domain

import (
	"time"

	"github.com/google/uuid"
)

// AggregateRoot is the base for all aggregate roots in the domain.
// It tracks domain events and provides common fields for auditing.
type AggregateRoot struct {
	id        uuid.UUID
	events    []DomainEvent
	version   int
	createdAt time.Time
	updatedAt time.Time
}

// NewAggregateRoot creates a new aggregate root with a generated ID.
func NewAggregateRoot() AggregateRoot {
	now := time.Now().UTC()
	return AggregateRoot{
		id:        uuid.New(),
		events:    make([]DomainEvent, 0),
		version:   0,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructAggregateRoot reconstructs an aggregate root from persistence.
func ReconstructAggregateRoot(id uuid.UUID, version int, createdAt, updatedAt time.Time) AggregateRoot {
	return AggregateRoot{
		id:        id,
		events:    make([]DomainEvent, 0),
		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// ID returns the aggregate's unique identifier.
func (a *AggregateRoot) ID() uuid.UUID {
	return a.id
}

// Version returns the aggregate's current version for optimistic concurrency.
func (a *AggregateRoot) Version() int {
	return a.version
}

// CreatedAt returns when the aggregate was first created.
func (a *AggregateRoot) CreatedAt() time.Time {
	return a.createdAt
}

// UpdatedAt returns when the aggregate was last modified.
func (a *AggregateRoot) UpdatedAt() time.Time {
	return a.updatedAt
}

// AddEvent records a domain event to be dispatched after persistence.
func (a *AggregateRoot) AddEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

// Events returns all pending domain events.
func (a *AggregateRoot) Events() []DomainEvent {
	return a.events
}

// ClearEvents clears all pending domain events (called after dispatch).
func (a *AggregateRoot) ClearEvents() {
	a.events = make([]DomainEvent, 0)
}

// Touch updates the updatedAt timestamp and increments the version.
func (a *AggregateRoot) Touch() {
	a.updatedAt = time.Now().UTC()
	a.version++
}
