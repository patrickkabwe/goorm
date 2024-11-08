package goorm_test

import (
	"os"
	"testing"

	"github.com/patrickkabwe/goorm"
)

var db *goorm.GeneratedDB

func TestMain(m *testing.M) {
	// Setup database
	var err error
	db, err = goorm.NewGoorm(&goorm.GoormConfig{
		DSN:    goorm.GetENV("POSTGRES_DSN"),
		Driver: goorm.Postgres,
		Logger: goorm.NewDefaultLogger(),
	})

	if err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	_ = db.User.Delete(goorm.P{})

	// Close database
	db.Close()
	os.Exit(code)
}
