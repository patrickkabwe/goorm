package main

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/patrickkabwe/goorm"
)

func buildTable(dialect goorm.Dialect, structName string, structType *ast.StructType) goorm.Table {
	tableName := toSnakeCase(structName) + "s"
	table := goorm.Table{
		Name: tableName,
	}

	seenFields := make(map[string]bool)

	indexes := make(map[string]goorm.Index)

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		var fieldName = field.Names[0].Name
		var dbTag string

		// Get db tag if exists
		if field.Tag != nil {
			tag := strings.Trim(field.Tag.Value, "`")
			for _, tagPart := range strings.Split(tag, " ") {
				if strings.HasPrefix(tagPart, DB_COL) {
					dbTag = strings.Trim(strings.Split(tagPart, ":")[1], "\"")
					break
				}
			}
		}

		columnOptions := goorm.GetColumnOptions(field, dialect)

		// Skip if no db tag or already seen
		if dbTag == "" || seenFields[dbTag] {
			continue
		}
		seenFields[dbTag] = true

		var hasIndexTag = strings.Contains(dbTag, "unique") || strings.Contains(dbTag, "index")

		if hasIndexTag {
			indexes = map[string]goorm.Index{
				fmt.Sprintf("idx_%s", dbTag): {
					Name:    fmt.Sprintf("idx_%s", dbTag),
					Columns: dbTag,
					Unique:  strings.Contains(dbTag, "unique"),
				},
			}
		}

		// Check if it's a foreign key field (ends with _id)
		// Add foreign key column
		if strings.HasSuffix(dbTag, "_id") {
			table.Columns = append(table.Columns, goorm.Column{
				Name:    dbTag,
				Type:    columnOptions[fieldName].Type,
				Options: columnOptions[fieldName].Options,
			})

			// Add index
			idxName := fmt.Sprintf("idx_%s", dbTag)
			indexes[idxName] = goorm.Index{
				Name:    idxName,
				Columns: dbTag,
				Unique:  strings.Contains(dbTag, "unique"),
			}
			table.Indexes = append(table.Indexes, indexes[idxName])

			// Add foreign key constraint
			refTableName := strings.TrimSuffix(dbTag, "_id")

			fk := goorm.ForeignKey{
				Name:      fmt.Sprintf("fk_%s_%s", tableName, refTableName),
				Column:    dbTag,
				RefTable:  refTableName + "s",
				RefColumn: "id",
				Options:   "ON DELETE CASCADE",
			}
			table.ForeignKeys = append(table.ForeignKeys, fk)
			continue
		}

		table.Columns = append(table.Columns, goorm.Column{
			Name:    dbTag,
			Type:    columnOptions[fieldName].Type,
			Options: columnOptions[fieldName].Options,
		})
	}

	if len(table.Indexes) > 0 {
		table.Indexes[len(table.Indexes)-1].Last = true
	}
	if len(table.ForeignKeys) > 0 {
		table.ForeignKeys[len(table.ForeignKeys)-1].Last = true
	}

	return table
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
