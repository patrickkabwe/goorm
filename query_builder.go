package goorm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type QueryBuilder struct {
	query        strings.Builder
	db           *sql.DB
	logger       Logger
	Dialect      Dialect
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
		params:  make([]interface{}, 0),
		logger:  logger,
		Dialect: dialect,
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

// Distinct adds the DISTINCT keyword to the query
// e.g. SELECT DISTINCT * FROM users
func (q *QueryBuilder) SelectDistinct(fields ...string) *QueryBuilder {
	if len(fields) == 0 {
		q.query.WriteString("SELECT DISTINCT *")
		return q
	}

	q.query.WriteString("SELECT DISTINCT ")
	q.operations = append(q.operations, "SELECT DISTINCT")

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
	if hasOperation(q.operations, "SELECT") || hasOperation(q.operations, "SELECT DISTINCT") {
		query := q.query.String()
		parts := strings.SplitN(query, "FROM", 2)
		if len(parts) > 0 {
			selectPart := parts[0]
			selectPart = strings.TrimPrefix(selectPart, "SELECT ")
			selectPart = strings.TrimPrefix(selectPart, "SELECT DISTINCT ")
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
	q.query.WriteString("INSERT INTO " + table)
	return q
}

func (q *QueryBuilder) Columns(values ...any) *QueryBuilder {
	q.query.WriteString("(")
	for i, value := range values {
		if i > 0 {
			q.query.WriteString(", ")
		}
		q.query.WriteString(strings.Join(strings.Split(fmt.Sprint(value), " "), ","))
	}
	q.query.WriteString(")")
	return q
}

func (q *QueryBuilder) Values(values ...any) *QueryBuilder {
	q.query.WriteString(" VALUES (")
	for i, value := range values {
		if i > 0 {
			q.query.WriteString(", ")
		}
		q.query.WriteString(q.Dialect.GetPlaceholder(i + 1))
		q.params = append(q.params, value)
	}
	q.query.WriteString(")")
	return q
}

func (q *QueryBuilder) Update(table string) *QueryBuilder {
	q.query.WriteString("UPDATE " + table)
	return q
}

func (q *QueryBuilder) Set(values map[string]string) *QueryBuilder {
	q.query.WriteString(" SET ")
	cnt := 0
	for field, value := range values {
		cnt++
		q.query.WriteString(fmt.Sprintf("%s=%s", field, q.Dialect.GetPlaceholder(cnt)))
		if len(values) > cnt {
			q.query.WriteString(",")
		}
		q.params = append(q.params, value)
	}
	return q
}

func (q *QueryBuilder) Delete(table string) *QueryBuilder {
	q.query.WriteString("DELETE FROM " + table)
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

// SubQuery adds a sub-query to the query
// e.g. SELECT * FROM (SELECT * FROM users WHERE id = 1) AS users
// e.g. SELECT * FROM (SELECT * FROM users WHERE id = 1) AS users WHERE id = 2
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
	return q.Join("LEFT", table, condition)
}

func (q *QueryBuilder) RightJoin(table string, condition string) *QueryBuilder {
	return q.Join("RIGHT", table, condition)
}

func (q *QueryBuilder) InnerJoin(table string, condition string) *QueryBuilder {
	return q.Join("INNER", table, condition)
}

// Join adds a JOIN clause to the query
func (q *QueryBuilder) Join(joinType string, table string, condition string) *QueryBuilder {
	q.query.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", joinType, table, condition))
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

func (q *QueryBuilder) Returning(ctx context.Context, model interface{}, fields ...string) error {
	if len(fields) > 0 {
		q.returning = append(q.returning, fields...)
	}

	return q.exec(ctx, model)
}

func (q *QueryBuilder) DropColumn(table string) *QueryBuilder {
	q.query.WriteString("DROP COLUMN " + table)
	return q
}

func (q *QueryBuilder) DropTable(table string) *QueryBuilder {
	q.query.WriteString("DROP TABLE " + table)
	return q
}

func (q *QueryBuilder) Truncate(table string) *QueryBuilder {
	q.query.WriteString("TRUNCATE TABLE " + table)
	return q
}

func (q *QueryBuilder) Cascade() *QueryBuilder {
	q.query.WriteString(" CASCADE")
	return q
}

func (q *QueryBuilder) RestartIdentity() *QueryBuilder {
	q.query.WriteString(" RESTART IDENTITY")
	return q
}

// GetSql returns the query string
func (q *QueryBuilder) GetSql() string {
	return q.formatQuery()
}

func (q *QueryBuilder) formatQuery() string {
	query := q.query.String()

	if len(q.returning) > 0 && q.Dialect != nil && supportsReturning(q.Dialect) {
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

	if !strings.HasSuffix(query, ";") {
		query += ";"
	}

	q.logger.Info(query, "args", q.params)

	return query
}

// Exec executes the query
func (q *QueryBuilder) Exec(ctx context.Context) (*sql.Rows, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := q.GetSql()
	defer q.Reset()

	var rows *sql.Rows
	var err error
	if len(q.returning) > 0 {
		if q.Dialect != nil && supportsReturning(q.Dialect) {
			rows, err = q.db.QueryContext(ctx, query, q.params...)
		} else {
			// Fallback for databases that don't support RETURNING
			// This might involve doing the insert/update first
			// then running a separate SELECT query
			rows, err = q.handleReturningFallback(ctx)
		}
	} else {
		rows, err = q.db.QueryContext(ctx, query, q.params...)
	}

	if err != nil {
		q.logger.Error(err.Error())
		return nil, err
	}

	return rows, nil
}

// Scan maps struct fields to db field
func (q *QueryBuilder) Scan(ctx context.Context, model interface{}) error {
	rows, err := q.Exec(ctx)
	if err != nil {
		return err
	}
	return q.mapToModel(rows, model)
}

func (q *QueryBuilder) exec(ctx context.Context, model interface{}) error {
	rows, err := q.Exec(ctx)
	if err != nil {
		return err
	}
	return q.mapToModel(rows, model)
}

func (q *QueryBuilder) Reset() {
	q.query.Reset()
	q.operations = make([]string, 0)
	q.returning = make([]string, 0)
	q.params = make([]interface{}, 0)
	q.currentTable = ""
}

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

func (q *QueryBuilder) mapToModel(rows *sql.Rows, model interface{}) error {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Pointer {
		return fmt.Errorf("model must be a pointer")
	}
	modelValue = modelValue.Elem()

	isSlice := modelValue.Kind() == reflect.Slice
	var elemType reflect.Type
	if isSlice {
		elemType = modelValue.Type().Elem()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
	} else {
		elemType = modelValue.Type()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	selectedColumns := make(map[string]bool)
	for _, col := range columns {
		selectedColumns[col] = true
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	var results reflect.Value
	if isSlice {
		results = reflect.MakeSlice(modelValue.Type(), 0, 0)
	}

	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return err
		}

		valueMap := make(map[string]interface{})
		for i, col := range columns {
			value := *(values[i].(*interface{}))
			valueMap[col] = value
		}

		newStruct := reflect.New(elemType).Elem()

		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			fieldValue := newStruct.Field(i)

			if !fieldValue.CanSet() {
				continue
			}

			dbTag := field.Tag.Get("db")
			if dbTag == "" {
				continue
			}

			fieldType := field.Type
			isPtr := fieldType.Kind() == reflect.Pointer
			if isPtr {
				fieldType = fieldType.Elem()
			}

			if fieldType.Kind() != reflect.Struct {
				if value, exists := valueMap[dbTag]; exists {
					if err := setFieldValue(fieldValue, value); err != nil {
						return err
					}
				}

			} else {
				hasSelectedFields := false
				for j := 0; j < fieldType.NumField(); j++ {
					nestedTag := fieldType.Field(j).Tag.Get("db")
					if selectedColumns[nestedTag] {

						hasSelectedFields = true
					}
				}

				if hasSelectedFields {
					nestedStruct := reflect.New(fieldType)

					for j := 0; j < fieldType.NumField(); j++ {
						nestedField := fieldType.Field(j)
						nestedFieldValue := nestedStruct.Elem().Field(j)

						if !nestedFieldValue.CanSet() {
							continue
						}

						nestedTag := nestedField.Tag.Get("db")
						if nestedTag == "" {
							continue
						}

						possibleNames := []string{
							dbTag + "_" + nestedTag,
							strings.TrimSuffix(dbTag, "s") + "_" + nestedTag,
						}

						if nestedTag != "id" {
							possibleNames = append(possibleNames, nestedTag)
						}

						for _, name := range possibleNames {
							if value, exists := valueMap[name]; exists {
								if err := setFieldValue(nestedFieldValue, value); err != nil {
									return err
								}
								break
							}
						}
					}

					if isPtr {
						fieldValue.Set(nestedStruct)
					} else {
						fieldValue.Set(nestedStruct.Elem())
					}
				} else if isPtr {
					fieldValue.Set(reflect.Zero(field.Type))
				}

			}
		}

		if isSlice {
			if modelValue.Type().Elem().Kind() == reflect.Ptr {
				ptrValue := reflect.New(elemType)
				ptrValue.Elem().Set(newStruct)
				results = reflect.Append(results, ptrValue)
			} else {
				results = reflect.Append(results, newStruct)
			}
		} else {
			modelValue.Set(newStruct)
			break
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if isSlice {
		modelValue.Set(results)
	}

	return nil
}

func setFieldValue(field reflect.Value, value interface{}) error {
	if !field.CanSet() || value == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.Int64:
		switch v := value.(type) {
		case int64:
			field.SetInt(v)
		case int32:
			field.SetInt(int64(v))
		case int:
			field.SetInt(int64(v))
		case []uint8:
			intVal, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intVal)
		}
	case reflect.String:
		switch v := value.(type) {
		case string:
			field.SetString(v)
		case []uint8:
			field.SetString(string(v))
		}
	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			field.SetBool(v)
		case []uint8:
			boolVal, err := strconv.ParseBool(string(v))
			if err != nil {
				return err
			}
			field.SetBool(boolVal)
		}
	case reflect.Float64:
		switch v := value.(type) {
		case float64:
			field.SetFloat(v)
		case []uint8:
			floatVal, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				return err
			}
			field.SetFloat(floatVal)
		}
	}
	return nil
}
