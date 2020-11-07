package search

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

const (
	DriverPostgres      = "postgres"
	DriverMysql         = "mysql"
	DriverMssql         = "mssql"
	DriverOracle        = "oracle"
	DriverNotSupport    = "no support"
	DefaultPagingFormat = " LIMIT %s OFFSET %s"
	OraclePagingFormat  = " OFFSET %s ROWS FETCH NEXT %s ROWS ONLY"
	desc                = "DESC"
	asc                 = "ASC"
)

type DefaultSearchResultBuilder struct {
	Database     *sql.DB
	QueryBuilder QueryBuilder
	ModelType    reflect.Type
	Mapper       Mapper
	DriverName   string
}

func NewSearchResultBuilder(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type, mapper Mapper) *DefaultSearchResultBuilder {
	driverName := getDriverName(db)
	builder := &DefaultSearchResultBuilder{Database: db, QueryBuilder: queryBuilder, ModelType: modelType, Mapper: mapper, DriverName: driverName}
	return builder
}
func NewDefaultSearchResultBuilder(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type) *DefaultSearchResultBuilder {
	return NewSearchResultBuilder(db, queryBuilder, modelType, nil)
}
func (b *DefaultSearchResultBuilder) BuildSearchResult(ctx context.Context, m interface{}) (interface{}, int64, error) {
	sql, params := b.QueryBuilder.BuildQuery(m)
	var searchModel = GetSearchModel(m)
	return BuildFromQuery(ctx, b.Database, b.ModelType, sql, params, searchModel.PageIndex, searchModel.PageSize, searchModel.FirstPageSize, b.Mapper, b.DriverName)
}
func GetSearchModel(m interface{}) *SearchModel {
	if sModel, ok := m.(*SearchModel); ok {
		return sModel
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*SearchModel); ok {
				return sModel1
			}
		}
	}
	return nil
}
func IsLastPage(models interface{}, count int64, pageIndex int64, pageSize int64, initPageSize int64) bool {
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
	return receivedItems >= count
}
func replaceParameters(sql string, number int, prefix string) string {
	for i := 0; i < number; i++ {
		count := i + 1
		sql = strings.Replace(sql, "?", prefix+fmt.Sprintf("%v", count), 1)
	}
	return sql
}

func BuildQueryByDriver(sql string, number int, driverName string) string {
	switch driverName {
	case DriverPostgres:
		return replaceParameters(sql, number, "$")
	case DriverOracle:
		return replaceParameters(sql, number, ":val")
	default:
		return replaceParameters(sql, number, "?")
	}
}

func BuildFromQuery(ctx context.Context, db *sql.DB, modelType reflect.Type, query string, params []interface{}, pageIndex int64, pageSize int64, initPageSize int64, mapper Mapper, driverName string) (interface{}, int64, error) {
	var total int64
	query = BuildQueryByDriver(query, len(params), driverName)
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	models := reflect.New(modelsType).Interface()
	queryPaging := BuildPagingQuery(query, pageIndex, pageSize, initPageSize, driverName)
	queryCount, paramsCount := BuildCountQuery(query, params)
	fieldsIndex, er12 := GetColumnIndexes(modelType, driverName)
	if er12 != nil {
		return nil, -1, er12
	}
	er1 := Query(db, models, modelType, fieldsIndex, queryPaging, params...)
	if er1 != nil {
		return nil, -1, er1
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
				offset = pageSize*(pageIndex-2) + initPageSize
			}
		} else {
			limit = pageSize
			offset = pageSize * (pageIndex - 1)
		}

		var pagingQuery string
		if driver == DriverOracle {
			pagingQuery = fmt.Sprintf(OraclePagingFormat, strconv.Itoa(int(offset)), strconv.Itoa(int(limit)))
		} else {
			pagingQuery = fmt.Sprintf(DefaultPagingFormat, strconv.Itoa(int(limit)), strconv.Itoa(int(offset)))
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

func BuildSearchResult(ctx context.Context, models interface{}, count int64, pageIndex int64, pageSize int64, initPageSize int64, mapper Mapper) (interface{}, int64, error) {
	searchResult := SearchResult{}
	searchResult.Total = count
	if mapper == nil {
		return models, count, nil
	}
	r2, er3 := mapper.DbToModels(ctx, models)
	if er3 != nil {
		return models, count, nil
	}
	return r2, count, er3
}

func getDriverName(db *sql.DB) string {
	driver := reflect.TypeOf(db.Driver()).String()
	switch driver {
	case "*pq.Driver":
		return DriverPostgres
	case "*mysql.MySQLDriver":
		return DriverMysql
	case "*mssql.Driver":
		return DriverMssql
	case "*godror.drv":
		return DriverOracle
	default:
		log.Panicf(DriverNotSupport)
		return DriverNotSupport
	}
}
