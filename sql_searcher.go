package search

import (
	"context"
	"database/sql"
	"reflect"
)

func NewSearcherWithMap(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *Searcher {
	builder := NewSearchBuilderWithMap(db, modelType, buildQuery, mp, options...)
	return NewSearcher(builder.Search)
}

func NewSearcherWithQuery(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), options ...func(context.Context, interface{}) (interface{}, error)) *Searcher {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewSearcherWithMap(db, modelType, buildQuery, mp)
}
func NewDefaultSearcherWithMap(db *sql.DB, tableName string, modelType reflect.Type, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *Searcher {
	driver := getDriver(db)
	buildParam := getBuild(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driver, buildParam)
	builder := NewSearchBuilderWithMap(db, modelType, queryBuilder.BuildQuery, mp, options...)
	return NewSearcher(builder.Search)
}
func NewDefaultSearcher(db *sql.DB, tableName string, modelType reflect.Type, options ...func(context.Context, interface{}) (interface{}, error)) *Searcher {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewDefaultSearcherWithMap(db, tableName, modelType, mp)
}
