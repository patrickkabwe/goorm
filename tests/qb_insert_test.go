package tests_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilderInsertClause(t *testing.T) {
	ctx := context.Background()
	email := "test@gmail.com"
	user := &User{}
	err := qb.
		InsertInto("users").
		Columns("name", "email").
		Values("patrick", email).
		Returning(ctx, user, "id", "email")

	if assert.NoError(t, err) {
		assert.Equal(t, user.Email, email)
	}
}
