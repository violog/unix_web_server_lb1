package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"
)

// Error indicates that there was an error in users service.
var Error = errs.Class("users service error")

// Service is handling users related logic.
type Service struct {
	users DB
}

// New is constructor for Service.
func New(users DB) *Service {
	return &Service{
		users: users,
	}
}

// Create creates user.
func (service *Service) Create(ctx context.Context, email, password string) error {
	user := User{
		ID:        uuid.New(),
		Email:     email,
		Password:  []byte(password),
		CreatedAt: time.Now(),
	}

	err := user.EncodePass()
	if err != nil {
		return Error.Wrap(err)
	}

	return Error.Wrap(service.users.Create(ctx, user))
}

// GetByEmail returns user by email.
func (service *Service) GetByEmail(ctx context.Context, email string) (User, error) {
	user, err := service.users.GetByEmail(ctx, email)

	return user, Error.Wrap(err)
}

// Delete deletes user.
func (service *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return Error.Wrap(service.users.Delete(ctx, id))
}

// EncodePass encode the password and generate "hash" to store from users password.
func (user *User) EncodePass() error {
	hash, err := bcrypt.GenerateFromPassword(user.Password, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = hash
	return nil
}
