package goorm

import (
	"database/sql"
	"fmt"
	"reflect"
)

type RelationType string

const (
	HasOne    RelationType = "hasOne"
	HasMany   RelationType = "hasMany"
	BelongsTo RelationType = "belongsTo"
)

// Relationship metadata
type Relation struct {
	Name       string
	Type       RelationType
	Model      interface{}
	ForeignKey string
	References string
}

// BaseModel provides generic query building capabilities
type BaseModel[T any] struct {
	db        *DB
	engine    *Engine
	logger    Logger
	tx        *sql.Tx
	relations map[string]Relation
}

func NewBaseModel[T any](db *DB) *BaseModel[T] {
	return &BaseModel[T]{
		db:     db,
		engine: db.engine,
		logger: db.logger,
		tx:     db.tx,
	}
}

func (m *BaseModel[T]) RegisterRelation(rel Relation) {
	if m.relations == nil {
		m.relations = make(map[string]Relation)
	}
	m.relations[rel.Name] = rel
}

func (m *BaseModel[T]) withTx(tx *sql.Tx) interface{} {
	return &BaseModel[T]{
		db:     m.db,
		engine: m.engine,
		logger: m.logger,
		tx:     tx,
	}
}

func (m *BaseModel[T]) loadRelations(results *[]T, includes map[string]P) error {
	fmt.Println("ITEM", results)
	if results == nil || len(*results) == 0 {
		return nil
	}

	for name, includeParams := range includes {
		relation, ok := m.relations[name]
		if !ok {
			m.logger.Debug("Relation not found", "relation", name)
			continue
		}

		if relation.Model == nil {
			m.logger.Debug("Relation model is nil", "relation", name)
			continue
		}

		switch relation.Type {
		case HasMany:
			err := m.loadHasManyRelation(results, relation, includeParams)
			if err != nil {
				return fmt.Errorf("load has many relation %s: %w", name, err)
			}
		case HasOne:
			err := m.loadHasOneRelation(results, relation, includeParams)
			if err != nil {
				return fmt.Errorf("load has one relation %s: %w", name, err)
			}
		case BelongsTo:

			err := m.loadBelongsToRelation(results, relation, includeParams)
			if err != nil {
				return fmt.Errorf("load belongs to relation %s: %w", name, err)
			}
		}
	}
	return nil
}

func (m *BaseModel[T]) loadHasManyRelation(results *[]T, rel Relation, params P) error {
	if results == nil || len(*results) == 0 || rel.Model == nil {
		return nil
	}

	// Get parent IDs
	var ids []interface{}
	for _, item := range *results {
		// if item == nil {
		//     continue
		// }
		v := reflect.ValueOf(item)
		if !v.IsValid() {
			continue
		}
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				continue
			}
			v = v.Elem()
		}

		field := v.FieldByName(rel.References)
		if !field.IsValid() {
			m.logger.Debug("Reference field not found", "field", rel.References)
			continue
		}

		ids = append(ids, field.Interface())
	}

	if len(ids) == 0 {
		return nil
	}

	// Add ID condition to existing params
	params.Where = append(params.Where, In(rel.ForeignKey, ids...))

	// Execute query for related records
	query, args := buildQuery(m.engine, rel.Model, params)
	m.logger.Debug("Loading relation", "relation", rel.Name, "query", query, "args", args)

	var rows *sql.Rows
	var err error
	if m.tx != nil {
		rows, err = m.tx.Query(query, args...)
	} else {
		rows, err = m.engine.db.Query(query, args...)
	}
	if err != nil {
		return fmt.Errorf("query relation: %w", err)
	}
	defer rows.Close()

	// Create map of related records grouped by foreign key
	relMap := make(map[interface{}][]reflect.Value)
	relType := reflect.TypeOf(rel.Model)
	if relType.Kind() == reflect.Ptr {
		relType = relType.Elem()
	}

	for rows.Next() {
		// Create new instance of related type
		relPtr := reflect.New(relType)
		if err := scanStruct(rows, relPtr.Interface()); err != nil {
			return fmt.Errorf("scan relation: %w", err)
		}

		v := relPtr.Elem()
		field := v.FieldByName(rel.ForeignKey)
		if !field.IsValid() {
			continue
		}

		fkValue := field.Interface()
		relMap[fkValue] = append(relMap[fkValue], v)
	}

	// Assign related records back to parent records
	for i := range *results {
		// if (*results)[i] == nil {
		//     continue
		// }

		v := reflect.ValueOf(&(*results)[i]).Elem()
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				continue
			}
			v = v.Elem()
		}

		referencesField := v.FieldByName(rel.References)
		if !referencesField.IsValid() {
			continue
		}
		parentID := referencesField.Interface()

		relField := v.FieldByName(rel.Name)
		if !relField.IsValid() || !relField.CanSet() {
			continue
		}

		if related, ok := relMap[parentID]; ok && len(related) > 0 {
			relSlice := reflect.MakeSlice(relField.Type(), len(related), len(related))
			for j, rel := range related {
				relSlice.Index(j).Set(rel)
			}
			relField.Set(relSlice)
		} else {
			relField.Set(reflect.MakeSlice(relField.Type(), 0, 0))
		}
	}

	return nil
}

