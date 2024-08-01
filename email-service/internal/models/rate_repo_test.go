package models_test

import (
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/database"
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
