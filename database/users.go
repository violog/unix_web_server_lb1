package database

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"todo/users"
)

// ErrUsers indicates that there was an error in users repository.
var ErrUsers = errs.Class("user repository error")

type usersDB struct {
	conn *sql.DB
}

// Create creates user in the database.
func (usersDB *usersDB) Create(ctx context.Context, user users.User) error {
	query := `INSERT INTO users(id, email, password_hash, created_at)
	          VALUES($1,$2,$3,$4)`

	_, err := usersDB.conn.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.CreatedAt)

	return ErrUsers.Wrap(err)
}

// GetByEmail returns user by email form the database.
func (usersDB *usersDB) GetByEmail(ctx context.Context, email string) (users.User, error) {
	var user users.User
	query := `SELECT id, email, password_hash, created_at
	          FROM users
	          WHERE email = $1`

	err := usersDB.conn.QueryRowContext(ctx, query, email).Scan(&user.ID,
		&user.Email, &user.Password, &user.CreatedAt)
	if errs.Is(err, sql.ErrNoRows) {
		return user, users.ErrNoUser.Wrap(err)
	}

	return user, ErrUsers.Wrap(err)
}

// Delete deletes user from the database.
func (usersDB *usersDB) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users
	          WHERE id = $1`

	res, err := usersDB.conn.ExecContext(ctx, query, id)
	if err != nil {
		return ErrUsers.Wrap(err)
	}

	rowsCount, err := res.RowsAffected()
	if err == nil && rowsCount == 0 {
		return users.ErrNoUser.New("")
	}

	return ErrUsers.Wrap(err)
}
