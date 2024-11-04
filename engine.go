package goorm

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

// Engine represents the core ORM engine
type Engine struct {
	db      *sql.DB
	logger  Logger
	dialect Dialect
}

func getDialect(driverName Driver) (Dialect, error) {
	var dialect Dialect

	switch driverName {
	case Mysql:
		dialect = &MYSQL{}
	case Postgres:
		dialect = &PostgreSQL{}
	case SQlite:
		dialect = &SQLite{}
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driverName)
	}

	return dialect, nil
}

// NewEngine creates a new ORM engine with the specified dialect and connection
func NewEngine(driverName Driver, dataSource string, logger Logger) (*Engine, error) {
	dialect, err := getDialect(driverName)

	if err != nil {
		return nil, err
	}

	db, err := sql.Open(string(dialect.GetName()), dataSource)

	if err != nil {
		return nil, err
	}

	return &Engine{
		db:      db,
		dialect: dialect,
		logger:  logger,
	}, nil
}

func (e *Engine) Close() {
	e.db.Close()
}

// Example usage of the engine for basic CRUD operations
func (e *Engine) Create(ctx context.Context, value interface{}) error {
	// Implementation for creating records
	return nil
}

func (e *Engine) Find(ctx context.Context, dest interface{}, where ...interface{}) error {
	// Implementation for finding records
	return nil
}

func (e *Engine) Update(ctx context.Context, value interface{}) error {
	// Implementation for updating records
	return nil
}

func (e *Engine) Delete(ctx context.Context, value interface{}) error {
	// Implementation for deleting records
	return nil
}
