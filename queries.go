package goorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// Params represents query parameters
type P struct {
	Data    interface{}
	Where   []Condition
	Select  map[string]bool
	Limit   *int
	Offset  *int
	OrderBy []string
	Include map[string]P
}

// Condition represents a where condition
type Condition struct {
	Field string
	Op    string
	Value interface{}
}

type Query[T any] interface {
	FindMany(params P) ([]T, error)
	FindFirst(params P) (*T, error)
	Create(params P) (*T, error)
	Update(params P) error
	Delete(params P) error
}

// Operator constants
const (
	OpEq    = "="
	OpLike  = "LIKE"
	OpGt    = ">"
	OpLt    = "<"
	OpGte   = ">="
	OpLte   = "<="
	OpIn    = "IN"
	OpNotEq = "!="
)

// Where combines multiple conditions with AND
func Where(conditions ...Condition) []Condition {
	return conditions
}

// Condition builders
func Eq(field string, value interface{}) Condition {
	return Condition{Field: field, Op: OpEq, Value: value}
}

func Like(field string, value string) Condition {
	return Condition{Field: field, Op: OpLike, Value: value}
}

func Gt(field string, value interface{}) Condition {
	return Condition{Field: field, Op: OpGt, Value: value}
}

func Lt(field string, value interface{}) Condition {
	return Condition{Field: field, Op: OpLt, Value: value}
}

func Gte(field string, value interface{}) Condition {
	return Condition{Field: field, Op: OpGte, Value: value}
}

func Lte(field string, value interface{}) Condition {
	return Condition{Field: field, Op: OpLte, Value: value}
}

func NotEq(field string, value interface{}) Condition {
	return Condition{Field: field, Op: OpNotEq, Value: value}
}

func In(field string, values ...interface{}) Condition {
	return Condition{Field: field, Op: OpIn, Value: values}
}

