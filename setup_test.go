package goorm_test

import (
	"os"
	"testing"

	"github.com/patrickkabwe/goorm"
)

var db *goorm.GeneratedDB

func TestMain(m *testing.M) {
	// Setup database
	db = goorm.NewGoorm(&goorm.GoormConfig{
		DSN:    "postgres://postgres:postgres@localhost:5432/goorm-test?sslmode=disable",
		Driver: goorm.Postgres,
		Logger: goorm.NewDefaultLogger(),
	})

	// Run tests
	m.Run()

	// Close database
	db.Close()
	os.Exit(0)
}
