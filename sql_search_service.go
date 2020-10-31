package search

import (
	"context"
	"database/sql"
	"reflect"
)

type SqlSearchService struct {
	SearchBuilder SearchResultBuilder
}

func NewSqlSearchService(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type, mapper Mapper) *SqlSearchService {
	searchBuilder := NewSearchResultBuilder(db, queryBuilder, modelType, mapper)
	return &SqlSearchService{searchBuilder}
}
func NewDefaultSqlSearchService(db *sql.DB, tableName string, modelType reflect.Type, mapper Mapper) *SqlSearchService {
	queryBuilder := NewQueryBuilder(tableName, modelType)
	searchBuilder := NewSearchResultBuilder(db, queryBuilder, modelType, mapper)
	return &SqlSearchService{searchBuilder}
}
func (s *SqlSearchService) Search(ctx context.Context, m interface{}) (*SearchResult, error) {
	return s.SearchBuilder.BuildSearchResult(ctx, m)
}
