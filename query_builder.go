package goorm

import (
	"database/sql"
	"strings"
)

type QueryBuilder struct {
	query strings.Builder
	db    *sql.DB
}

func NewQueryBuilder(db *sql.DB) *QueryBuilder {
	return &QueryBuilder{
		db: db,
	}
}

func (q *QueryBuilder) Close() {
	q.db.Close()
}

func (q *QueryBuilder) Select(fields ...string) *QueryBuilder {
	q.query.WriteString("SELECT ")
	for i, field := range fields {
		if i > 0 {
			q.query.WriteString(", ")
		}
		q.query.WriteString(field)
	}
	return q
}

func (q *QueryBuilder) From(table string) *QueryBuilder {
	q.query.WriteString(" FROM " + table)
	return q
}

func (q *QueryBuilder) Where(condition string) *QueryBuilder {
	q.query.WriteString(" WHERE " + condition)
	return q
}

func (q *QueryBuilder) GetSql() string {
	return q.query.String()
}
