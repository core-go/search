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

type SearchBuilder struct {
	Database      *sql.DB
	BuildQuery    func(sm interface{}) (string, []interface{})
	ModelType     reflect.Type
	extractSearch func(m interface{}) (int64, int64, int64, error)
	Map           func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewSearchBuilder(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), options ...func(context.Context, interface{}) (interface{}, error)) *SearchBuilder {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	builder := &SearchBuilder{Database: db, BuildQuery: buildQuery, ModelType: modelType, Map: mp, extractSearch: ExtractSearch}
	return builder
}
func NewSearchBuilderWithMap(db *sql.DB, modelType reflect.Type, buildQuery func(sm interface{}) (string, []interface{}), mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *SearchBuilder {
	var extractSearch func(m interface{}) (int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearch = options[0]
	}
	builder := &SearchBuilder{Database: db, BuildQuery: buildQuery, ModelType: modelType, extractSearch: extractSearch, Map: mp}
	return builder
}
func NewDefaultSearchBuilder(db *sql.DB, tableName string, modelType reflect.Type, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (int64, int64, int64, error)) *SearchBuilder {
	var extractSearch func(m interface{}) (int64, int64, int64, error)
	if len(options) >= 1 {
		extractSearch = options[0]
	}
	driverName := GetDriver(db)
	queryBuilder := NewDefaultQueryBuilder(tableName, modelType, driverName)
	return NewSearchBuilderWithMap(db, modelType, queryBuilder.BuildQuery, mp, extractSearch)
}
func (b *SearchBuilder) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	sql, params := b.BuildQuery(m)
	pageIndex, pageSize, firstPageSize, err := b.extractSearch(m)
	if err != nil {
		return nil, 0, err
	}
	return BuildFromQuery(ctx, b.Database, b.ModelType, sql, params, pageIndex, pageSize, firstPageSize, b.Map)
}

func BuildFromQuery(ctx context.Context, db *sql.DB, modelType reflect.Type, query string, params []interface{}, pageIndex int64, pageSize int64, initPageSize int64, mp func(context.Context, interface{}) (interface{}, error)) (interface{}, int64, error) {
	var total int64
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	models := reflect.New(modelsType).Interface()
	driverName := GetDriver(db)
	if pageSize <= 0 {
		fieldsIndex, er12 := GetColumnIndexes(modelType, driverName)
		if er12 != nil {
			return nil, -1, er12
		}
		er1 := Query(db, models, fieldsIndex, query, params...)
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
		if driverName == DriverOracle {
			queryPaging := BuildPagingQueryByDriver(query, pageIndex, pageSize, initPageSize, driverName)
			er1 := QueryAndCount(db, models, &total, driverName, queryPaging, params...)
			if er1 != nil {
				return nil, -1, er1
			}
			return BuildSearchResult(ctx, models, total, mp)
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
			return BuildSearchResult(ctx, models, total, mp)
		}
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

func BuildSearchResult(ctx context.Context, models interface{}, count int64, mp func(context.Context, interface{}) (interface{}, error)) (interface{}, int64, error) {
	if mp == nil {
		return models, count, nil
	}
	r2, er3 := dbToModels(ctx, models, mp)
	return r2, count, er3
}

func GetDriver(db *sql.DB) string {
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
			} else  {
				y := x.Interface()
				mp(ctx, y)
			}

		}
	}
	return models, nil
}
