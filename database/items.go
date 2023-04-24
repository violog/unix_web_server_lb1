package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"todo/items"
)

// ErrItems indicates that there was an error in items repository.
var ErrItems = errs.Class("item repository error")

type itemsDB struct {
	conn *sql.DB
}

// Create creates item in the database.
func (itemsDB *itemsDB) Create(ctx context.Context, item items.Item) error {
	query := `INSERT INTO items(id, user_id, name, description, status)
	          VALUES($1,$2,$3,$4,$5)`

	_, err := itemsDB.conn.ExecContext(ctx, query, item.ID, item.UserID, item.Name, item.Description, item.Status)

	return ErrItems.Wrap(err)
}

// List returns all items from the database.
func (itemsDB *itemsDB) List(ctx context.Context, userID uuid.UUID) (_ []items.Item, err error) {
	query := `SELECT id, user_id, name, description, status
	          FROM items
	          WHERE user_id = $1`

	rows, err := itemsDB.conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, ErrItems.Wrap(err)
	}

	defer func() {
		err = errs.Combine(err, rows.Close())
	}()

	var userItems []items.Item
	for rows.Next() {
		var item items.Item
		err = rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Description, &item.Status)
		if err != nil {
			return nil, ErrItems.Wrap(err)
		}

		userItems = append(userItems, item)
	}

	return userItems, ErrItems.Wrap(rows.Err())
}

// Get returns item by id from the database.
func (itemsDB *itemsDB) Get(ctx context.Context, id uuid.UUID) (items.Item, error) {
	var item items.Item
	query := `SELECT id, user_id, name, description, status
	          FROM items
	          WHERE id = $1`

	err := itemsDB.conn.QueryRowContext(ctx, query, id).Scan(&item.ID,
		&item.UserID, &item.Name, &item.Description, &item.Status)
	if errs.Is(err, sql.ErrNoRows) {
		return item, items.ErrNoItem.Wrap(err)
	}

	return item, ErrItems.Wrap(err)
}

// Update updates name and description of item in the database.
func (itemsDB *itemsDB) Update(ctx context.Context, item items.Item) error {
	query := `UPDATE items
	          SET name = $1, description = $2
	          WHERE id = $3`

	res, err := itemsDB.conn.ExecContext(ctx, query, item.Name, item.Description, item.ID)
	if err != nil {
		return ErrItems.Wrap(err)
	}

	rowsCount, err := res.RowsAffected()
	if err == nil && rowsCount == 0 {
		return items.ErrNoItem.New("")
	}

	return ErrItems.Wrap(err)
}

// UpdateStatus updates status of item in the database.
func (itemsDB *itemsDB) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus items.Status) error {
	query := `UPDATE items
	          SET status = $1
	          WHERE id = $2`

	res, err := itemsDB.conn.ExecContext(ctx, query, newStatus, id)
	if err != nil {
		return ErrItems.Wrap(err)
	}

	rowsCount, err := res.RowsAffected()
	if err == nil && rowsCount == 0 {
		return items.ErrNoItem.New("")
	}

	return ErrItems.Wrap(err)
}

// Delete deletes item from the database.
func (itemsDB *itemsDB) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM items
	          WHERE id = $1`

	res, err := itemsDB.conn.ExecContext(ctx, query, id)
	if err != nil {
		return ErrItems.Wrap(err)
	}

	rowsCount, err := res.RowsAffected()
	if err == nil && rowsCount == 0 {
		return items.ErrNoItem.New("")
	}

	return ErrItems.Wrap(err)
}
