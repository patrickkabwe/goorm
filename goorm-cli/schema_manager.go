package main

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/patrickkabwe/goorm"
)

type SchemaManager struct {
	db      *sql.DB
	dialect goorm.Dialect
}

func NewSchemaPusher(db *sql.DB, driverName goorm.Driver) *SchemaManager {
	var dialect goorm.Dialect
	switch driverName {
	case goorm.Mysql:
		dialect = &goorm.MYSQL{}
	case goorm.Postgres:
		dialect = &goorm.PostgreSQL{}
	case goorm.SQlite:
		dialect = &goorm.SQLite{}
	default:
		panic("unsupported dialect: " + string(driverName))
	}

	return &SchemaManager{
		db:      db,
		dialect: dialect,
	}
}

func (s *SchemaManager) Push() error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "../models.go", nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse models file: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	structNames := make(map[string]bool)
	ast.Inspect(f, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				structNames[typeSpec.Name.Name] = true
			}
		}
		return true
	})

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if err := s.processTable(tx, typeSpec.Name.Name, structType); err != nil {
				return fmt.Errorf("process table %s: %w", typeSpec.Name.Name, err)
			}
		}
	}

	return tx.Commit()
}

func (s *SchemaManager) Migrate(migrationName string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "../models.go", nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse models.go: %v", err))
	}

	var migration goorm.Migration
	migration.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	// First pass: collect all struct names for relationship resolution
	structNames := make(map[string]bool)
	ast.Inspect(f, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				structNames[typeSpec.Name.Name] = true
			}
		}
		return true
	})

	// Second pass: process structs with relationship knowledge
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			table := buildTable(s.dialect, typeSpec.Name.Name, structType)
			migration.Tables = append(migration.Tables, table)
			for _, idx := range table.Indexes {
				migration.IndexOrder = append(migration.IndexOrder, goorm.IndexOrder{
					Name:    idx.Name,
					Unique:  idx.Unique,
					Table:   table.Name,
					Columns: idx.Columns,
				})
			}
			migration.DropOrder = append([]string{table.Name}, migration.DropOrder...)
		}
	}

	// Create migrations directory if it doesn't exist
	migrationsDir := "../migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create migrations directory: %v", err))
	}

	// Generate migration filename with timestamp
	timestamp := time.Now().Format("20060102150405")
	migrationName = fmt.Sprintf("%s_%s.sql", timestamp, migrationName)
	migration.MigrationName = migrationName
	filename := filepath.Join(migrationsDir, migrationName)

	// Generate the migration file
	tmpl := template.Must(template.New("migration").Parse(migrationTemplate))

	out, err := os.Create(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to create migration file: %v", err))
	}
	defer out.Close()

	if err := tmpl.Execute(out, migration); err != nil {
		panic(fmt.Sprintf("Failed to execute template: %v", err))
	}

	return nil
}

func (s *SchemaManager) processTable(tx *sql.Tx, structName string, structType *ast.StructType) error {
	tableName := toSnakeCase(structName) + "s"

	exists, err := s.dialect.TableExists(tx, tableName)
	if err != nil {
		return err
	}

	if !exists {
		table := buildTable(s.dialect, structName, structType)
		sql := s.dialect.CreateTableSQL(table)
		if _, err := tx.Exec(sql); err != nil {
			return err
		}
		for _, idx := range table.Indexes {
			indexSql := s.dialect.CreateIndexSQL(tableName, idx)
			if _, err := tx.Exec(indexSql); err != nil {
				return err
			}
		}
	} else {
		currentCols, err := s.dialect.GetColumns(tx, tableName)
		if err != nil {
			return err
		}

		table := buildTable(s.dialect, structName, structType)

		for _, col := range table.Columns {
			if _, exists := currentCols[col.Name]; !exists {
				sql := s.dialect.AddColumnSQL(tableName, col.Name, goorm.ColumnInfo{
					Name:  col.Name,
					Type:  col.Type,
					Extra: col.Options,
				})

				if _, err := tx.Exec(sql); err != nil {
					return err
				}
			}
		}

		currentFKs, err := s.dialect.GetForeignKeys(tx, tableName)
		if err != nil {
			return err
		}

		for _, fk := range table.ForeignKeys {
			if _, exists := currentFKs[fk.Name]; !exists {
				sql := s.dialect.AddForeignKeySQL(tableName, fk)
				if _, err := tx.Exec(sql); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
