package tests_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilderUpdateClause(t *testing.T) {
	ctx := context.Background()
	u, err := createUser(ctx, qb)
	if err != nil {
		t.Errorf("failed %v", err)
	}
	assert.NoError(t, err)
	name := "patrick101"
	user := &User{}
	err = qb.
		Update("users").
		Set(
			map[string]string{
				"name": name,
			},
		).
		Where(fmt.Sprintf("id = %s", qb.Dialect.GetPlaceholder(2)), u.ID).
		Returning(ctx, user, "name")

	if assert.NoError(t, err) {
		assert.Equal(t, user.Name, name)
	}
}
