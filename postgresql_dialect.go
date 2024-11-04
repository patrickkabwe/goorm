package goorm

import (
	"database/sql"
	"fmt"
	"strings"
)

// PostgreSQL dialect implementation
type PostgreSQL struct{}

func (p *PostgreSQL) GetName() Driver {
	return "pgx"
}

func (p *PostgreSQL) GetPlaceholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (p *PostgreSQL) Quote(identifier string) string {
	return "\"" + identifier + "\""
}

func (m *PostgreSQL) TableExists(tx *sql.Tx, tableName string) (bool, error) {
	var exists bool
	query := `
       	SELECT EXISTS (
			SELECT 1 FROM pg_tables 
			WHERE schemaname = 'public' 
			AND tablename = $1
		);
    `
	err := tx.QueryRow(query, tableName).Scan(&exists)
	return exists, err
}

func (m *PostgreSQL) GetColumns(tx *sql.Tx, tableName string) (map[string]ColumnInfo, error) {
	columns := make(map[string]ColumnInfo)
	query := `
		SELECT 
			column_name,
			data_type as column_type,
			is_nullable,
			column_default,
			CASE 
				WHEN is_identity = 'YES' THEN 'auto_increment'
				ELSE ''
			END as extra
		FROM information_schema.columns 
		WHERE table_name = $1 
		AND table_schema = 'public'
		ORDER BY ordinal_position;
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

func (m *PostgreSQL) GetForeignKeys(tx *sql.Tx, tableName string) (map[string]ForeignKey, error) {
	fks := make(map[string]ForeignKey)
	query := `
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS referenced_table_name,
			ccu.column_name AS referenced_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND tc.table_name = $1;
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

func (m *PostgreSQL) CreateTableSQL(table Table) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", m.Quote(table.Name)))

	for ti, col := range table.Columns {
		b.WriteString("\n  ")
		b.WriteString(m.Quote(col.Name))
		b.WriteString(" ")
		b.WriteString(col.Type)
		if col.Options != "" {
			b.WriteString(" ")
			b.WriteString(col.Options)
		}
		if ti < len(table.Columns)-1 {
			b.WriteString(",")
		}
	}

	for fi, fk := range table.ForeignKeys {
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

		if fi < len(table.ForeignKeys)-1 {
			b.WriteString(",")
		}
	}

	b.WriteString("\n)")

	fmt.Println(b.String())

	return b.String()
}

func (m *PostgreSQL) CreateIndexSQL(table string, index Index) string {
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
		m.Quote(index.Name),
		m.Quote(table),
		index.Columns,
	)
}

func (m *PostgreSQL) AddColumnSQL(table, column string, info ColumnInfo) string {
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s%s",
		m.Quote(table),
		m.Quote(column),
		info.Type,
		formatOptions(info),
	)
}

func (m *PostgreSQL) ModifyColumnSQL(table, column string, info ColumnInfo) string {
	return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s%s",
		m.Quote(table),
		m.Quote(column),
		info.Type,
		formatOptions(info),
	)
}

func (m *PostgreSQL) AddForeignKeySQL(table string, fk ForeignKey) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)%s",
		m.Quote(table),
		fk.Name,
		m.Quote(fk.Column),
		m.Quote(fk.RefTable),
		m.Quote(fk.RefColumn),
		formatOptions(fk.Options),
	)
}

func (m *PostgreSQL) DropForeignKeySQL(table string, fk ForeignKey) string {
	return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s",
		m.Quote(table),
		fk.Name,
	)
}

func (m *PostgreSQL) SQLType(goType string) string {
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
