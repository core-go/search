package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

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
