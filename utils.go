package goorm

import (
	"os"
	"strings"
)

const ()

func GetENV(key string) string {
	return os.Getenv(key)
}

func splitFields(fields string) []string {
	var result []string
	var current strings.Builder
	inFunction := 0
	inQuote := false
	quoteChar := rune(0)

	for _, char := range fields {
		switch char {
		case '(':
			if !inQuote {
				inFunction++
			}
			current.WriteRune(char)
		case ')':
			if !inQuote {
				inFunction--
			}
			current.WriteRune(char)
		case '"', '\'':
			if inQuote && char == quoteChar {
				inQuote = false
				quoteChar = 0
			} else if !inQuote {
				inQuote = true
				quoteChar = char
			}
			current.WriteRune(char)
		case ',':
			if !inQuote && inFunction == 0 {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

func hasOperation(operations []string, operation string) bool {
	for _, op := range operations {
		if op == operation {
			return true
		}
	}
	return false
}

func supportsReturning(dialect Dialect) bool {
	switch dialect.GetName() {
	case Postgres:
		return true
	default:
		return false
	}
}

func isColumnNameInWhere(parts []string, pos int) bool {
	if pos >= len(parts)-1 {
		return false
	}

	nextPart := strings.ToUpper(parts[pos+1])

	// Common patterns where the current part might be a column
	switch nextPart {
	case "=", ">", "<", ">=", "<=", "!=", "<>", "IN", "IS", "LIKE", "BETWEEN":
		return true
	}

	if pos > 0 {
		prevPart := strings.ToUpper(parts[pos-1])
		switch prevPart {
		case "AND", "OR", "WHERE", "NOT":
			return true
		}
	}

	return false
}
