package database_test

import (
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/database"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/database/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockUser struct {
	Email    string
	Password string
}

type EmptyStruct struct{}

func TestSingletonDBConnection(t *testing.T) {
	db := database.SetUpTest(t)

	conn1 := db.Connection()
	conn2 := db.Connection()
	assert.Equal(t, conn1, conn2)
}

func TestInit(t *testing.T) {
	testCases := []struct {
		name        string
		config      config.Config
		expectError bool
	}{
		{
			name: "sqlite",
			config: config.Config{
				DatabaseService: "sqlite",
				DatabaseDSN:     "file::memory:?cache=shared",
			},
			expectError: false,
		},
		{
			name: "unknown",
			config: config.Config{
				DatabaseService: "unknown",
				DatabaseDSN:     "unknown",
			},
			expectError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := database.DB{Config: tc.config}
			err := db.Init()
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestMigrate(t *testing.T) {
	testCases := []struct {
		name        string
		models      []interface{}
		expectError bool
	}{
		{
			name:        "null",
			models:      nil,
			expectError: false,
		},
		{
			name:        "models",
			models:      []interface{}{&MockUser{}},
			expectError: false,
		},
		{
			name:        "same-models-twice",
			models:      []interface{}{&MockUser{}, &MockUser{}},
			expectError: false,
		},
		{
			name:        "empty",
			models:      []interface{}{&EmptyStruct{}},
			expectError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := database.SetUpTest(t)
			err := db.Migrate(tc.models...)
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNewFail(t *testing.T) {
	testCases := []struct {
		name        string
		config      config.Config
		expectError bool
	}{
		{
			name: "sqlite",
			config: config.Config{
				DatabaseService: "sqlite",
				DatabaseDSN:     "file::memory:?cache=shared",
			},
			expectError: false,
		},
		{
			name: "invalid-service",
			config: config.Config{
				DatabaseService: "unknown",
				DatabaseDSN:     "file::memory:?cache=shared",
			},
			expectError: true,
		},
		{
			name: "unknown",
			config: config.Config{
				DatabaseService: "unknown",
				DatabaseDSN:     "unknown",
			},
			expectError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := database.New(tc.config)
			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, db)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, db)
		})
	}
}
