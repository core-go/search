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
	return NewSearchServiceWithMapper(db, queryBuilder, modelType, nil)
}
func NewSearchServiceWithMapper(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type, mapper Mapper) *SqlSearchService {
	searchBuilder := NewSearchResultBuilderWithMapper(db, queryBuilder, modelType, mapper)
	return &SqlSearchService{searchBuilder}
}
func NewDefaultSearchServiceWithMapper(db *sql.DB, tableName string, modelType reflect.Type, mapper Mapper) *SqlSearchService {
	driverName := GetDriverName(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driverName)
	searchBuilder := NewSearchResultBuilderWithMapper(db, queryBuilder, modelType, mapper)
	return &SqlSearchService{searchBuilder}
}
func NewDefaultSearchService(db *sql.DB, tableName string, modelType reflect.Type) *SqlSearchService {
	return NewDefaultSearchServiceWithMapper(db, tableName, modelType, nil)
}
func (s *SqlSearchService) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	return s.SearchBuilder.BuildSearchResult(ctx, m)
}
