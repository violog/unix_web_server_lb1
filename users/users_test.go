package users_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo/database"
	"todo/users"
)

func TestUsers(t *testing.T) {
	user := users.User{
		ID:        uuid.New(),
		Email:     "testUser@gmail.com",
		Password:  []byte("password"),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	// TODO: create tempdb for tests.
	db, err := database.New("postgres://postgres:123456@localhost/todo?sslmode=disable")
	require.NoError(t, err)

	err = db.CreateSchema(ctx)
	require.NoError(t, err)

	usersRepository := db.Users()

	t.Run("create", func(t *testing.T) {
		err = usersRepository.Create(ctx, user)
		require.NoError(t, err)
	})

	t.Run("get by email", func(t *testing.T) {
		userFromDB, err := usersRepository.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		compareUsers(t, userFromDB, user)
	})

	t.Run("delete", func(t *testing.T) {
		err = usersRepository.Delete(ctx, user.ID)
		require.NoError(t, err)
	})

	err = db.Close()
	require.NoError(t, err)
}

func compareUsers(t *testing.T, user1, user2 users.User) {
	assert.Equal(t, user1.ID, user2.ID)
	assert.Equal(t, user1.Email, user2.Email)
	assert.Equal(t, user1.Password, user2.Password)
	assert.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, 1*time.Second)
}
