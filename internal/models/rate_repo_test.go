package models_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/internal/database"
	"github.com/Hukyl/genesis-kma-school-entry/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateRepositoryCreate(t *testing.T) {
	db := database.SetUpTest(t, &models.Rate{})
	repo := models.NewRateRepository(db)
	rate := &models.Rate{CurrencyFrom: "USD", CurrencyTo: "UAH", Rate: 27.5}
	err := repo.Create(rate)
	require.NoError(t, err)
	assert.NotZero(t, rate.ID)
	assert.NotNil(t, rate.Created)
}
