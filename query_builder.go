package goorm

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type QueryBuilder struct {
	query        strings.Builder
	db           *sql.DB
	dialect      Dialect
	logger       Logger
	params       []interface{}
	currentTable string
	operations   []string
	returning    []string
}

func NewQueryBuilder(db *sql.DB, dialect Dialect, logger Logger) *QueryBuilder {
	if logger == nil {
		logger = NewDefaultLogger()
	}
	return &QueryBuilder{
		db:      db,
		dialect: dialect,
		params:  make([]interface{}, 0),
		logger:  logger,
	}
}

func (q *QueryBuilder) Close() error {
	if q.db != nil {
		return q.db.Close()
	}
	return nil
}

func (q *QueryBuilder) Select(fields ...string) *QueryBuilder {
	if len(fields) == 0 {
		q.query.WriteString("SELECT *")
		return q
	}

	q.query.WriteString("SELECT ")
	q.operations = append(q.operations, "SELECT")

	formattedFields := make([]string, len(fields))
	for i, field := range fields {
		field = strings.TrimSpace(field) // Bug Fix 3: Trim spaces
		if strings.Contains(field, "(") ||
			field == "*" ||
			strings.Contains(field, ".") ||
			strings.Contains(field, " as ") ||
			strings.Contains(field, " AS ") {
			formattedFields[i] = field
		} else {
			formattedFields[i] = field
		}
	}

	q.query.WriteString(strings.Join(formattedFields, ", "))
	return q
}

func (q *QueryBuilder) From(table string) *QueryBuilder {
	table = strings.TrimSpace(table)
	if table == "" {
		return q
	}

	// Extract table name and alias
	parts := strings.Fields(table)
	q.currentTable = parts[0]

	// Handle schema.table format
	if strings.Contains(q.currentTable, ".") {
		tableParts := strings.Split(q.currentTable, ".")
		q.currentTable = tableParts[len(tableParts)-1]
	}

	q.operations = append(q.operations, "FROM")

	// Prefix columns in SELECT clause if it exists
	if hasOperation(q.operations, "SELECT") {
		query := q.query.String()
		parts := strings.SplitN(query, "FROM", 2)
		if len(parts) > 0 {
			selectPart := parts[0]
			selectPart = strings.TrimPrefix(selectPart, "SELECT ")
			fields := splitFields(selectPart)

			formattedFields := make([]string, len(fields))
			for i, field := range fields {
				field = strings.TrimSpace(field)
				if strings.Contains(field, "(") ||
					field == "*" ||
					strings.Contains(field, ".") ||
					strings.Contains(field, " as ") ||
					strings.Contains(field, " AS ") {
					formattedFields[i] = field
				} else {
					formattedFields[i] = fmt.Sprintf("%s.%s", q.currentTable, field)
				}
			}

			q.query.Reset()
			q.query.WriteString("SELECT ")
			q.query.WriteString(strings.Join(formattedFields, ", "))
		}
	}

	q.query.WriteString(" FROM ")
	q.query.WriteString(table)

	return q
}

func (q *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	if condition == "" {
		return q
	}

	q.params = append(q.params, args...)
	q.operations = append(q.operations, "WHERE")

	if q.currentTable != "" {
		condition = q.prefixColumns(condition)
	}

	q.query.WriteString(" WHERE ")
	q.query.WriteString(condition)

	return q
}

func (q *QueryBuilder) And(condition string, args ...any) *QueryBuilder {
	return q.handleClause("AND", condition, args...)
}

func (q *QueryBuilder) Or(condition string, args ...any) *QueryBuilder {
	return q.handleClause("OR", condition, args...)
}

func (q *QueryBuilder) Not(condition string, args ...any) *QueryBuilder {
	return q.handleClause("NOT", condition, args...)
}

func (q *QueryBuilder) Like(column string, value string) *QueryBuilder {
	q.query.WriteString(" LIKE " + column)
	q.query.WriteString(" '")
	q.query.WriteString(value)
	q.query.WriteString("'")
	return q
}

func (q *QueryBuilder) NotLike(column string, value string) *QueryBuilder {
	q.query.WriteString(" NOT LIKE " + column)
	q.query.WriteString(" '")
	q.query.WriteString(value)
	q.query.WriteString("'")
	return q
}

func (q *QueryBuilder) In(column string, values ...any) *QueryBuilder {
	q.query.WriteString(" IN (")
	for i, value := range values {
		if i > 0 {
			q.query.WriteString(", ")
		}
		q.query.WriteString(fmt.Sprint(value))
	}
	q.query.WriteString(")")
	return q
}

func (q *QueryBuilder) NotIn(column string, values ...any) *QueryBuilder {
	q.query.WriteString(" NOT IN (")
	for i, value := range values {
		if i > 0 {
			q.query.WriteString(", ")
		}
		q.query.WriteString(fmt.Sprint(value))
	}
	q.query.WriteString(")")
	return q
}

func (q *QueryBuilder) IsNull(column string) *QueryBuilder {
	q.query.WriteString(" IS NULL")
	return q
}

