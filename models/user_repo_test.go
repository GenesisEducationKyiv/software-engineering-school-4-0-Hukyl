package models_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/database"
	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/stretchr/testify/assert"
)

func TestUserRepositoryCreate(t *testing.T) {
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user := &models.User{Email: "example@gmail.com"}
	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepositoryFindAll(t *testing.T) {
	// Prepare
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user1 := &models.User{Email: "example1@gmail.com"}
	user2 := &models.User{Email: "example2@gmail.com"}
	err := repo.Create(user1)
	assert.NoError(t, err)
	err = repo.Create(user2)
	assert.NoError(t, err)
	// Act
	users, err := repo.FindAll()
	// Assert
	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepositoryExists(t *testing.T) {
	// Prepare
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user := &models.User{Email: "example@gmail.com"}
	err := repo.Create(user)
	assert.NoError(t, err)
	// Act
	exists, err := repo.Exists(user)
	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepositoryNotExists(t *testing.T) {
	// Prepare
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user := &models.User{Email: "example@gmail.com"}
	// Act
	exists, err := repo.Exists(user)
	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
}
