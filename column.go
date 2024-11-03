package goorm

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

// SQLType maps Go types to SQL types
var SQLType = map[string]string{
	"string":  "varchar(255)",
	"int":     "int",
	"int64":   "bigint",
	"bool":    "boolean",
	"text":    "text",
	"*string": "varchar(255)",
	"*int":    "int",
	"*int64":  "bigint",
	"*bool":   "boolean",
}

func GetColumnOptions(field *ast.Field, driver Driver) map[string]Column {
	tags := make(map[string]Column)

	if field.Tag == nil {
		return tags
	}

	tagValue := strings.Trim(field.Tag.Value, "`")
	var fieldName string
	if len(field.Names) > 0 {
		fieldName = field.Names[0].Name
	}

	structTag := reflect.StructTag(tagValue)

	if dbTag, ok := structTag.Lookup(DB_OPT_LOOKUP); ok {
		sqlType := getColumnType(dbTag, driver)
		options := parseColumnConstraints(dbTag, driver)

		tags[fieldName] = Column{
			Name:    fieldName,
			Type:    sqlType,
			Options: options,
		}
	} else {
		tags[fieldName] = Column{
			Name:    fieldName,
			Type:    getSQLType(field.Type),
			Options: "",
		}
	}

	return tags
}

func getSQLType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:

		if sqlType, ok := SQLType[t.Name]; ok {
			return sqlType
		}
		if t.Name == "string" {
			return "varchar(255)"
		}
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			if sqlType, ok := SQLType["*"+ident.Name]; ok {
				return sqlType
			}
		}
	}
	return "text"
}

func parseColumnConstraints(tag string, dialect Driver) string {
	var constraints []string

	parts := strings.Split(tag, ",")
	hasAutoIncrement := false
	hasPrimaryKey := false

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Skip empty parts
		if part == "" {
			continue
		}

		if strings.Contains(part, ":") {
			kv := strings.SplitN(part, ":", 2)
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			switch key {
			case "default":
				constraints = append(constraints, fmt.Sprintf("DEFAULT %s", value))
			case "check":
				checkCond := strings.Trim(value, "()")
				constraints = append(constraints, fmt.Sprintf("CHECK (%s)", checkCond))
			case "length":
				// Handle length in the type conversion, not here
				continue
			}
			continue
		}

		switch part {
		case "primary key":
			hasPrimaryKey = true
			constraints = append(constraints, "PRIMARY KEY")
		case "auto_increment":
			hasAutoIncrement = true
		case "not null":
			constraints = append(constraints, "NOT NULL")
		}

		fmt.Println(hasAutoIncrement, hasPrimaryKey)
	}

	return strings.Join(constraints, " ")
}

func getColumnType(tag string, dialect Driver) string {
	var length string
	parts := strings.Split(tag, ",")

	// Extract length if specified
	for _, part := range parts {
		if strings.HasPrefix(part, "length:") {
			length = strings.TrimPrefix(part, "length:")
		}
	}

	// Check for explicit type
	for _, part := range parts {
		if strings.HasPrefix(part, "type:") {
			declaredType := strings.TrimPrefix(part, "type:")
			if length != "" {
				return fmt.Sprintf("%s(%s)", convertType(declaredType, dialect), length)
			}
			return convertType(declaredType, dialect)
		}
	}

	// Default type handling
	if length != "" {
		return fmt.Sprintf("varchar(%s)", length)
	}
	return "varchar(255)" // default
}

func convertType(declaredType string, dialect Driver) string {
	switch strings.ToLower(declaredType) {
	case "serial":
		return "SERIAL"
	case "int":
		return "INTEGER"
	case "varchar":
		return "VARCHAR"
	case "text":
		return "TEXT"
	case "boolean":
		return "BOOLEAN"
	case "timestamp":
		return "TIMESTAMP"
	default:
		return strings.ToUpper(declaredType)
	}
}