func (q *QueryBuilder) IsNotNull(column string) *QueryBuilder {
	q.query.WriteString(" IS NOT NULL")
	return q
}

func (q *QueryBuilder) Between(column string, values ...any) *QueryBuilder {
	q.query.WriteString(" BETWEEN ")
	for i, value := range values {
		if i > 0 {
			q.query.WriteString(" AND ")
		}
		q.query.WriteString(fmt.Sprint(value))
	}
	return q
}

func (q *QueryBuilder) NotBetween(column string, values ...any) *QueryBuilder {
	q.query.WriteString(" NOT BETWEEN ")
	for i, value := range values {
		if i > 0 {
			q.query.WriteString(" AND ")
		}
		q.query.WriteString(fmt.Sprint(value))
	}
	return q
}

func (q *QueryBuilder) InsertInto(table string) *QueryBuilder {
	q.query.WriteString(" INSERT INTO " + table)
	return q
}

func (q *QueryBuilder) Values(values ...any) *QueryBuilder {
	q.query.WriteString(" VALUES ")
	for i, value := range values {
		if i > 0 {
			q.query.WriteString(", ")
		}
		q.query.WriteString(
			"(" +
				strings.Join(strings.Split(fmt.Sprint(value), " "), ",") +
				")",
		)
	}
	return q
}

func (q *QueryBuilder) Update(table string) *QueryBuilder {
	q.query.WriteString(" UPDATE " + table)
	return q
}

func (q *QueryBuilder) Set(field string, value any) *QueryBuilder {
	q.query.WriteString(" SET " + field + " = " + fmt.Sprint(value))
	return q
}

func (q *QueryBuilder) Delete(table string) *QueryBuilder {
	q.query.WriteString(" DELETE FROM " + table)
	return q
}

func (q *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	q.query.WriteString(" GROUP BY ")
	for i, field := range fields {
		if i > 0 {
			q.query.WriteString(", ")
		}
		q.query.WriteString(field)
	}
	return q
}

func (q *QueryBuilder) Having(condition string) *QueryBuilder {
	q.query.WriteString(" HAVING " + condition)
	return q
}

func (q *QueryBuilder) OrderBy(fields ...string) *QueryBuilder {
	q.query.WriteString(" ORDER BY ")
	for i, field := range fields {
		if i > 0 {
			q.query.WriteString(", ")
		}
		q.query.WriteString(field)
	}
	return q
}

func (q *QueryBuilder) Limit(limit int) *QueryBuilder {
	q.query.WriteString(" LIMIT " + strconv.Itoa(limit))
	return q
}

func (q *QueryBuilder) Offset(offset int) *QueryBuilder {
	q.query.WriteString(" OFFSET " + strconv.Itoa(offset))
	return q
}

func (q *QueryBuilder) Distinct() *QueryBuilder {
	q.query.WriteString(" DISTINCT")
	return q
}

func (q *QueryBuilder) SubQuery(query string) *QueryBuilder {
	q.query.WriteString("(" + query + ")")
	return q
}

func (q *QueryBuilder) Exists(query string) *QueryBuilder {
	q.query.WriteString(" EXISTS (" + query + ")")
	return q
}

func (q *QueryBuilder) NotExists(query string) *QueryBuilder {
	q.query.WriteString(" NOT EXISTS (" + query + ")")
	return q
}

func (q *QueryBuilder) LeftJoin(table string, condition string) *QueryBuilder {
	q.query.WriteString(" LEFT JOIN " + table + " ON " + condition)
	return q
}

func (q *QueryBuilder) RightJoin(table string, condition string) *QueryBuilder {
	q.query.WriteString(" RIGHT JOIN " + table + " ON " + condition)
	return q
}

func (q *QueryBuilder) InnerJoin(table string, condition string) *QueryBuilder {
	q.query.WriteString(" INNER JOIN " + table + " ON " + condition)
	return q
}

func (q *QueryBuilder) FullJoin(table string, condition string) *QueryBuilder {
	q.query.WriteString(" FULL JOIN " + table + " ON " + condition)
	return q
}

func (q *QueryBuilder) CrossJoin(table string) *QueryBuilder {
	q.query.WriteString(" CROSS JOIN " + table)
	return q
}

func (q *QueryBuilder) CaseWhen(when string, then string) *QueryBuilder {
	q.query.WriteString(" WHEN " + when + " THEN " + then)
	return q
}

func (q *QueryBuilder) CaseElse(e string) *QueryBuilder {
	q.query.WriteString(" ELSE " + e)
	return q
}

func (q *QueryBuilder) CaseEnd() *QueryBuilder {
	q.query.WriteString(" END")
	return q
}

func (q *QueryBuilder) GetSql() string {
	return q.FormatQuery()
}

func (q *QueryBuilder) Returning(fields ...string) *QueryBuilder {
	if len(fields) > 0 {
		q.returning = append(q.returning, fields...)
	}
	return q
}

