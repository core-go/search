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

type SearchBuilder struct {
	DB          *gocql.ClusterConfig
	BuildQuery  func(sm interface{}) (string, []interface{})
	ModelType   reflect.Type
	Map         func(ctx context.Context, model interface{}) (interface{}, error)
	fieldsIndex map[string]int
}
func NewSearchQuery(db *gocql.ClusterConfig, modelType reflect.Type, buildQuery func(interface{}) (string, []interface{}), options ...func(context.Context, interface{}) (interface{}, error)) (*SearchBuilder, error) {
	return NewSearchBuilder(db, modelType, buildQuery, options...)
}
func NewSearchBuilder(db *gocql.ClusterConfig, modelType reflect.Type, buildQuery func(interface{}) (string, []interface{}), options ...func(context.Context, interface{}) (interface{}, error)) (*SearchBuilder, error) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	fieldsIndex, err := GetColumnIndexes(modelType)
	if err != nil {
		return nil, err
	}
	builder := &SearchBuilder{DB: db, fieldsIndex: fieldsIndex, BuildQuery: buildQuery, ModelType: modelType, Map: mp}
	return builder, nil
}

func (b *SearchBuilder) Search(ctx context.Context, m interface{}, results interface{}, limit int64, refId string) (string, error) {
	sql, params := b.BuildQuery(m)
	ses, err := b.DB.CreateSession()
	defer ses.Close()

	if err != nil {
		return "", err
	}
	nextPageToken, er2 := QueryWithMap(ses, b.fieldsIndex, results, sql, params, limit, refId, b.Map)
	return nextPageToken, er2
}
func QueryWithMap(ses *gocql.Session, fieldsIndex map[string]int, results interface{}, sql string, values []interface{}, max int64, refId string, options...func(context.Context, interface{}) (interface{}, error)) (string, error) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) > 0 && options[0] != nil {
		mp = options[0]
	}
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
	if mp != nil {
		_, err := MapModels(context.Background(), results, mp)
		return nextPageToken, err
	}
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
