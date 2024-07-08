package models_test

import (
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriberRepositoryCreate(t *testing.T) {
	// Arrange
	t.Parallel()
	db := database.SetUpTest(t, &models.Subscriber{})
	repo := models.NewSubscriberRepository(db)
	subscriber := &models.Subscriber{Email: "test@gmail.com"}
	// Act
	err := repo.Create(subscriber)
	// Assert
	require.NoError(t, err)
	assert.NotZero(t, subscriber.ID)
	assert.NotNil(t, subscriber.Created)
}

func TestSubscriberRepositoryFindAll(t *testing.T) {
	// Arrange
	t.Parallel()
	db := database.SetUpTest(t, &models.Subscriber{})
	repo := models.NewSubscriberRepository(db)
	subscribers := []*models.Subscriber{
		{Email: "test@gmail.com"},
		{Email: "test2@gmail.com"},
	}
	for _, subscriber := range subscribers {
		require.NoError(t, repo.Create(subscriber))
	}
	// Act
	foundSubscribers, err := repo.FindAll()
	// Assert
	require.NoError(t, err)
	assert.Equal(t, len(subscribers), len(foundSubscribers))
	for _, subscriber := range subscribers {
		for _, foundSubscriber := range foundSubscribers {
			if subscriber.Email == foundSubscriber.Email {
				assert.Equal(t, subscriber.Email, foundSubscriber.Email)
				assert.Equal(t, subscriber.Created, foundSubscriber.Created)
				break
			}
		}
	}
}

func TestSubscriberRepositoryExists(t *testing.T) {
	// Arrange
	t.Parallel()
	db := database.SetUpTest(t, &models.Subscriber{})
	repo := models.NewSubscriberRepository(db)
	subscriber := &models.Subscriber{Email: "example@gmail.com"}
	require.NoError(t, repo.Create(subscriber))
	// Act
	exists, err := repo.Exists(subscriber)
	// Assert
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestSubscriberRepositoryDelete(t *testing.T) {
	// Arrange
	t.Parallel()
	db := database.SetUpTest(t, &models.Subscriber{})
	repo := models.NewSubscriberRepository(db)
	subscriber := &models.Subscriber{Email: "example@gmail.com"}
	require.NoError(t, repo.Create(subscriber))
	// Act
	err := repo.Delete(subscriber)
	// Assert
	require.NoError(t, err)
	exists, err := repo.Exists(subscriber)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestSubscriberRepositoryDeleteNonExisting(t *testing.T) {
	// Arrange
	t.Parallel()
	db := database.SetUpTest(t, &models.Subscriber{})
	repo := models.NewSubscriberRepository(db)
	subscriber := &models.Subscriber{Email: "example@gmail.com"}
	// Act
	err := repo.Delete(subscriber)
	// Assert
	require.NoError(t, err)
	exists, err := repo.Exists(subscriber)
	require.NoError(t, err)
	assert.False(t, exists)
}
