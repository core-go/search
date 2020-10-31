package search

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
)

const desc = "DESC"
const asc = "ASC"

type DefaultSearchResultBuilder struct {
	Database     *sql.DB
	QueryBuilder QueryBuilder
	ModelType    reflect.Type
	Mapper       Mapper
}

func NewSearchResultBuilder(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type, mapper Mapper) *DefaultSearchResultBuilder {
	builder := &DefaultSearchResultBuilder{Database: db, QueryBuilder: queryBuilder, ModelType: modelType, Mapper: mapper}
	return builder
}

func (b *DefaultSearchResultBuilder) BuildSearchResult(ctx context.Context, m interface{}) (*SearchResult, error) {
	sql, params := b.QueryBuilder.BuildQuery(m)
	var searchModel *SearchModel
	if sModel, ok := m.(*SearchModel); ok {
		searchModel = sModel
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*SearchModel); ok {
				searchModel = sModel1
			}
		}
	}
	return BuildFromQuery(ctx, b.Database, b.ModelType, sql, params, searchModel.Page, searchModel.Limit, searchModel.FirstLimit, b.Mapper)
}

func BuildFromQuery(ctx context.Context, db *sql.DB, modelType reflect.Type, query string, params []interface{}, pageIndex int64, pageSize int64, initPageSize int64, mapper Mapper) (*SearchResult, error) {
	var countSelect struct {
		Total int
	}
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	models := reflect.New(modelsType).Interface()
	queryPaging, paramsPaging := BuildPagingQuery(query, params, pageIndex, pageSize, initPageSize)
	queryCount, paramsCount := BuildCountQuery(query, params)
	er1 := Query(db, models, queryPaging, paramsPaging...)
	if er1 != nil {
		return nil, er1
	}
	er2 := Query(db, &countSelect, queryCount, paramsCount...)
	if er2 != nil {
		countSelect.Total = 0
	}
	return BuildSearchResult(ctx, models, int64(countSelect.Total), pageIndex, pageSize, initPageSize, mapper)
}

func BuildPagingQuery(sql string, params []interface{}, pageIndex int64, pageSize int64, initPageSize int64) (string, []interface{}) {
	if pageSize > 0 {
		sql = sql + ` LIMIT ? OFFSET ? `
		if initPageSize > 0 {
			if pageIndex == 1 {
				params = append(params, initPageSize, 0)
			} else {
				params = append(params, pageSize, pageSize*(pageIndex-2)+initPageSize)
			}
		} else {
			params = append(params, pageSize, pageSize*(pageIndex-1))
		}
	}
	return sql, params
}

func BuildCountQuery(sql string, params []interface{}) (string, []interface{}) {
	i := strings.Index(sql, "SELECT ")
	if i < 0 {
		return sql, params
	}
	j := strings.Index(sql, " FROM ")
	if j < 0 {
		return sql, params
	}
	k := strings.Index(sql, " ORDER BY ")
	h := strings.Index(sql, " DISTINCT ")
	if h > 0 {
		sql3 := `SELECT count(*) as total FROM (` + sql[i:] + `) as main`
		return sql3, params
	}
	if k > 0 {
		sql3 := `SELECT count(*) as total ` + sql[j:k]
		return sql3, params
	} else {
		sql3 := `SELECT count(*) as total ` + sql[j:]
		return sql3, params
	}
}

func BuildSearchResult(ctx context.Context, models interface{}, count int64, pageIndex int64, pageSize int64, initPageSize int64, mapper Mapper) (*SearchResult, error) {
	searchResult := SearchResult{}
	searchResult.Total = count

	searchResult.Last = false
	lengthModels := int64(reflect.Indirect(reflect.ValueOf(models)).Len())
	var receivedItems int64

	if initPageSize > 0 {
		if pageIndex == 1 {
			receivedItems = initPageSize
		} else if pageIndex > 1 {
			receivedItems = pageSize*(pageIndex-2) + initPageSize + lengthModels
		}
	} else {
		receivedItems = pageSize*(pageIndex-1) + lengthModels
	}
	searchResult.Last = receivedItems >= count

	if mapper == nil {
		searchResult.Results = models
		return &searchResult, nil
	}
	r2, er3 := mapper.DbToModels(ctx, models)
	if er3 != nil {
		searchResult.Results = models
		return &searchResult, nil
	}
	searchResult.Results = r2
	return &searchResult, er3
}