// Generic find methods
func (m *BaseModel[T]) FindMany(params P) ([]T, error) {
	query, args := buildQuery(m.engine, new(T), params)

	var rows *sql.Rows
	var err error

	if m.tx != nil {
		rows, err = m.tx.Query(query, args...)
	} else {
		rows, err = m.engine.db.Query(query, args...)
	}
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		var item T
		if err := scanStruct(rows, &item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if len(params.Include) > 0 && len(results) > 0 {
		err = m.loadRelations(&results, params.Include)
		if err != nil {
			return nil, fmt.Errorf("load relations: %w", err)
		}
	}

	return results, nil
}

func (m *BaseModel[T]) FindFirst(params P) (*T, error) {
	limit := 1
	params.Limit = &limit
	results, err := m.FindMany(params)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

func (m *BaseModel[T]) Create(params P) (*T, error) {
	data, ok := params.Data.(T)
	if !ok {
		return nil, fmt.Errorf("invalid data type for create")
	}

	query, args := buildInsertQuery(m.engine, data)

	var result sql.Result
	var err error

	if m.tx != nil {
		result, err = m.tx.Exec(query, args...)
	} else {
		result, err = m.engine.db.Exec(query, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("execute create: %w", err)
	}

	// Set the ID if it's an auto-increment field
	if id, err := result.LastInsertId(); err == nil {
		setValue(data, "ID", id)
	}

	return &data, nil
}

func (m *BaseModel[T]) Update(params P) error {
	data, ok := params.Data.(T)
	if !ok {
		return fmt.Errorf("invalid data type for update")
	}

	if len(params.Where) == 0 {
		return fmt.Errorf("update requires where conditions")
	}

	query, args := buildUpdateQuery(m.engine, data, params.Where)
	m.logger.Debug("Update query", "query", query, "args", args)

	var err error
	if m.tx != nil {
		_, err = m.tx.Exec(query, args...)
	} else {
		_, err = m.engine.db.Exec(query, args...)
	}

	if err != nil {
		return fmt.Errorf("execute update: %w", err)
	}

	return nil
}

func (m *BaseModel[T]) Delete(params P) error {
	if len(params.Where) == 0 {
		return fmt.Errorf("delete requires where conditions")
	}

	query, args := buildDeleteQuery(m.engine, new(T), params)
	m.logger.Debug("Delete query", "query", query, "args", args)

	var err error
	if m.tx != nil {
		_, err = m.tx.Exec(query, args...)
	} else {
		_, err = m.engine.db.Exec(query, args...)
	}

	if err != nil {
		return fmt.Errorf("execute delete: %w", err)
	}

	return nil
}

// Internal query building function
func buildInsertQuery(engine *Engine, data interface{}) (string, []interface{}) {
	v := reflect.ValueOf(data)
	t := v.Type()

	var columns []string
	var placeholders []string
	var args []interface{}

	for i := 1; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if dbTag := field.Tag.Get(DB_COL_LOOKUP); dbTag != "" {
			if strings.Contains(field.Tag.Get(DB_OPT_LOOKUP), "auto_increment") {
				continue
			}
			columns = append(columns, dbTag)
			placeholders = append(placeholders, engine.dialect.GetPlaceholder(i))
			args = append(args, value.Interface())
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		getTableName(data),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	engine.logger.Info("Insert query", "query", query, "args", args)

	return query, args
}

func buildUpdateQuery(engine *Engine, data interface{}, conditions []Condition) (string, []interface{}) {
	v := reflect.ValueOf(data)
	t := v.Type()

	var setStatements []string
	var args []interface{}

	// Build SET clause
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if dbTag := field.Tag.Get(DB_COL_LOOKUP); dbTag != "" && !isZero(value) {
			setStatements = append(setStatements, fmt.Sprintf("%s = %s", dbTag, engine.dialect.GetPlaceholder(i+1)))
			args = append(args, value.Interface())
		}
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s",
		getTableName(data),
		strings.Join(setStatements, ", "),
	)

	// Add WHERE clause
	if len(conditions) > 0 {
		whereStatements := make([]string, 0, len(conditions))
		for _, cond := range conditions {
			whereStatements = append(whereStatements, fmt.Sprintf("%s %s ?", cond.Field, cond.Op))
			args = append(args, cond.Value)
		}
		query += " WHERE " + strings.Join(whereStatements, " AND ")
	}

	return query, args
}

func buildDeleteQuery[T any](engine *Engine, model T, params P) (string, []interface{}) {
	query := fmt.Sprintf("DELETE FROM %s", getTableName(model))
	var args []interface{}

	if len(params.Where) > 0 {
		conditions := make([]string, 0, len(params.Where))
		for ci, cond := range params.Where {
			conditions = append(conditions, fmt.Sprintf("%s %s %s", cond.Field, cond.Op, engine.dialect.GetPlaceholder(ci+1)))
			args = append(args, cond.Value)
		}
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}

func buildQuery[T any](engine *Engine, model T, params P) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}
	var joins []string
	argCount := 1

	query.WriteString("SELECT ")

	// Handle includes by building JOIN clauses and selecting fields
	if len(params.Include) > 0 {
		selectClauses := buildSelectClauses(model, params.Include)
		query.WriteString(selectClauses)
	} else {
		mainTable := getTableName(model)
		query.WriteString(mainTable + ".*")
	}

	query.WriteString(" FROM ")
	query.WriteString(getTableName(model))

	// Build JOIN clauses for includes
	if len(params.Include) > 0 {
		joinClauses, joinArgs := buildJoinClauses(engine, model, params.Include)
		joins = append(joins, joinClauses...)
		args = append(args, joinArgs...)
		argCount += len(joinArgs)
	}

	// Add all JOIN clauses
	for _, join := range joins {
		query.WriteString(" " + join)
	}

	// Add WHERE conditions
	if len(params.Where) > 0 {
		query.WriteString(" WHERE ")
		conditions := make([]string, 0, len(params.Where))

		for _, cond := range params.Where {
			if cond.Op == OpIn {
				placeholders := make([]string, len(cond.Value.([]interface{})))
				for i := range placeholders {
					placeholders[i] = engine.dialect.GetPlaceholder(argCount)
					argCount++
				}
				conditions = append(conditions,
					fmt.Sprintf("%s %s (%s)",
						cond.Field,
						cond.Op,
						strings.Join(placeholders, ","),
					),
				)
				args = append(args, cond.Value.([]interface{})...)
			} else {
				conditions = append(conditions,
					fmt.Sprintf("%s %s %s",
						cond.Field,
						cond.Op,
						engine.dialect.GetPlaceholder(argCount),
					),
				)
				args = append(args, cond.Value)
				argCount++
			}
		}

		query.WriteString(strings.Join(conditions, " AND "))
	}

	// Add ORDER BY
	if len(params.OrderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(params.OrderBy, ", "))
	}

	// Add LIMIT and OFFSET
	if params.Limit != nil {
		query.WriteString(fmt.Sprintf(" LIMIT %d", *params.Limit))
	}
	if params.Offset != nil {
		query.WriteString(fmt.Sprintf(" OFFSET %d", *params.Offset))
	}

	engine.logger.Info("Query", "query", query.String(), "args", args)

	return query.String(), args
}

func buildSelectClauses(model interface{}, includes map[string]P) string {
	mainTable := getTableName(model)
	var selects []string

	// Add all fields from main table
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Add main table fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if dbTag := field.Tag.Get(DB_COL_LOOKUP); dbTag != "" {
			selects = append(selects,
				fmt.Sprintf("%s.%s as %s_%s",
					mainTable,
					dbTag,
					mainTable,
					dbTag,
				),
			)
		}
	}

	// Add fields from included tables
	for _, includedParams := range includes {
		includedModel := reflect.New(reflect.TypeOf(includedParams)).Elem().Interface()
		includedTable := getTableName(includedModel)

		t := reflect.TypeOf(includedModel)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if dbTag := field.Tag.Get(DB_COL_LOOKUP); dbTag != "" {
				selects = append(selects,
					fmt.Sprintf("%s.%s as %s_%s",
						includedTable,
						dbTag,
						includedTable,
						dbTag,
					),
				)
			}
		}
	}

	return strings.Join(selects, ", ")
}

