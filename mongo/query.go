package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"strings"
)

func BuildSearchResult(ctx context.Context, collection *mongo.Collection, results interface{}, query bson.D, fields bson.M, sort bson.D, limit int64, skip int64) (int64, error) {
	optionsFind := options.Find()
	if fields != nil {
		optionsFind.Projection = fields
	}
	if skip > 0 {
		optionsFind.SetSkip(skip)
	}
	if limit > 0 {
		optionsFind.SetLimit(limit)
	}
	if sort != nil {
		optionsFind.SetSort(sort)
	}

	cursor, er0 := collection.Find(ctx, query, optionsFind)
	if er0 != nil {
		return 0, er0
	}

	er1 := cursor.All(ctx, results)
	if er1 != nil {
		return 0, er1
	}
	options := options.Count()
	return collection.CountDocuments(ctx, query, options)
}

func BuildSort(s string, modelType reflect.Type) bson.D {
	var sort = bson.D{}
	if len(s) == 0 {
		return sort
	}
	sorts := strings.Split(s, ",")
	for i := 0; i < len(sorts); i++ {
		sortField := strings.TrimSpace(sorts[i])
		if len(sortField) > 0 {
			fieldName := sortField
			c := sortField[0:1]
			if c == "-" || c == "+" {
				fieldName = sortField[1:]
			}
			columnName := GetBsonNameForSort(modelType, fieldName)
			if len(columnName) > 0 {
				sortType := GetSortType(c)
				sort = append(sort, bson.E{Key: columnName, Value: sortType})
			}
		}
	}
	return sort
}

func GetFieldByJson(modelType reflect.Type, jsonName string) (int, string, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		tag1, ok1 := field.Tag.Lookup("json")
		if ok1 && strings.Split(tag1, ",")[0] == jsonName {
			if tag2, ok2 := field.Tag.Lookup("bson"); ok2 {
				return i, field.Name, strings.Split(tag2, ",")[0]
			}
			return i, field.Name, ""
		}
	}
	return -1, jsonName, jsonName
}

func GetBsonNameForSort(modelType reflect.Type, sortField string) string {
	sortField = strings.TrimSpace(sortField)
	idx, fieldName, name := GetFieldByJson(modelType, sortField)
	if len(name) > 0 {
		return name
	}
	if idx >= 0 {
		return fieldName
	}
	return sortField
}

func GetSortType(sortType string) int {
	if sortType == "-" {
		return -1
	} else {
		return 1
	}
}

func GetFields(fields []string, modelType reflect.Type) bson.M {
	if len(fields) <= 0 {
		return nil
	}
	ex := false
	var fs = bson.M{}
	for _, key := range fields {
		_, _, columnName := GetFieldByJson(modelType, key)
		if len(columnName) >= 0 {
			fs[columnName] = 1
			ex = true
		}
	}
	if ex == false {
		return nil
	}
	return fs
}
