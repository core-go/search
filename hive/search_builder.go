package hive

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	hv "github.com/beltran/gohive"
)

const (
	desc                = "desc"
	asc                 = "asc"
	DefaultPagingFormat = " limit %s offset %s "
)

type SearchBuilder[T any, F any] struct {
	Connection *hv.Connection
	BuildQuery func(F) string
	Mp         func(*T)
	Map        map[string]int
}

func NewSearchBuilder[T any, F any](connection *hv.Connection, buildQuery func(F) string, options ...func(*T)) (*SearchBuilder[T, F], error) {
	var mp func(*T)
	if len(options) >= 1 {
		mp = options[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() != reflect.Struct {
		return nil, errors.New("T must be a struct")
	}
	fieldsIndex, err := GetColumnIndexes(modelType)
	if err != nil {
		return nil, err
	}
	builder := &SearchBuilder[T, F]{Connection: connection, Map: fieldsIndex, BuildQuery: buildQuery, Mp: mp}
	return builder, nil
}

func (b *SearchBuilder[T, F]) Search(ctx context.Context, m F, limit int64, offset int64) ([]T, int64, error) {
	sql := b.BuildQuery(m)
	query := BuildPagingQuery(sql, limit, offset)
	cursor := b.Connection.Cursor()
	defer cursor.Close()
	var res []T
	cursor.Exec(ctx, sql)
	if cursor.Err != nil {
		return res, -1, cursor.Err
	}
	err := Query(ctx, cursor, b.Map, &res, query)
	if err != nil {
		return res, -1, err
	}
	countQuery := BuildCountQuery(sql)
	cursor.Exec(ctx, countQuery)
	if cursor.Err != nil {
		return res, -1, cursor.Err
	}
	var count int64
	for cursor.HasMore(ctx) {
		cursor.FetchOne(ctx, &count)
		if cursor.Err != nil {
			return res, count, cursor.Err
		}
	}
	if b.Mp != nil {
		l := len(res)
		for i := 0; i < l; i++ {
			b.Mp(&res[i])
		}
	}
	return res, count, err
}
func Count(ctx context.Context, cursor *hv.Cursor, query string) (int64, error) {
	var count int64
	cursor.Exec(ctx, query)
	if cursor.Err != nil {
		return -1, cursor.Err
	}
	for cursor.HasMore(ctx) {
		cursor.FetchOne(ctx, &count)
		if cursor.Err != nil {
			return count, cursor.Err
		}
	}
	return 0, nil
}
func BuildPagingQuery(sql string, limit int64, offset int64) string {
	if offset < 0 {
		offset = 0
	}
	if limit > 0 {
		pagingQuery := fmt.Sprintf(DefaultPagingFormat, strconv.FormatInt(limit, 10), strconv.FormatInt(offset, 10))
		sql += pagingQuery
	}
	return sql
}
func BuildCountQuery(sql string) string {
	i := strings.Index(sql, "select ")
	if i < 0 {
		return sql
	}
	j := strings.Index(sql, " from ")
	if j < 0 {
		return sql
	}
	k := strings.Index(sql, " order by ")
	h := strings.Index(sql, " distinct ")
	if h > 0 {
		sql3 := `select count(*) as total from (` + sql[i:] + `) as main`
		return sql3
	}
	if k > 0 {
		sql3 := `select count(*) as total ` + sql[j:k]
		return sql3
	} else {
		sql3 := `select count(*) as total ` + sql[j:]
		return sql3
	}
}
func GetSort(sortString string, modelType reflect.Type) string {
	var sort = make([]string, 0)
	sorts := strings.Split(sortString, ",")
	for i := 0; i < len(sorts); i++ {
		sortField := strings.TrimSpace(sorts[i])
		fieldName := sortField
		c := sortField[0:1]
		if c == "-" || c == "+" {
			fieldName = sortField[1:]
		}
		columnName := GetColumnNameForSearch(modelType, fieldName)
		if len(columnName) > 0 {
			sortType := GetSortType(c)
			sort = append(sort, columnName+" "+sortType)
		}
	}
	if len(sort) > 0 {
		return strings.Join(sort, ",")
	} else {
		return ""
	}
}
func BuildSort(sortString string, modelType reflect.Type) string {
	sort := GetSort(sortString, modelType)
	if len(sort) > 0 {
		return ` order by ` + sort
	} else {
		return ""
	}
}
func GetColumnNameForSearch(modelType reflect.Type, sortField string) string {
	sortField = strings.TrimSpace(sortField)
	i, _, column := GetFieldByJson(modelType, sortField)
	if i > -1 {
		return column
	}
	return ""
}
func GetSortType(sortType string) string {
	if sortType == "-" {
		return desc
	} else {
		return asc
	}
}
func GetFieldByJson(modelType reflect.Type, jsonName string) (int, string, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		tag1, ok1 := field.Tag.Lookup("json")
		if ok1 && strings.Split(tag1, ",")[0] == jsonName {
			if tag2, ok2 := field.Tag.Lookup("gorm"); ok2 {
				if has := strings.Contains(tag2, "column"); has {
					str1 := strings.Split(tag2, ";")
					num := len(str1)
					for k := 0; k < num; k++ {
						str2 := strings.Split(str1[k], ":")
						for j := 0; j < len(str2); j++ {
							if str2[j] == "column" {
								return i, field.Name, str2[j+1]
							}
						}
					}
				}
			}
			return i, field.Name, ""
		}
	}
	return -1, jsonName, jsonName
}