func buildJoinClauses(engine *Engine, model interface{}, includes map[string]P) ([]string, []interface{}) {
	var joins []string
	var args []interface{}

	mainTable := getTableName(model)
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for name, includedParams := range includes {
		var foreignKey, referencedKey string
		joinType := "LEFT JOIN"

		// Find the relationship field
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Name == name {
				// Check if it's a slice (HasMany) or pointer (HasOne/BelongsTo)
				if field.Type.Kind() == reflect.Slice {
					foreignKey = strings.ToLower(t.Name()) + "_id"
					referencedKey = "id"
				} else if field.Type.Kind() == reflect.Ptr {
					// Check for foreign key field
					fkField := t.Field(i - 1) // Assume foreign key field is just before the relationship field
					if fkTag := fkField.Tag.Get(DB_COL_LOOKUP); strings.HasSuffix(fkTag, "_id") {
						foreignKey = fkTag
						referencedKey = "id"
					}
				}
				break
			}
		}
		includedTable := strings.ToLower(name) + "s"

		joins = append(joins, fmt.Sprintf("%s %s ON %s.%s = %s.%s",
			joinType,
			strings.ToLower(name)+"s",
			mainTable,
			foreignKey,
			includedTable,
			referencedKey,
		))

		// Add any WHERE conditions from the included params
		if len(includedParams.Where) > 0 {
			for ci, cond := range includedParams.Where {
				joins = append(joins, fmt.Sprintf("AND %s.%s %s %s",
					includedTable,
					cond.Field,
					cond.Op,
					engine.dialect.GetPlaceholder(ci+1),
				))
				args = append(args, cond.Value)
			}
		}
	}

	return joins, args
}

// Modify scanStruct to handle aliased columns
func scanStruct(rows *sql.Rows, dest interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(dest).Elem()
	t := v.Type()

	values := make([]interface{}, len(columns))
	for i, column := range columns {
		// Split column name if it's aliased (table_column)
		parts := strings.Split(column, "_")
		if len(parts) > 1 {
			column = parts[len(parts)-1]
		}

		for j := 0; j < t.NumField(); j++ {
			field := t.Field(j)
			if dbTag := field.Tag.Get(DB_COL_LOOKUP); dbTag == column {
				values[i] = v.Field(j).Addr().Interface()
				break
			}
		}
		if values[i] == nil {
			var placeholder interface{}
			values[i] = &placeholder
		}
	}

	return rows.Scan(values...)
}

func getTableName(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.String {
		return strings.ToLower(t.Name()) + "s"
	}

	return strings.ToLower(t.Name()) + "s"
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func setValue(data interface{}, fieldName string, value interface{}) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if f := v.FieldByName(fieldName); f.IsValid() && f.CanSet() {
		val := reflect.ValueOf(value)
		if f.Type() != val.Type() {
			val = val.Convert(f.Type())
		}
		f.Set(val)
	}
}
