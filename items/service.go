package items

import (
	"context"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

// Error indicates that there was an error in items service.
var Error = errs.Class("items service error")

// Service is handling items related logic.
type Service struct {
	items DB
}

// New is constructor for Service.
func New(items DB) *Service {
	return &Service{
		items: items,
	}
}

// Create creates item.
func (service *Service) Create(ctx context.Context, userID uuid.UUID, name, description string) error {
	item := Item{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		Description: description,
		Status:      StatusTODO,
	}

	return Error.Wrap(service.items.Create(ctx, item))
}

// List returns all items.
func (service *Service) List(ctx context.Context, userID uuid.UUID) ([]Item, error) {
	items, err := service.items.List(ctx, userID)

	return items, Error.Wrap(err)
}

// Get returns item by id.
func (service *Service) Get(ctx context.Context, id uuid.UUID) (Item, error) {
	item, err := service.items.Get(ctx, id)

	return item, Error.Wrap(err)
}

// Update changes name and description of item.
func (service *Service) Update(ctx context.Context, id uuid.UUID, name, description string) error {
	return Error.Wrap(service.items.Update(ctx, Item{
		ID:          id,
		Name:        name,
		Description: description,
	}))
}

// UpdateStatus updates status of certain item.
func (service *Service) UpdateStatus(ctx context.Context, id uuid.UUID) error {
	item, err := service.Get(ctx, id)
	if err != nil {
		return Error.Wrap(err)
	}

	switch item.Status {
	case StatusTODO:
		item.Status = StatusInProgress
	case StatusInProgress:
		item.Status = StatusCompleted
	case StatusCompleted:
		return Error.New("could not update status of item")
	}

	return Error.Wrap(service.items.UpdateStatus(ctx, id, item.Status))
}

// Delete deletes certain item.
func (service *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return Error.Wrap(service.items.Delete(ctx, id))
}
