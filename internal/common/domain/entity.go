package domain

import (
	"time"

	"github.com/google/uuid"
)

// Entity is the base for all entities within an aggregate.
// Entities have identity but are not aggregate roots.
type Entity struct {
	id        uuid.UUID
	createdAt time.Time
	updatedAt time.Time
}

// NewEntity creates a new entity with a generated ID.
func NewEntity() Entity {
	now := time.Now().UTC()
	return Entity{
		id:        uuid.New(),
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructEntity reconstructs an entity from persistence.
func ReconstructEntity(id uuid.UUID, createdAt, updatedAt time.Time) Entity {
	return Entity{
		id:        id,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// ID returns the entity's unique identifier.
func (e *Entity) ID() uuid.UUID {
	return e.id
}

// CreatedAt returns when the entity was first created.
func (e *Entity) CreatedAt() time.Time {
	return e.createdAt
}

// UpdatedAt returns when the entity was last modified.
func (e *Entity) UpdatedAt() time.Time {
	return e.updatedAt
}

// Touch updates the updatedAt timestamp.
func (e *Entity) Touch() {
	e.updatedAt = time.Now().UTC()
}
