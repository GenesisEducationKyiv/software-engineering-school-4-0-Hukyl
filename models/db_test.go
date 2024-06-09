package models_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/stretchr/testify/assert"
)

func TestSingletonDBConnection(t *testing.T) {
	db := models.SetUpTestDB(t)

	conn1 := db.Connection()
	conn2 := db.Connection()
	assert.Equal(t, conn1, conn2)
}