func (m *BaseModel[T]) loadHasOneRelation(results *[]T, rel Relation, params P) error {
	if results == nil || len(*results) == 0 || rel.Model == nil {
		return nil
	}

	var ids []interface{}
	for _, item := range *results {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		id := v.FieldByName(rel.References).Interface()
		ids = append(ids, id)
	}

	params.Where = append(params.Where, In(rel.ForeignKey, ids...))
	query, args := buildQuery(m.engine, rel.Model, params)
	m.logger.Debug("Loading hasOne relation", "relation", rel.Name, "query", query, "args", args)

	var rows *sql.Rows
	var err error
	if m.tx != nil {
		rows, err = m.tx.Query(query, args...)
	} else {
		rows, err = m.engine.db.Query(query, args...)
	}
	if err != nil {
		return fmt.Errorf("query relation %s: %w", rel.Name, err)
	}
	defer rows.Close()

	relMap := make(map[interface{}]reflect.Value)
	relType := reflect.TypeOf(rel.Model)

	for rows.Next() {
		relPtr := reflect.New(relType)
		if err := scanStruct(rows, relPtr.Interface()); err != nil {
			return fmt.Errorf("scan relation %s: %w", rel.Name, err)
		}

		v := relPtr.Elem()
		fkValue := v.FieldByName(rel.ForeignKey).Interface()
		if _, exists := relMap[fkValue]; !exists {
			relMap[fkValue] = v
		}
	}

	for i := range *results {
		parent := &(*results)[i]
		v := reflect.ValueOf(parent)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		parentID := v.FieldByName(rel.References).Interface()
		relField := v.FieldByName(rel.Name)

		if !relField.IsValid() {
			continue
		}

		if related, ok := relMap[parentID]; ok {
			if relField.Kind() == reflect.Ptr {
				newPtr := reflect.New(related.Type())
				newPtr.Elem().Set(related)
				relField.Set(newPtr)
			}
		}
	}

	return nil
}

func (m *BaseModel[T]) loadBelongsToRelation(results *[]T, rel Relation, params P) error {
	if results == nil || len(*results) == 0 || rel.Model == nil {
		return nil
	}

	var ids []interface{}
	seen := make(map[interface{}]bool)
	for _, item := range *results {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		fk := v.FieldByName(rel.ForeignKey).Interface()
		if !seen[fk] {
			ids = append(ids, fk)
			seen[fk] = true
		}
	}

	params.Where = append(params.Where, In("id", ids...))
	query, args := buildQuery(m.engine, rel.Model, params)

	var rows *sql.Rows
	var err error
	if m.tx != nil {
		rows, err = m.tx.Query(query, args...)
	} else {
		rows, err = m.engine.db.Query(query, args...)
	}
	if err != nil {
		return fmt.Errorf("query relation %s: %w", rel.Name, err)
	}
	defer rows.Close()

	relMap := make(map[interface{}]reflect.Value)
	relType := reflect.TypeOf(rel.Model)

	for rows.Next() {
		relPtr := reflect.New(relType)
		if err := scanStruct(rows, relPtr.Interface()); err != nil {
			return fmt.Errorf("scan relation %s: %w", rel.Name, err)
		}

		v := relPtr.Elem()
		idValue := v.FieldByName("ID").Interface()
		relMap[idValue] = v
	}

	for i := range *results {
		child := &(*results)[i]
		v := reflect.ValueOf(child)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		foreignKey := v.FieldByName(rel.ForeignKey).Interface()
		relField := v.FieldByName(rel.Name)

		if !relField.IsValid() {
			continue
		}

		if parent, ok := relMap[foreignKey]; ok {
			if relField.Kind() == reflect.Ptr {
				newPtr := reflect.New(parent.Type())
				newPtr.Elem().Set(parent)
				relField.Set(newPtr)
			}
		}
	}

	return nil
}
