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
	DriverSqlite3       = "sqlite3"
	DriverNotSupport    = "no support"
	DefaultPagingFormat = " limit %s offset %s "
	OraclePagingFormat  = " offset %s rows fetch next %s rows only "
	desc                = "desc"
	asc                 = "asc"
)

type SearchBuilder struct {
	Database   *sql.DB
	BuildQuery func(sm interface{}) (string, []interface{})
	ModelType  reflect.Type
	Extract    func(m interface{}) (int64, int64, int64, error)
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewSearchBuilder(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), options ...func(context.Context, interface{}) (interface{}, error)) *SearchBuilder {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	builder := &SearchBuilder{Database: db, BuildQuery: buildQuery, ModelType: modelType, Map: mp, Extract: ExtractSearch}
	return builder
}
func NewSearchBuilderWithMap(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *SearchBuilder {
	var extractSearch func(m interface{}) (int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearch = options[0]
	} else {
		extractSearch = ExtractSearch
	}
	builder := &SearchBuilder{Database: db, BuildQuery: buildQuery, ModelType: modelType, Extract: extractSearch, Map: mp}
	return builder
}
func NewDefaultSearchBuilder(db *sql.DB, tableName string, modelType reflect.Type, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *SearchBuilder {
	driver := getDriver(db)
	buildParam := getBuild(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driver, buildParam)
	return NewSearchBuilderWithMap(db, modelType, queryBuilder.BuildQuery, mp, options...)
}
func (b *SearchBuilder) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	sql, params := b.BuildQuery(m)
	pageIndex, pageSize, firstPageSize, err := b.Extract(m)
	if err != nil {
		return nil, 0, err
	}
	return BuildFromQuery(ctx, b.Database, b.ModelType, sql, params, pageIndex, pageSize, firstPageSize, b.Map)
}

func BuildFromQuery(ctx context.Context, db *sql.DB, modelType reflect.Type, query string, params []interface{}, pageIndex int64, pageSize int64, initPageSize int64, mp func(context.Context, interface{}) (interface{}, error)) (interface{}, int64, error) {
	var total int64
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	models := reflect.New(modelsType).Interface()
	driver := getDriver(db)
	if pageSize <= 0 {
		fieldsIndex, er12 := getColumnIndexes(modelType, driver)
		if er12 != nil {
			return nil, -1, er12
		}
		er1 := Query(ctx, db, models, fieldsIndex, query, params...)
		if er1 != nil {
			return nil, -1, er1
		}
		objectValues := reflect.Indirect(reflect.ValueOf(models))
		if objectValues.Kind() == reflect.Slice {
			i := objectValues.Len()
			total = int64(i)
		}
		return BuildSearchResult(ctx, models, total, mp)
	} else {
		if driver == DriverOracle {
			queryPaging := BuildPagingQueryByDriver(query, pageIndex, pageSize, initPageSize, driver)
			er1 := QueryAndCount(ctx, db, models, &total, driver, queryPaging, params...)
			if er1 != nil {
				return nil, -1, er1
			}
			return BuildSearchResult(ctx, models, total, mp)
		} else {
			queryPaging := BuildPagingQuery(query, pageIndex, pageSize, initPageSize, driver)
			queryCount, paramsCount := BuildCountQuery(query, params)
			fieldsIndex, er12 := getColumnIndexes(modelType, driver)
			if er12 != nil {
				return nil, -1, er12
			}
			er1 := Query(ctx, db, models, fieldsIndex, queryPaging, params...)
			if er1 != nil {
				return nil, -1, er1
			}
			total, er2 := Count(ctx, db, queryCount, paramsCount...)
			if er2 != nil {
				total = 0
			}
			return BuildSearchResult(ctx, models, total, mp)
		}
	}
}
func BuildPagingQueryByDriver(sql string, pageIndex int64, pageSize int64, initPageSize int64, driver string) string {
	s2 := BuildPagingQuery(sql, pageIndex, pageSize, initPageSize, driver)
	if driver != DriverOracle {
		return s2
	} else {
		l := len(" distinct ")
		i := strings.Index(sql, " distinct ")
		if i < 0 {
			i = strings.Index(sql, " DISTINCT ")
		}
		if i < 0 {
			l = len("select") + 1
			i = strings.Index(s2, "select")
		}
		if i < 0 {
			i = strings.Index(s2, "SELECT")
		}
		if i >= 0 {
			return s2[0:l] + " count(*) over() as total, " + s2[l:]
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

func BuildSearchResult(ctx context.Context, models interface{}, count int64, mp func(context.Context, interface{}) (interface{}, error)) (interface{}, int64, error) {
	if mp == nil {
		return models, count, nil
	}
	r2, er3 := dbToModels(ctx, models, mp)
	return r2, count, er3
}

func getDriver(db *sql.DB) string {
	if db == nil {
		return DriverNotSupport
	}
	driver := reflect.TypeOf(db.Driver()).String()
	switch driver {
	case "*pq.Driver":
		return DriverPostgres
	case "*godror.drv":
		return DriverOracle
	case "*mysql.MySQLDriver":
		return DriverMysql
	case "*mssql.Driver":
		return DriverMssql
	case "*sqlite3.SQLiteDriver":
		return DriverSqlite3
	default:
		return DriverNotSupport
	}
}
func buildParam(i int) string {
	return "?"
}
func buildOracleParam(i int) string {
	return ":val" + strconv.Itoa(i)
}
func buildMsSqlParam(i int) string {
	return "@p" + strconv.Itoa(i)
}
func buildDollarParam(i int) string {
	return "$" + strconv.Itoa(i)
}
func getBuild(db *sql.DB) func(i int) string {
	driver := reflect.TypeOf(db.Driver()).String()
	switch driver {
	case "*pq.Driver":
		return buildDollarParam
	case "*godror.drv":
		return buildOracleParam
	case "*mysql.MySQLDriver":
		return buildMsSqlParam
	default:
		return buildParam
	}
}
func BuildParametersFrom(i int, numCol int, buildParam func(i int) string) string {
	var arrValue []string
	for j := 0; j < numCol; j++ {
		arrValue = append(arrValue, buildParam(i+j+1))
	}
	return strings.Join(arrValue, ",")
}

func dbToModels(ctx context.Context, models interface{}, mp func(context.Context, interface{}) (interface{}, error)) (interface{}, error) {
	valueModelObject := reflect.Indirect(reflect.ValueOf(models))
	if valueModelObject.Kind() == reflect.Ptr {
		valueModelObject = reflect.Indirect(valueModelObject)
	}
	if valueModelObject.Kind() == reflect.Slice {
		le := valueModelObject.Len()
		for i := 0; i < le; i++ {
			x := valueModelObject.Index(i)
			k := x.Kind()
			if k == reflect.Struct {
				y := x.Addr().Interface()
				mp(ctx, y)
			} else {
				y := x.Interface()
				mp(ctx, y)
			}

		}
	}
	return models, nil
}
