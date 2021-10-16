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

func BuildResourceName(s string) string {
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
func UrlToModel(filter interface{}, params url.Values, filterParamIndex map[string]int, filterIndex int, paramIndex map[string]int) interface{} {
	value := reflect.Indirect(reflect.ValueOf(filter))
	if value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	for paramKey, valueArr := range params {
		paramValue := ""
		if len(valueArr) > 0 {
			paramValue = valueArr[0]
		}
		if err, field := FindField(value, paramKey, filterParamIndex, filterIndex, paramIndex); err == nil {
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
	return filter
}
func FindField(value reflect.Value, paramKey string, filterParamIndex map[string]int, filterIndex int, paramIndex map[string]int) (error, reflect.Value) {
	if index, ok := filterParamIndex[paramKey]; ok {
		filterField := value.Field(filterIndex)
		if filterField.Kind() == reflect.Ptr {
			filterField = reflect.Indirect(filterField)
		}
		return nil, filterField.Field(index)
	} else if index, ok := paramIndex[paramKey]; ok {
		return nil, value.Field(index)
	}
	return errors.New("can't find field " + paramKey), value
}
func BuildParamIndex(filterType reflect.Type) map[string]int {
	params := map[string]int{}
	numField := filterType.NumField()
	for i := 0; i < numField; i++ {
		field := filterType.Field(i)
		fullJsonTag := field.Tag.Get("json")
		tagDetails := strings.Split(fullJsonTag, ",")
		if len(tagDetails) > 0 && len(tagDetails[0]) > 0 {
			params[tagDetails[0]] = i
		}
	}
	return params
}

func BuildFilter(r *http.Request, filterType reflect.Type, isExtendedFilter bool, userIdName string, filterParamIndex map[string]int, filterIndex int, paramIndex map[string]int) (interface{}, int, error) {
	var filter = CreateFilter(filterType, isExtendedFilter)
	method := r.Method
	x := 1
	if method == http.MethodGet {
		ps := r.URL.Query()
		fs := ps.Get("fields")
		if len(fs) == 0 {
			x = -1
		}
		UrlToModel(filter, ps, filterParamIndex, filterIndex, paramIndex)
	} else if method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
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
	SetUserId(filter, userId)
	return filter, x, nil
}
func BuildResultMap(models interface{}, count int64, nextPageToken string, config SearchResultConfig) map[string]interface{} {
	result := make(map[string]interface{})

	result[config.Total] = count
	result[config.Results] = models
	if len(nextPageToken) > 0 {
		result[config.NextPageToken] = nextPageToken
	}
	return result
}
func ResultToCsv(fields []string, models interface{}, count int64, nextPageToken string, embedField string) (string, bool) {
	if len(fields) > 0 {
		result1 := ToCsv(fields, models, count, nextPageToken, embedField)
		return result1, true
	} else {
		return "", false
	}
}
