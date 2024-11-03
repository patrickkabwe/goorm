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

func TestFindFirst(t *testing.T) {
	testCreateUser(t)
	user, err := db.User.FindFirst(orm.P{
		Where: orm.Where(
			orm.Eq("name", "John"),
		),
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, user)
}

func TestCreate(t *testing.T) {
	user, err := db.User.Create(orm.P{
		Data: orm.User{
			Name:  "John",
			Email: "john@example.com",
			Age:   30,
		},
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, user)
}

func TestUpdate(t *testing.T) {
	testCreateUser(t)

	err := db.User.Update(orm.P{
		Where: orm.Where(
			orm.Eq("name", "John"),
		),
		Data: orm.User{
			Name: "John Doe",
		},
	})

	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	testCreateUser(t)

	err := db.User.Delete(orm.P{
		Where: orm.Where(
			orm.Eq("id", "20"),
		),
	})

	assert.NoError(t, err)
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
