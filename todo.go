package todo

import (
	"context"
	"errors"
	"net"

	"github.com/zeebo/errs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"todo/console"
	"todo/items"
	"todo/pkg/auth"
	"todo/users"
	"todo/users/userauth"
)

// DB provides access to all databases and database related functionality.
type DB interface {
	// Users provides access to users db.
	Users() users.DB

	// Items provides access to items db.
	Items() items.DB

	// CreateSchema creates db schema.
	CreateSchema(ctx context.Context) error

	// Close closes connection with db.
	Close() error
}

// Todo represents project.
type Todo struct {
	Logger   *zap.Logger
	Database DB

	Users struct {
		Service *users.Service
		Auth    *userauth.Service
	}

	Items struct {
		Service *items.Service
	}

	// Admin web server with web UI.
	Console struct {
		Listener net.Listener
		Endpoint *console.Server
	}
}

// New is constructor for Todo.
func New(database DB) (*Todo, error) {
	todo := &Todo{Database: database}
	todo.Logger, _ = zap.NewProduction()
	err := database.CreateSchema(context.Background())
	if err != nil {
		return todo, err
	}

	{
		todo.Users.Service = users.New(
			todo.Database.Users(),
		)

		todo.Users.Auth = userauth.NewService(
			todo.Database.Users(),
			auth.TokenSigner{
				Secret: []byte("secret-token"),
			},
		)
	}

	{
		todo.Items.Service = items.New(
			todo.Database.Items(),
		)
	}

	{ // console setup
		todo.Console.Listener, err = net.Listen("tcp", ":8087")
		if err != nil {
			return nil, err
		}

		todo.Console.Endpoint, err = console.NewServer(
			todo.Console.Listener,
			todo.Users.Auth,
			todo.Logger,
			todo.Items.Service,
			todo.Users.Service,
		)
	}

	return todo, nil
}

// Run runs app servers as separate goroutine.
func (todo *Todo) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return ignoreCancel(todo.Console.Endpoint.Run(ctx))
	})

	return group.Wait()
}

// Close closes all the resources.
func (todo *Todo) Close() error {
	var errlist errs.Group

	errlist.Add(todo.Console.Endpoint.Close())

	return errlist.Err()
}

// we ignore cancellation and stopping errors since they are expected.
func ignoreCancel(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
