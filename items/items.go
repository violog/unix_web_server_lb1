package items

import (
	"context"
	"github.com/zeebo/errs"

	"github.com/google/uuid"
)

var ErrNoItem = errs.Class("item does not exist")

// DB is exposing access to items db.
type DB interface {
	// Create creates item in the database.
	Create(ctx context.Context, item Item) error
	// List returns all items from the database.
	List(ctx context.Context, userID uuid.UUID) ([]Item, error)
	// Get returns item by id from the database.
	Get(ctx context.Context, id uuid.UUID) (Item, error)
	// Update updates name and description of item in the database.
	Update(ctx context.Context, item Item) error
	// UpdateStatus updates status of item in the database.
	UpdateStatus(ctx context.Context, id uuid.UUID, newStatus Status) error
	// Delete deletes item from the database.
	Delete(ctx context.Context, id uuid.UUID) error
}

// Item defines item list.
type Item struct {
	ID          uuid.UUID `bson:"id"`
	UserID      uuid.UUID `json:"userId"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Status      Status    `bson:"status"`
}

// Status defines list of possible statuses of items.
type Status string

const (
	// StatusTODO defines type of status of item which haven't started.
	StatusTODO Status = "TODO"
	// StatusInProgress defines type of status of item which in progress.
	StatusInProgress Status = "In Progress"
	// StatusCompleted defines type of status of item which is completed.
	StatusCompleted Status = "Completed"
)
