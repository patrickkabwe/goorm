package tests_test

import (
	"context"
	"testing"
)

func TestQueryBuilderSelectWithJoinQuery(t *testing.T) {
	ctx := context.Background()
	u, err := createUser(ctx, qb)
	if err != nil {
		t.Errorf("failed %v", err)
	}

	err = createProfile(ctx, qb, u.ID)
	if err != nil {
		t.Errorf("failed %v", err)
	}

	_, err = qb.
		Select("id", "name", "email", "profiles.id as profile_id", "profiles.user_id").
		From("users").
		LeftJoin("profiles", "users.id = profiles.user_id").
		Where("users.id =$1", u.ID).
		GroupBy("profiles.id", "users.id").
		Limit(1).
		Exec(ctx)

	if err != nil {
		t.Errorf("failed %v", err)
	}

}
