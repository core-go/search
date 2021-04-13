package search

import (
	"context"
	"database/sql"
	"reflect"
)

func NewSearcherWithQuery(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), options ...func(context.Context, interface{}) (interface{}, error)) *Searcher {
	builder := NewSearchBuilder(db, modelType, buildQuery, options...)
	return NewSearcher(builder.Search)
}

func NewDefaultSearcher(db *sql.DB, tableName string, modelType reflect.Type, options ...func(context.Context, interface{}) (interface{}, error)) *Searcher {
	driver := getDriver(db)
	buildParam := getBuild(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driver, buildParam)
	builder := NewSearchBuilder(db, modelType, queryBuilder.BuildQuery, options...)
	return NewSearcher(builder.Search)
}
