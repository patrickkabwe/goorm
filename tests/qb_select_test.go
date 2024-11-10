package tests_test

import (
	"context"
	"testing"
)

func TestQueryBuilderSelectQuery(t *testing.T) {
	ctx := context.Background()
	u, err := createUser(ctx, qb)
	if err != nil {
		t.Errorf("failed %v", err)
	}

	_, err = qb.
		Select("id", "name", "email").
		From("users").
		Where("users.id = $1", u.ID).
		Limit(1).
		Exec(ctx)

	if err != nil {
		t.Errorf("failed %v", err)
	}
}
