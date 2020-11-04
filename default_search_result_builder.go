package search

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	DRIVER_POSTGRES 	= "postgres"
	DRIVER_MYSQL    	= "mysql"
	DRIVER_MSSQL    	= "mssql"
	DRIVER_ORACLE    	= "oracle"
	DRIVER_NOT_SUPPORT  = "no support"
	DEFAULT_PAGING_FORMAT = " LIMIT %s OFFSET %s"
	ORACLE_PAGING_FORMAT  = " OFFSET %s ROWS FETCH NEXT %s ROWS ONLY"
	desc = "DESC"
	asc = "ASC"
)


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
	return BuildFromQuery(ctx, b.Database, b.ModelType, sql, params, searchModel.PageIndex, searchModel.PageSize, searchModel.FirstPageSize, b.Mapper)
}

func BuildFromQuery(ctx context.Context, db *sql.DB, modelType reflect.Type, query string, params []interface{}, pageIndex int64, pageSize int64, initPageSize int64, mapper Mapper) (*SearchResult, error) {
	var total int64
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	models := reflect.New(modelsType).Interface()
	queryPaging := BuildPagingQuery(query, pageIndex, pageSize, initPageSize, getDriver(db))
	queryCount, paramsCount := BuildCountQuery(query, params)
	fieldsIndex, er0 := GetColumnIndexes(modelType)
	if er0 != nil {
		return nil, er0
	}
	er1 := Query(db, models, modelType, fieldsIndex, queryPaging, params...)
	if er1 != nil {
		return nil, er1
	}
	total, er2 := Count(db, queryCount, paramsCount...)
	if er2 != nil {
		total = 0
	}
	return BuildSearchResult(ctx, models, total, pageIndex, pageSize, initPageSize, mapper)
}

func BuildPagingQuery(sql string, pageIndex int64, pageSize int64, initPageSize int64, driver string) string {
	if pageSize > 0 {
		var limit, offset int64
		if initPageSize > 0 {
			if pageIndex == 1 {
				limit = initPageSize
				offset = 0
			} else {
				limit = pageSize
				offset = pageSize*(pageIndex-2)+initPageSize
			}
		} else {
			limit = pageSize
			offset = pageSize*(pageIndex-1)
		}

		var pagingQuery string
		if driver == DRIVER_ORACLE {
			pagingQuery = fmt.Sprintf(ORACLE_PAGING_FORMAT, strconv.Itoa(int(offset)), strconv.Itoa(int(limit)))
		} else {
			pagingQuery = fmt.Sprintf(DEFAULT_PAGING_FORMAT, strconv.Itoa(int(limit)), strconv.Itoa(int(offset)))
		}
		sql += pagingQuery
	}

	return sql
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

func getDriver(db *sql.DB) string {
	driver := reflect.TypeOf(db.Driver()).String()
	switch driver {
	case "*postgres.Driver":
		return DRIVER_POSTGRES
	case "*mysql.MySQLDriver":
		return DRIVER_MYSQL
	case "*mssql.Driver":
		return DRIVER_MSSQL
	default:
		return DRIVER_NOT_SUPPORT
	}
}
