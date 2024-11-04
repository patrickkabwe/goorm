package goorm

import (
	"database/sql"
	"strings"
)

// Driver represents the database driver
type Driver string

const (
	Postgres Driver = "pgx"
	Mysql    Driver = "mysql"
	SQlite   Driver = "sqlite"
)

type ColumnInfo struct {
	Name       string
	Type       string
	IsNullable bool
	Default    string
	Extra      string
}

type Column struct {
	Name    string
	Type    string
	Options string
}

type Index struct {
	Name    string
	Columns string
	Last    bool
	Unique  bool
}

type ForeignKey struct {
	Name      string
	Column    string
	RefTable  string
	RefColumn string
	Options   string
	Last      bool
}

type Table struct {
	Name        string
	Columns     []Column
	Indexes     []Index
	ForeignKeys []ForeignKey
}

type IndexOrder struct {
	Name   string
	Table  string
	Unique bool
	Columns string
}

type Migration struct {
	MigrationName string
	Timestamp     string
	Tables        []Table
	DropOrder     []string
	IndexOrder    []IndexOrder
}

// Dialect defines the interface that each database dialect must implement
type Dialect interface {
	// GetPlaceholder returns the placeholder for prepared statements (?, $1, etc)
	GetPlaceholder(index int) string
	// Quote wraps identifiers (table names, column names) with appropriate quotes
	Quote(identifier string) string
	// GetName returns the name of the dialect
	GetName() Driver
	// Check table existence
	TableExists(tx *sql.Tx, tableName string) (bool, error)
	// Get current table columns
	GetColumns(tx *sql.Tx, tableName string) (map[string]ColumnInfo, error)
	// Get current foreign keys
	GetForeignKeys(tx *sql.Tx, tableName string) (map[string]ForeignKey, error)
	// Generate CREATE TABLE statement
	CreateTableSQL(table Table) string
	// Generate CREATE INDEX statement
	CreateIndexSQL(table string, index Index) string
	// Generate ALTER TABLE statement for adding column
	AddColumnSQL(table, column string, info ColumnInfo) string
	// Generate ALTER TABLE statement for modifying column
	ModifyColumnSQL(table, column string, info ColumnInfo) string
	// Generate ALTER TABLE statement for adding foreign key
	AddForeignKeySQL(table string, fk ForeignKey) string
	// Generate ALTER TABLE statement for dropping foreign key
	DropForeignKeySQL(table string, fk ForeignKey) string
	// Convert Go type to SQL type
	SQLType(goType string) string
}

func formatOptions(info interface{}) string {
	switch v := info.(type) {
	case ColumnInfo:
		var opts []string
		if !v.IsNullable {
			opts = append(opts, "NOT NULL")
		}
		if v.Default != "" {
			opts = append(opts, "DEFAULT "+v.Default)
		}
		if v.Extra != "" {
			opts = append(opts, v.Extra)
		}
		if len(opts) > 0 {
			return " " + strings.Join(opts, " ")
		}
	case string:
		if v != "" {
			return " " + v
		}
	}
	return ""
}
