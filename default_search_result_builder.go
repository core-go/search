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
	DriverPostgres      = "postgres"
	DriverMysql         = "mysql"
	DriverMssql         = "mssql"
	DriverOracle        = "oracle"
	DriverNotSupport    = "no support"
	DefaultPagingFormat = " limit %s offset %s "
	OraclePagingFormat  = " offset %s rows fetch next %s rows only "
	desc                = "desc"
	asc                 = "asc"
)

type DefaultSearchResultBuilder struct {
	Database      *sql.DB
	QueryBuilder  QueryBuilder
	ModelType     reflect.Type
	extractSearch func(m interface{}) (int64, int64, int64, error)
	Mapper        Mapper
	DriverName    string
}

func NewSearchResultBuilderWithMapper(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type, extractSearch func(m interface{}) (int64, int64, int64, error), mapper Mapper) *DefaultSearchResultBuilder {
	driverName := GetDriverName(db)
	builder := &DefaultSearchResultBuilder{Database: db, QueryBuilder: queryBuilder, ModelType: modelType, extractSearch: extractSearch, Mapper: mapper, DriverName: driverName}
	return builder
}
func NewSearchResultBuilder(db *sql.DB, queryBuilder QueryBuilder, modelType reflect.Type) *DefaultSearchResultBuilder {
	return NewSearchResultBuilderWithMapper(db, queryBuilder, modelType, ExtractSearch, nil)
}
func NewDefaultSearchResultBuilderWithMapper(db *sql.DB, tableName string, modelType reflect.Type, extractSearch func(m interface{}) (int64, int64, int64, error), mapper Mapper) *DefaultSearchResultBuilder {
	driverName := GetDriverName(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driverName)
	return NewSearchResultBuilderWithMapper(db, queryBuilder, modelType, extractSearch, mapper)
}
func NewDefaultSearchResultBuilder(db *sql.DB, tableName string, modelType reflect.Type, extractSearch func(m interface{}) (int64, int64, int64, error)) *DefaultSearchResultBuilder {
	driverName := GetDriverName(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driverName)
	return NewSearchResultBuilderWithMapper(db, queryBuilder, modelType, extractSearch, nil)
}
func (b *DefaultSearchResultBuilder) BuildSearchResult(ctx context.Context, m interface{}) (interface{}, int64, error) {
	sql, params := b.QueryBuilder.BuildQuery(m)
	pageIndex, pageSize, firstPageSize, err := b.extractSearch(m)
	if err != nil {
		return nil, 0, err
	}
	return BuildFromQuery(ctx, b.Database, b.ModelType, sql, params, pageIndex, pageSize, firstPageSize, b.Mapper, b.DriverName)
}

func BuildFromQuery(ctx context.Context, db *sql.DB, modelType reflect.Type, query string, params []interface{}, pageIndex int64, pageSize int64, initPageSize int64, mapper Mapper, driverName string) (interface{}, int64, error) {
	var total int64
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	models := reflect.New(modelsType).Interface()
	if driverName == DriverOracle {
		queryPaging := BuildPagingQueryByDriver(query, pageIndex, pageSize, initPageSize, driverName)
		er1 := QueryAndCount(db, models, &total, driverName, queryPaging, params...)
		if er1 != nil {
			return nil, -1, er1
		}
		return BuildSearchResult(ctx, models, total, mapper)
	} else {
		queryPaging := BuildPagingQuery(query, pageIndex, pageSize, initPageSize, driverName)
		queryCount, paramsCount := BuildCountQuery(query, params)
		fieldsIndex, er12 := GetColumnIndexes(modelType, driverName)
		if er12 != nil {
			return nil, -1, er12
		}
		er1 := Query(db, models, fieldsIndex, queryPaging, params...)
		if er1 != nil {
			return nil, -1, er1
		}
		total, er2 := Count(db, queryCount, paramsCount...)
		if er2 != nil {
			total = 0
		}
		return BuildSearchResult(ctx, models, total, mapper)
	}
}
func BuildPagingQueryByDriver(sql string, pageIndex int64, pageSize int64, initPageSize int64, driver string) string {
	s2 := BuildPagingQuery(sql, pageIndex, pageSize, initPageSize, driver)
	if driver != DriverOracle {
		return s2
	} else {
		i := strings.Index(s2, "select")
		if i < 0 {
			i = strings.Index(s2, "SELECT")
		}
		if i >= 0 {
			return s2[0:i+7] + " count(*) over() as total, " + s2[i+7:]
		}
		return s2
	}
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
	i := strings.Index(sql, "select ")
	if i < 0 {
		return sql, params
	}
	j := strings.Index(sql, " from ")
	if j < 0 {
		return sql, params
	}
	k := strings.Index(sql, " order by ")
	h := strings.Index(sql, " distinct ")
	if h > 0 {
		sql3 := `select count(*) as total from (` + sql[i:] + `) as main`
		return sql3, params
	}
	if k > 0 {
		sql3 := `select count(*) as total ` + sql[j:k]
		return sql3, params
	} else {
		sql3 := `select count(*) as total ` + sql[j:]
		return sql3, params
	}
}

func BuildSearchResult(ctx context.Context, models interface{}, count int64, mapper Mapper) (interface{}, int64, error) {
	if mapper == nil {
		return models, count, nil
	}
	r2, er3 := mapper.DbToModels(ctx, models)
	if er3 != nil {
		return models, count, nil
	}
	return r2, count, er3
}

func GetDriverName(db *sql.DB) string {
	if db == nil {
		return DriverNotSupport
	}
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
		return DriverNotSupport
	}
}

func BuildParam(index int, driver string) string {
	switch driver {
	case DriverPostgres:
		return "$" + strconv.Itoa(index)
	case DriverOracle:
		return ":val" + strconv.Itoa(index)
	default:
		return "?"
	}
}

func BuildParametersFrom(i int, numCol int, driver string) string {
	var arrValue []string
	for j := 0; j < numCol; j++ {
		arrValue = append(arrValue, BuildParam(i+j+1, driver))
	}
	return strings.Join(arrValue, ",")
}
