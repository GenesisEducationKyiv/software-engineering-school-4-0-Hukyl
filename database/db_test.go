package database_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/database"
	"github.com/Hukyl/genesis-kma-school-entry/database/config"
	"github.com/stretchr/testify/assert"
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

func TestInvalidConfig(t *testing.T) {
	config := config.Config{
		DatabaseService: "unknown",
		DatabaseDSN:     "unknown",
	}
	db, err := database.New(config)
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestInit(t *testing.T) {
	config := config.Config{
		DatabaseService: "sqlite",
		DatabaseDSN:     "file::memory:?cache=shared",
	}
	db := database.DB{Config: config}
	err := db.Init()
	assert.NoError(t, err)
}

func TestInitFail(t *testing.T) {
	config := config.Config{
		DatabaseService: "unknown",
		DatabaseDSN:     "file::memory:?cache=shared",
	}
	db := database.DB{Config: config}
	err := db.Init()
	assert.Error(t, err)
}

func TestMigrateNull(t *testing.T) {
	db := database.SetUpTest(t)
	err := db.Migrate()
	assert.NoError(t, err)
}

func TestMigrateModels(t *testing.T) {
	db := database.SetUpTest(t)
	err := db.Migrate(&MockUser{})
	assert.NoError(t, err)
}

func TestMigrateModelsMultipleTimes(t *testing.T) {
	db := database.SetUpTest(t)
	err := db.Migrate(&MockUser{})
	assert.NoError(t, err)
	err = db.Migrate(&MockUser{})
	assert.NoError(t, err)
}

func TestMigrateEmptyModel(t *testing.T) {
	db := database.SetUpTest(t)
	err := db.Migrate(&EmptyStruct{})
	assert.Error(t, err)
}

func TestNew(t *testing.T) {
	config := config.Config{
		DatabaseService: "sqlite",
		DatabaseDSN:     "file::memory:?cache=shared",
	}
	db, err := database.New(config)
	assert.NotNil(t, db)
	assert.NoError(t, err)
}

func TestNewFail(t *testing.T) {
	config := config.Config{
		DatabaseService: "123",
		DatabaseDSN:     "file::memory:?cache=shared",
	}
	db, err := database.New(config)
	assert.Nil(t, db)
	assert.Error(t, err)
}
