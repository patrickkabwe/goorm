package goorm_test

import (
	"testing"

	orm "github.com/patrickkabwe/goorm"
	"github.com/stretchr/testify/assert"
)

func TestFindMany(t *testing.T) {
	testCreateUser(t)

	users, err := db.User.FindMany(orm.P{
		Where: orm.Where(
			orm.Eq("name", "John"),
		),
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, users)
}

func testCreateUser(t *testing.T) {
	_, err := db.User.Create(orm.P{
		Data: orm.User{
			Name:  "John",
			Email: "john@example.com",
			Age:   30,
		},
	})

	assert.NoError(t, err)
}
