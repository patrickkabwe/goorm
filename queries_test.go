package goorm_test

import (
	"testing"

	orm "github.com/patrickkabwe/goorm"
	"github.com/stretchr/testify/require"
)

func TestFindMany(t *testing.T) {
	testCreateUser(t)

	users, err := db.User.FindMany(orm.P{
		Where: orm.Where(
			orm.Eq("name", "John"),
		),
	})

	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func testCreateUser(t *testing.T) {
	_, err := db.User.Create(orm.P{
		Data: orm.User{
			Name:  "John",
			Email: "john@example.com",
			Age:   30,
		},
	})

	require.NoError(t, err)
}
