package search

import (
	"context"
	"database/sql"
	"reflect"
)

type Searcher struct {
	Search func(ctx context.Context, m interface{}) (interface{}, int64, error)
}

func NewSearcherWithMap(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *Searcher {
	var extractSearch func(m interface{}) (int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearch = options[0]
	}
	builder := NewSearchBuilderWithMap(db, modelType, buildQuery, mp, extractSearch)
	return &Searcher{Search: builder.Search}
}
func NewSearcherWithFunc(search func(context.Context, interface{}) (interface{}, int64, error)) *Searcher {
	return &Searcher{Search: search}
}
func NewSearcher(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), options ...func(context.Context, interface{}) (interface{}, error)) *Searcher {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewSearcherWithMap(db, modelType, buildQuery, mp)
}
func NewDefaultSearcherWithMap(db *sql.DB, tableName string, modelType reflect.Type, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *Searcher {
	var extractSearch func(m interface{}) (int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearch = options[0]
	}
	driverName := GetDriver(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driverName)
	builder := NewSearchBuilderWithMap(db, modelType, queryBuilder.BuildQuery, mp, extractSearch)
	return &Searcher{Search: builder.Search}
}
func NewDefaultSearcher(db *sql.DB, tableName string, modelType reflect.Type, options ...func(context.Context, interface{}) (interface{}, error)) *Searcher {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewDefaultSearcherWithMap(db, tableName, modelType, mp, nil)
}
