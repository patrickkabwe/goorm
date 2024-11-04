package goorm

import (
	"database/sql"
	"fmt"
	"strings"
)

// SQLite dialect implementation
type SQLite struct{}

func (s *SQLite) GetName() Driver {
	return "sqlite3"
}

func (s *SQLite) GetPlaceholder(index int) string {
	return "?"
}

func (s *SQLite) Quote(identifier string) string {
	return `"` + identifier + `"`
}

func (m *SQLite) TableExists(tx *sql.Tx, tableName string) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_schema = DATABASE() 
            AND table_name = ?
        )
    `
	err := tx.QueryRow(query, tableName).Scan(&exists)
	return exists, err
}

func (m *SQLite) GetColumns(tx *sql.Tx, tableName string) (map[string]ColumnInfo, error) {
	columns := make(map[string]ColumnInfo)
	query := `
        SELECT 
            column_name, 
            column_type,
            is_nullable,
            column_default,
            extra
        FROM information_schema.columns
        WHERE table_schema = DATABASE()
        AND table_name = ?
    `

	rows, err := tx.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var isNullable string
		var defaultValue, extra sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &isNullable, &defaultValue, &extra); err != nil {
			return nil, err
		}

		col.IsNullable = isNullable == "YES"
		if defaultValue.Valid {
			col.Default = defaultValue.String
		}
		if extra.Valid {
			col.Extra = extra.String
		}

		columns[col.Name] = col
	}

	return columns, nil
}

func (m *SQLite) GetForeignKeys(tx *sql.Tx, tableName string) (map[string]ForeignKey, error) {
	fks := make(map[string]ForeignKey)
	query := `
        SELECT 
            constraint_name,
            column_name,
            referenced_table_name,
            referenced_column_name
        FROM information_schema.key_column_usage
        WHERE table_schema = DATABASE()
        AND table_name = ?
        AND referenced_table_name IS NOT NULL
    `

	rows, err := tx.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fk ForeignKey
		err := rows.Scan(&fk.Name, &fk.Column, &fk.RefTable, &fk.RefColumn)
		if err != nil {
			return nil, err
		}
		fks[fk.Name] = fk
	}

	return fks, nil
}

func (m *SQLite) CreateTableSQL(table Table) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", m.Quote(table.Name)))
	b.WriteString("  id varchar(36) NOT NULL,\n")
	b.WriteString("  created_at timestamp DEFAULT CURRENT_TIMESTAMP,\n")
	b.WriteString("  updated_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP")

	for _, col := range table.Columns {
		b.WriteString(",\n  ")
		b.WriteString(m.Quote(col.Name))
		b.WriteString(" ")
		b.WriteString(col.Type)
		if col.Options != "" {
			b.WriteString(" ")
			b.WriteString(col.Options)
		}
	}

	for _, idx := range table.Indexes {
		b.WriteString(",\n  KEY ")
		b.WriteString(idx.Name)
		b.WriteString(" (")
		b.WriteString(idx.Columns)
		b.WriteString(")")
	}

	for _, fk := range table.ForeignKeys {
		b.WriteString(",\n  CONSTRAINT ")
		b.WriteString(fk.Name)
		b.WriteString(" FOREIGN KEY (")
		b.WriteString(fk.Column)
		b.WriteString(") REFERENCES ")
		b.WriteString(fk.RefTable)
		b.WriteString(" (")
		b.WriteString(fk.RefColumn)
		b.WriteString(")")
		if fk.Options != "" {
			b.WriteString(" ")
			b.WriteString(fk.Options)
		}
	}

	b.WriteString("\n)")
	return b.String()
}

func (m *SQLite) CreateIndexSQL(table string, index Index) string {
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
		m.Quote(index.Name),
		m.Quote(table),
		index.Columns,
	)
}

func (m *SQLite) AddColumnSQL(table, column string, info ColumnInfo) string {
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s%s",
		m.Quote(table),
		m.Quote(column),
		info.Type,
		formatOptions(info),
	)
}

func (m *SQLite) ModifyColumnSQL(table, column string, info ColumnInfo) string {
	return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s%s",
		m.Quote(table),
		m.Quote(column),
		info.Type,
		formatOptions(info),
	)
}

func (m *SQLite) AddForeignKeySQL(table string, fk ForeignKey) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)%s",
		m.Quote(table),
		fk.Name,
		m.Quote(fk.Column),
		m.Quote(fk.RefTable),
		m.Quote(fk.RefColumn),
		formatOptions(fk.Options),
	)
}

func (m *SQLite) DropForeignKeySQL(table string, fk ForeignKey) string {
	return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s",
		m.Quote(table),
		fk.Name,
	)
}

func (m *SQLite) SQLType(goType string) string {
	switch goType {
	case "string":
		return "varchar(255)"
	case "int", "int32":
		return "int"
	case "int64":
		return "bigint"
	case "bool":
		return "tinyint(1)"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "time.Time":
		return "timestamp"
	default:
		return "text"
	}
}
