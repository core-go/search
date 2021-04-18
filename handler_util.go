package search

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func buildResourceName(s string) string {
	s2 := strings.ToLower(s)
	s3 := ""
	for i := range s {
		if s2[i] != s[i] {
			s3 += "-" + string(s2[i])
		} else {
			s3 += string(s2[i])
		}
	}
	if string(s3[0]) == "-" || string(s3[0]) == "_" {
		return s3[1:]
	}
	return s3
}
func UrlToModel(searchModel interface{}, params url.Values, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) interface{} {
	value := reflect.Indirect(reflect.ValueOf(searchModel))
	if value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	for paramKey, valueArr := range params {
		paramValue := ""
		if len(valueArr) > 0 {
			paramValue = valueArr[0]
		}
		if err, field := FindField(value, paramKey, searchModelParamIndex, searchModelIndex, paramIndex); err == nil {
			kind := field.Kind()

			var v interface{}
			// Need handle more case of kind
			if kind == reflect.Int {
				v, _ = strconv.Atoi(paramValue)
			} else if kind == reflect.Int64 {
				v, _ = strconv.ParseInt(paramValue, 10, 64)
			} else if kind == reflect.String {
				v = paramValue
			} else if kind == reflect.Slice {
				sliceKind := reflect.TypeOf(field.Interface()).Elem().Kind()
				if sliceKind == reflect.String {
					v = strings.Split(paramValue, ",")
				} else {
					log.Println("Unhandled slice kind:", kind)
					continue
				}
			} else if kind == reflect.Struct {
				newModel := reflect.New(reflect.Indirect(field).Type()).Interface()
				if errDecode := json.Unmarshal([]byte(paramValue), newModel); errDecode != nil {
					panic(errDecode)
				}
				v = newModel
			} else {
				log.Println("Unhandled kind:", kind)
				continue
			}
			field.Set(reflect.Indirect(reflect.ValueOf(v)))
		} else {
			log.Println(err)
		}
	}
	return searchModel
}
func FindField(value reflect.Value, paramKey string, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) (error, reflect.Value) {
	if index, ok := searchModelParamIndex[paramKey]; ok {
		searchModelField := value.Field(searchModelIndex)
		if searchModelField.Kind() == reflect.Ptr {
			searchModelField = reflect.Indirect(searchModelField)
		}
		return nil, searchModelField.Field(index)
	} else if index, ok := paramIndex[paramKey]; ok {
		return nil, value.Field(index)
	}
	return errors.New("can't find field " + paramKey), value
}
func BuildParamIndex(searchModelType reflect.Type) map[string]int {
	params := map[string]int{}
	numField := searchModelType.NumField()
	for i := 0; i < numField; i++ {
		field := searchModelType.Field(i)
		fullJsonTag := field.Tag.Get("json")
		tagDetails := strings.Split(fullJsonTag, ",")
		if len(tagDetails) > 0 && len(tagDetails[0]) > 0 {
			params[tagDetails[0]] = i
		}
	}
	return params
}

func BuildSearchModel(r *http.Request, searchModelType reflect.Type, isExtendedSearchModelType bool, userIdName string, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) (interface{}, int, error) {
	var searchModel = CreateSearchModel(searchModelType, isExtendedSearchModelType)
	method := r.Method
	x := 1
	if method == http.MethodGet {
		ps := r.URL.Query()
		fs := ps.Get("fields")
		if len(fs) == 0 {
			x = -1
		}
		UrlToModel(searchModel, ps, searchModelParamIndex, searchModelIndex, paramIndex)
	} else if method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&searchModel); err != nil {
			return nil, x, err
		}
	}
	userId := ""
	if len(userId) == 0 {
		u := r.Context().Value(userIdName)
		if u != nil {
			u2, ok2 := u.(string)
			if ok2 {
				userId = u2
			}
		}
	}
	SetUserId(searchModel, userId)
	return searchModel, x, nil
}
func BuildResultMap(models interface{}, count int64, pageIndex int64, pageSize int64, firstPageSize int64, config SearchResultConfig) (map[string]interface{}, bool) {
	result := make(map[string]interface{})
	isLastPage := IsLastPage(models, count, pageIndex, pageSize, firstPageSize)

	result[config.Total] = count
	if isLastPage {
		result[config.LastPage] = isLastPage
	}
	result[config.Results] = models
	return result, isLastPage
}
func ResultToCsv(fields []string, models interface{}, count int64, isLastPage bool, embedField string) (string, bool) {
	if len(fields) > 0 {
		result1 := ToCsv(fields, models, count, isLastPage, embedField)
		return result1, true
	} else {
		return "", false
	}
}
