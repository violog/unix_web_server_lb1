package items_test

import (
	"context"
	"testing"
	"time"
	"todo/database"
	"todo/items"
	"todo/users"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
)

func TestItems(t *testing.T) {
	user := users.User{
		ID:        uuid.New(),
		Email:     "testUser@gmail.com",
		Password:  []byte("password"),
		CreatedAt: time.Now().UTC(),
	}

	item1 := items.Item{
		ID:          uuid.New(),
		UserID:      user.ID,
		Name:        "task1",
		Description: "test description",
		Status:      items.StatusTODO,
	}

	item2 := items.Item{
		ID:          uuid.New(),
		UserID:      user.ID,
		Name:        "task2",
		Description: "test description",
		Status:      items.StatusInProgress,
	}

	updatedItem1 := items.Item{
		ID:          item1.ID,
		UserID:      user.ID,
		Name:        "updated name",
		Description: "updated description",
		Status:      items.StatusTODO,
	}

	ctx := context.Background()
	// TODO: create tempdb for tests.
	db, err := database.New("postgres://postgres:123456@localhost/todo?sslmode=disable")
	require.NoError(t, err)

	err = db.CreateSchema(ctx)
	require.NoError(t, err)

	itemsRepository := db.Items()
	usersRepository := db.Users()

	t.Run("create", func(t *testing.T) {
		err = usersRepository.Create(ctx, user)
		require.NoError(t, err)

		err = itemsRepository.Create(ctx, item1)
		require.NoError(t, err)

		err = itemsRepository.Create(ctx, item2)
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		item, err := itemsRepository.Get(ctx, item1.ID)
		require.NoError(t, err)
		compareItems(t, item, item1)
	})

	t.Run("list", func(t *testing.T) {
		allItems, err := itemsRepository.List(ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, len(allItems), 2)
		compareItems(t, allItems[0], item1)
		compareItems(t, allItems[1], item2)
	})

	t.Run("update", func(t *testing.T) {
		err := itemsRepository.Update(ctx, updatedItem1)
		require.NoError(t, err)
	})

	t.Run("update status", func(t *testing.T) {
		err := itemsRepository.UpdateStatus(ctx, item2.ID, items.StatusInProgress)
		require.NoError(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		err := itemsRepository.Delete(ctx, item1.ID)
		require.NoError(t, err)

		err = itemsRepository.Delete(ctx, item2.ID)
		require.NoError(t, err)
	})

	err = db.Close()
	require.NoError(t, err)
}

func compareItems(t *testing.T, item1, item2 items.Item) {
	assert.Equal(t, item1.ID, item2.ID)
	assert.Equal(t, item1.UserID, item2.UserID)
	assert.Equal(t, item1.Name, item2.Name)
	assert.Equal(t, item1.Description, item2.Description)
	assert.Equal(t, item1.Status, item2.Status)
}
