package database_test

import (
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/database"
	"github.com/stretchr/testify/assert"
)

func TestSingletonDBConnection(t *testing.T) {
	db := database.SetUpTest(t)

	conn1 := db.Connection()
	conn2 := db.Connection()
	assert.Equal(t, conn1, conn2)
}
