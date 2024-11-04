package goorm_test

import (
	"context"
	"database/sql"
	"testing"

	orm "github.com/patrickkabwe/goorm"
)

func TestQueryBuilder(t *testing.T) {
	db, err := sql.Open(
		string(orm.Postgres),
		orm.GetENV("POSTGRES_DSN"),
	)

	if err != nil {
		t.Error(err)
	}

	defer db.Close()

	builder := orm.NewQueryBuilder(db, &orm.PostgreSQL{}, nil)
	ctx := context.Background()

	var sd map[string]any
	rows, err := builder.
		SelectDistinct("COUNT(*)", "id", "name", "email", "profiles.id", "profiles.user_id").
		From("users").
		LeftJoin("profiles", "users.id = profiles.user_id").
		Where("id = $1 AND name = $2", 1, "John").
		GroupBy("profiles.id", "users.id").
		Limit(1).
		Offset(1).
		Execute(ctx)

	if err != nil {
		t.Error(err)
	}

	for rows.Next() {
		err := rows.Scan(&sd)
		if err != nil {
			t.Error(err)
		}
	}

}
