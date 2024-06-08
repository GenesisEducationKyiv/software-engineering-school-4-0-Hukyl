package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/utils"
)

func TestSingletonDBConnection(t *testing.T) {
	utils.SetUpTestDB()
	defer utils.TearDownTestDB()

	db := models.NewDB()
	conn1 := db.Connection()
	conn2 := db.Connection()
	assert.Equal(t, conn1, conn2)
	db2 := models.NewDB()
	conn3 := db2.Connection()
	assert.Equal(t, conn1, conn3)
}
