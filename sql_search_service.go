package search

import (
	"context"
	"database/sql"
	"reflect"
)

type SqlSearchService struct {
	SearchBuilder SearchResultBuilder
}

func NewSearchService(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type) *SqlSearchService {
	return NewSearchServiceWithMapper(db, queryBuilder, modelType, ExtractSearch, nil)
}
func NewSearchServiceWithExtractor(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type, extractSearch func(m interface{}) (int64, int64, int64, error)) *SqlSearchService {
	return NewSearchServiceWithMapper(db, queryBuilder, modelType, extractSearch, nil)
}
func NewSearchServiceWithMapper(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type, extractSearch func(m interface{}) (int64, int64, int64, error), mapper Mapper) *SqlSearchService {
	searchBuilder := NewSearchResultBuilderWithMapper(db, queryBuilder, modelType, extractSearch, mapper)
	return &SqlSearchService{searchBuilder}
}
func NewDefaultSearchServiceWithMapper(db *sql.DB, tableName string, modelType reflect.Type, extractSearch func(m interface{}) (int64, int64, int64, error), mapper Mapper) *SqlSearchService {
	driverName := GetDriverName(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driverName)
	searchBuilder := NewSearchResultBuilderWithMapper(db, queryBuilder, modelType, extractSearch, mapper)
	return &SqlSearchService{searchBuilder}
}
func NewDefaultSearchService(db *sql.DB, tableName string, modelType reflect.Type) *SqlSearchService {
	return NewDefaultSearchServiceWithMapper(db, tableName, modelType, ExtractSearch, nil)
}
func NewDefaultSearchServiceWithExtractor(db *sql.DB, tableName string, modelType reflect.Type, extractSearch func(m interface{}) (int64, int64, int64, error)) *SqlSearchService {
	return NewDefaultSearchServiceWithMapper(db, tableName, modelType, extractSearch, nil)
}
func (s *SqlSearchService) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	return s.SearchBuilder.BuildSearchResult(ctx, m)
}
