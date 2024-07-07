package models_test

import (
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/database"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryCreate(t *testing.T) {
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user := &models.User{Email: "example@gmail.com"}
	err := repo.Create(user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepositoryFindAll(t *testing.T) {
	// Arrange
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user1 := &models.User{Email: "example1@gmail.com"}
	user2 := &models.User{Email: "example2@gmail.com"}
	err := repo.Create(user1)
	require.NoError(t, err)
	err = repo.Create(user2)
	require.NoError(t, err)
	// Act
	users, err := repo.FindAll()
	// Assert
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepositoryExists(t *testing.T) {
	// Arrange
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user := &models.User{Email: "example@gmail.com"}
	err := repo.Create(user)
	require.NoError(t, err)
	// Act
	exists, err := repo.Exists(user)
	// Assert
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepositoryNotExists(t *testing.T) {
	// Arrange
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user := &models.User{Email: "example@gmail.com"}
	// Act
	exists, err := repo.Exists(user)
	// Assert
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepositoryDelete(t *testing.T) {
	// Arrange
	db := database.SetUpTest(t, &models.User{})
	repo := models.NewUserRepository(db)
	user := &models.User{Email: "example@gmail.com"}
	err := repo.Create(user)
	require.NoError(t, err)
	// Act
	err = repo.Delete(user)
	// Assert
	require.NoError(t, err)
	exists, err := repo.Exists(user)
	require.NoError(t, err)
	assert.False(t, exists)
}
