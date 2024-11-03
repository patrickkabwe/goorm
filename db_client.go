package goorm

import (
	"database/sql"
	"reflect"
)

type IDB interface {
	WithTx(tx *sql.Tx) IDB
}

// DB represents the database instance
type DB struct {
	engine *Engine
	logger Logger
	tx     *sql.Tx
}

// NewDB creates a new database instance with initialized models
func NewDB(engine *Engine, logger Logger) *DB {
	db := &DB{
		engine: engine,
		logger: logger,
	}
	return db
}

func (db *DB) WithTx(tx *sql.Tx) IDB {
	// Create a new dynamic DB struct with the same fields
	newDB := reflect.New(reflect.TypeOf(db).Elem()).Interface().(*DB)
	newDB.engine = db.engine
	newDB.logger = db.logger
	newDB.tx = tx

	// Copy all model fields with the new transaction
	v := reflect.ValueOf(db).Elem()
	newV := reflect.ValueOf(newDB).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			// If it's a model pointer
			if modelField := v.Field(i); modelField.CanInterface() {
				if model, ok := modelField.Interface().(interface{ withTx(*sql.Tx) interface{} }); ok {
					newV.Field(i).Set(reflect.ValueOf(model.withTx(tx)))
				}
			}
		}
	}

	return newDB
}

// Close closes the database connection
func (d *DB) Close() {
	d.engine.Close()
}

func (db *DB) Migrate(model ...any) error {
	for _, m := range model {
		db.logger.Info("CREATE TABLE " + getTableName(m))
	}

	return nil
}
