package cassandra

import (
	"context"
	"encoding/hex"
	"reflect"
	"strings"

	"github.com/gocql/gocql"
)

const (
	desc = "desc"
	asc  = "asc"
)

type SearchBuilder[T any, K any, F any] struct {
	DB         *gocql.ClusterConfig
	Table      string
	BuildQuery func(F) (string, []interface{})
	Mp         func(*T)
	Map        map[string]int
}

func NewSearchBuilder[T any, K any, F any](db *gocql.ClusterConfig, table string, buildQuery func(F) (string, []interface{}), opts ...func(*T)) (*SearchBuilder[T, K, F], error) {
	var mp func(*T)
	if len(opts) >= 1 {
		mp = opts[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	fieldsIndex, err := GetColumnIndexes(modelType)
	if err != nil {
		return nil, err
	}
	builder := &SearchBuilder[T, K, F]{DB: db, Table: table, Map: fieldsIndex, BuildQuery: buildQuery, Mp: mp}
	return builder, nil
}

func (b *SearchBuilder[T, K, F]) Search(ctx context.Context, filter F, limit int64, next string) ([]T, string, error) {
	var objs []T
	sql, params := b.BuildQuery(filter)
	ses, err := b.DB.CreateSession()
	defer ses.Close()

	if err != nil {
		return objs, "", err
	}
	nextPageToken, er2 := QueryWithMap(ses, b.Map, &objs, sql, params, limit, next)
	if b.Mp != nil {
		l := len(objs)
		for i := 0; i < l; i++ {
			b.Mp(&objs[i])
		}
	}
	return objs, nextPageToken, er2
}

func QueryWithMap(ses *gocql.Session, fieldsIndex map[string]int, results interface{}, sql string, values []interface{}, max int64, refId string) (string, error) {
	next, er0 := hex.DecodeString(refId)
	if er0 != nil {
		return "", er0
	}
	query := ses.Query(sql, values...).PageState(next).PageSize(int(max))
	if query.Exec() != nil {
		return "", query.Exec()
	}
	err := ScanIter(query.Iter(), results, fieldsIndex)
	if err != nil {
		return "", err
	}
	nextPageToken := hex.EncodeToString(query.Iter().PageState())
	return nextPageToken, nil
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
