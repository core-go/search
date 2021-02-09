package search

import (
	"context"
	"database/sql"
	"reflect"
)

type SqlSearchService struct {
	SearchBuilder SearchResultBuilder
}

func NewSearchServiceWithMap(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *SqlSearchService {
	var extractSearch func(m interface{}) (int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearch = options[0]
	}
	searchBuilder := NewSearchResultBuilderWithMap(db, modelType, buildQuery, mp, extractSearch)
	return &SqlSearchService{searchBuilder}
}
func NewSearchService(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), options ...func(context.Context, interface{}) (interface{}, error)) *SqlSearchService {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewSearchServiceWithMap(db, modelType, buildQuery, mp)
}
func NewDefaultSearchServiceWithMap(db *sql.DB, tableName string, modelType reflect.Type, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *SqlSearchService {
	var extractSearch func(m interface{}) (int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearch = options[0]
	}
	driverName := GetDriver(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driverName)
	searchBuilder := NewSearchResultBuilderWithMap(db, modelType, queryBuilder.BuildQuery, mp, extractSearch)
	return &SqlSearchService{searchBuilder}
}
func NewDefaultSearchService(db *sql.DB, tableName string, modelType reflect.Type, options ...func(context.Context, interface{}) (interface{}, error)) *SqlSearchService {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewDefaultSearchServiceWithMap(db, tableName, modelType, mp, nil)
}
func (s *SqlSearchService) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	return s.SearchBuilder.BuildSearchResult(ctx, m)
}
