package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

// ErrNoUser indicates that user does not exists.
var ErrNoUser = errs.Class("user does not exists")

// DB is exposing access to users db.
type DB interface {
	// Create creates user in the database.
	Create(ctx context.Context, user User) error
	// GetByEmail returns user by email form the database.
	GetByEmail(ctx context.Context, email string) (User, error)
	// Delete deletes user from the database.
	Delete(ctx context.Context, id uuid.UUID) error
}

// User describes user entity.
type User struct {
	ID        uuid.UUID `bson:"id"`
	Email     string    `bson:"email"`
	Password  []byte    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
}
