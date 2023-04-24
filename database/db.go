package database

import (
	"context"
	"database/sql"
	"todo"
	"todo/items"
	"todo/users"

	_ "github.com/lib/pq" // using postgres driver.
	"github.com/zeebo/errs"
)

// Error indicates that there was an error in database.
var Error = errs.Class("database error")

// ensures that database implements todo.DB.
var _ todo.DB = (*database)(nil)

// database combines access to different database tables with a record
// of the db driver, db implementation, and db source URL.
//
// architecture: Master Database
type database struct {
	conn *sql.DB
}

// New returns todo.DB postgresql implementation.
func New(databaseURL string) (todo.DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &database{conn: conn}, nil
}

// CreateSchema create schema for all tables and databases.
func (db *database) CreateSchema(ctx context.Context) (err error) {
	createTableQuery :=
		`CREATE TABLE IF NOT EXISTS users (
            id            BYTEA     PRIMARY KEY    NOT NULL,
            email         VARCHAR                  NOT NULL,
            password_hash BYTEA                    NOT NULL,
            created_at    TIMESTAMP WITH TIME ZONE NOT NULL
        );
        CREATE TABLE IF NOT EXISTS items (
            id          BYTEA   PRIMARY KEY                            NOT NULL,
            user_id     BYTEA   REFERENCES users(id) ON DELETE CASCADE NOT NULL,
            name        VARCHAR                                        NOT NULL,
            description VARCHAR                                        NOT NULL,
            status      VARCHAR                                        NOT NULL
        );`

	_, err = db.conn.ExecContext(ctx, createTableQuery)
	if err != nil {
		return Error.Wrap(err)
	}

	return nil
}

// Close closes underlying db connection.
func (db *database) Close() error {
	return Error.Wrap(db.conn.Close())
}

// Users provides access to accounts db.
func (db *database) Users() users.DB {
	return &usersDB{conn: db.conn}
}

// Items provides access to accounts db.
func (db *database) Items() items.DB {
	return &itemsDB{conn: db.conn}
}