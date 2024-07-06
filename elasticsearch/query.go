package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

func BuildSearchResult(ctx context.Context, db *elasticsearch.Client, index []string, results interface{}, jsonName string, query map[string]interface{}, sort []map[string]interface{}, limit int64, offset int64, version string) (int64, error) {
	from := int(offset)
	size := int(limit)
	fullQuery := UpdateQuery(query)
	fullQuery["sort"] = sort
	req := esapi.SearchRequest{
		// Index: []string{indexName},
		Index: index,
		Body:  esutil.NewJSONReader(fullQuery),
		From:  &from,
		Size:  &size,
	}

	res, err := req.Do(ctx, db)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	var count int64
	if res.IsError() {
		return 0, errors.New("response error")
	} else {
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			return 0, err
		} else {
			hitsObj := r["hits"].(map[string]interface{})
			hits := hitsObj["hits"].([]interface{})
			count = int64(hitsObj["total"].(map[string]interface{})["value"].(float64))
			listResults := make([]interface{}, 0)
			for _, hit := range hits {
				hitObj := hit.(map[string]interface{})
				r := hitObj["_source"]
				rs := r.(map[string]interface{})
				if len(jsonName) > 0 {
					rs[jsonName] = hitObj["_id"]
				}
				if len(version) > 0 {
					rs[version] = hitObj["_version"]
				}
				listResults = append(listResults, r)
			}

			err := json.NewDecoder(esutil.NewJSONReader(listResults)).Decode(results)
			if err != nil {
				return count, err
			}
			return count, err
		}
	}
}
func UpdateQuery(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	result["query"] = map[string]interface{}{
		"bool": map[string]interface{}{
			"must": make([]map[string]interface{}, 0),
		},
	}
	queryFields := make([]map[string]interface{}, 0)
	for key, value := range m {
		q := make(map[string]interface{})
		if reflect.ValueOf(value).Kind() == reflect.Map {
			q["range"] = make(map[string]interface{})
			q["range"].(map[string]interface{})[key] = make(map[string]interface{})
			for operator, val := range value.(map[string]interface{}) {
				q["range"].(map[string]interface{})[key].(map[string]interface{})[operator[1:]] = val
			}
		} else {
			q["prefix"] = make(map[string]interface{})
			q["prefix"].(map[string]interface{})[key] = value
		}
		queryFields = append(queryFields, q)
	}
	result["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = queryFields
	return result
}
func BuildSort(s string, modelType reflect.Type) []map[string]interface{} {
	sort := []map[string]interface{}{}
	if len(s) == 0 {
		return sort
	}
	sorts := strings.Split(s, ",")
	for i := 0; i < len(sorts); i++ {
		sortField := strings.TrimSpace(sorts[i])
		fieldName := sortField

		var mapFieldName map[string]interface{}
		c := sortField[0:1]
		if c == "-" || c == "+" {
			//fieldName = sortField[1:]
			field, ok := getFieldName(modelType, sortField[1:])
			if !ok {
				return []map[string]interface{}{}
			}
			fieldName = field
			if c == "-" {
				mapFieldName = map[string]interface{}{
					fieldName: map[string]string{
						"order": "desc",
					},
				}
			} else {
				mapFieldName = map[string]interface{}{
					fieldName: map[string]string{
						"order": "asc",
					},
				}
			}
		}
		sort = append(sort, mapFieldName)
	}

	return sort
}

func getFieldName(structType reflect.Type, jsonTagValue string) (string, bool) {
	var (
		bsonTagValue string
		typeField    reflect.Kind
	)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == jsonTagValue {
			bsonTagValue = field.Tag.Get("bson")
			typeField = field.Type.Kind()
			break
		}
	}
	if bsonTagValue != "_id" {
		if typeField == reflect.String {
			return "", false
		}
		return jsonTagValue, true
	}
	return bsonTagValue, true
}