// Modify FormatQuery to include RETURNING clause
func (q *QueryBuilder) FormatQuery() string {
	query := q.query.String()

	// Add RETURNING clause if present and supported
	if len(q.returning) > 0 && q.dialect != nil && supportsReturning(q.dialect) {
		returningFields := make([]string, len(q.returning))
		for i, field := range q.returning {
			if !strings.Contains(field, ".") && q.currentTable != "" {
				returningFields[i] = fmt.Sprintf("%s.%s", q.currentTable, field)
			} else {
				returningFields[i] = field
			}
		}
		query += " RETURNING " + strings.Join(returningFields, ", ")
	}

	// Add semicolon if not present
	if !strings.HasSuffix(query, ";") {
		query += ";"
	}

	q.logger.Info(query, "args", q.params)

	return query
}

func GetQuery(dialect Dialect, model any, params P) (string, []interface{}) {
	q := NewQueryBuilder(nil, dialect, nil)
	return q.GetSql(), q.params
}

// Modify Execute to handle RETURNING clause
func (q *QueryBuilder) Execute(ctx context.Context) (*sql.Rows, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := q.GetSql()

	if len(q.returning) > 0 {
		if q.dialect != nil && supportsReturning(q.dialect) {
			return q.db.QueryContext(ctx, query, q.params...)
		}
		return q.handleReturningFallback(ctx)
	}

	dbQuery, err := q.db.QueryContext(ctx, query, q.params...)

	if err != nil {
		return nil, err
	}

	return dbQuery, nil
}
func (q *QueryBuilder) Reset() {
	q.query.Reset()
	q.operations = make([]string, 0)
	q.returning = make([]string, 0)
	q.params = q.params[:0]
	q.currentTable = ""
}

// Helper function to determine if a token could be a column name in WHERE clause
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

// Modify handleClause to use the same logic
func (q *QueryBuilder) handleClause(clause string, condition string, args ...interface{}) *QueryBuilder {
	if condition == "" {
		return q
	}

	if q.currentTable != "" {
		condition = q.prefixColumns(condition)
	}

	q.query.WriteString(fmt.Sprintf(" %s ", clause))
	q.query.WriteString(condition)
	q.params = append(q.params, args...)

	return q
}

func (q *QueryBuilder) handleReturningFallback(ctx context.Context) (*sql.Rows, error) {
	// Start a transaction
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute the original query without RETURNING
	originalQuery := q.query.String()
	if !strings.HasSuffix(originalQuery, ";") {
		originalQuery += ";"
	}

	result, err := tx.ExecContext(ctx, originalQuery, q.params...)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return nil, fmt.Errorf("failed to rollback transaction: %w", err)
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// For INSERT queries, get the last inserted ID
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(originalQuery)), "INSERT") {
		lastID, err := result.LastInsertId()
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return nil, fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return nil, fmt.Errorf("failed to get last insert ID: %w", err)
		}

		// Build SELECT query to fetch the returned fields
		var selectQuery strings.Builder
		selectQuery.WriteString("SELECT ")
		selectQuery.WriteString(strings.Join(q.returning, ", "))
		selectQuery.WriteString(" FROM ")
		selectQuery.WriteString(q.currentTable)
		selectQuery.WriteString(" WHERE id = ?") // Assuming 'id' is the primary key

		// Execute SELECT query
		rows, err := tx.QueryContext(ctx, selectQuery.String(), lastID)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return nil, fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return nil, fmt.Errorf("failed to fetch returned fields: %w", err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return rows, nil
	}

	// For UPDATE/DELETE queries, we need to fetch the affected rows
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(originalQuery)), "UPDATE") ||
		strings.HasPrefix(strings.ToUpper(strings.TrimSpace(originalQuery)), "DELETE") {

		// Get the WHERE clause from the original query
		whereParts := strings.Split(originalQuery, "WHERE")
		if len(whereParts) != 2 {
			err := tx.Rollback()
			if err != nil {
				return nil, fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return nil, fmt.Errorf("cannot handle RETURNING clause without WHERE condition")
		}

		// Build SELECT query
		var selectQuery strings.Builder
		selectQuery.WriteString("SELECT ")
		selectQuery.WriteString(strings.Join(q.returning, ", "))
		selectQuery.WriteString(" FROM ")
		selectQuery.WriteString(q.currentTable)
		selectQuery.WriteString(" WHERE ")
		selectQuery.WriteString(strings.TrimSuffix(whereParts[1], ";"))

		// Execute SELECT query
		rows, err := tx.QueryContext(ctx, selectQuery.String(), q.params...)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return nil, fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return nil, fmt.Errorf("failed to fetch returned fields: %w", err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return rows, nil
	}

	err = tx.Rollback()
	if err != nil {
		return nil, fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil, fmt.Errorf("unsupported query type for RETURNING fallback")
}

func (q *QueryBuilder) prefixColumns(condition string) string {
	parts := strings.Fields(condition)
	for i, part := range parts {
		if isColumnNameInWhere(parts, i) &&
			!strings.Contains(part, ".") &&
			!strings.HasPrefix(part, "$") &&
			!strings.HasPrefix(part, "'") &&
			!strings.HasPrefix(part, "\"") &&
			!strings.Contains(part, "(") {
			parts[i] = fmt.Sprintf("%s.%s", q.currentTable, part)
		}
	}
	return strings.Join(parts, " ")
}
