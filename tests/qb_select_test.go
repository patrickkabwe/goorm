package tests_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilderSelectQuery(t *testing.T) {
	ctx := context.Background()
	u, err := createUser(ctx, qb)
	if err != nil {
		t.Errorf("failed %v", err)
	}
	user := &User{}
	err = qb.
		Select("id", "name", "email").
		From("users").
		Where("users.id = $1", u.ID).
		Limit(1).
		Scan(ctx, user)

	if assert.NoError(t, err) {
		assert.NotEmpty(t, user.Name)
	}
}
