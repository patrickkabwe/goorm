package tests_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	orm "github.com/patrickkabwe/goorm"
)

var db *sql.DB
var qb *orm.QueryBuilder

type User struct {
	ID      int64    `db:"id"`
	Name    string   `db:"name"`
	Email   string   `db:"email"`
	Profile *Profile `db:"profiles"`
}

type Profile struct {
	ID     int64  `db:"id"`
	Avatar string `db:"avatar"`
	UserID int64  `db:"user_id"`
}

func TestMain(m *testing.M) {
	// Setup database
	var err error
	db, err = sql.Open(
		string(orm.Postgres),
		os.Getenv("POSTGRES_DSN"),
	)

	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	qb = orm.NewQueryBuilder(db, &orm.PostgreSQL{}, nil)
	err = createTables(db)
	if err != nil {
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	ctx := context.Background()
	_, err = qb.
		Truncate("users").
		RestartIdentity().
		Cascade().
		Exec(ctx)

	if err != nil {
		log.Fatalln(err)
	}
	// // Close database
	db.Close()
	os.Exit(code)
}

func createTables(db *sql.DB) error {
	query := `
		create table if not exists users(
			id serial primary key,
			email varchar(255) not null,
			name varchar(255) not null
		);

		create table if not exists profiles(
			id serial primary key,
			avatar varchar(255) not null,
			user_id integer,
			constraint fk_user_id foreign key(user_id) references users(id) on delete cascade
		);

		create table if not exists posts(
			id serial primary key,
			body text not null,
			user_id integer,
			constraint fk_user_id foreign key(user_id) references users(id) on delete cascade
		);

		create table if not exists comments(
			id serial primary key,
			comment text not null,
			user_id integer,
			post_id integer,
			constraint fk_user_id foreign key(user_id) references users(id) on delete cascade,
			constraint fk_post_id foreign key(post_id) references posts(id) on delete cascade
		);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func createUser(ctx context.Context, builder *orm.QueryBuilder) (*User, error) {
	user := &User{}
	err := builder.
		InsertInto("users").
		Columns("name", "email").
		Values("patrick", "patrick@gail.com").
		Returning(ctx, user, "name", "id")

	return user, err
}

func createProfile(ctx context.Context, builder *orm.QueryBuilder, userID int64) error {
	profile := &Profile{}
	err := builder.
		InsertInto("profiles").
		Columns("user_id", "avatar").
		Values(userID, "some_url").
		Returning(ctx, profile)

	return err
}
