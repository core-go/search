package cassandra

import (
	"errors"
	"github.com/gocql/gocql"
	"reflect"
	"strings"
)

func ScanIter(iter *gocql.Iter, results interface{}, options...map[string]int) error {
	modelType := reflect.TypeOf(results).Elem().Elem()

	tb, er2 := Scan(iter, modelType, options...)
	if er2 != nil {
		return er2
	}
	for _, element := range tb {
		appendToArray(results, element)
	}
	return nil
}
func appendToArray(arr interface{}, item interface{}) interface{} {
	arrValue := reflect.ValueOf(arr)
	elemValue := reflect.Indirect(arrValue)

	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemValue = reflect.Indirect(itemValue)
	}
	elemValue.Set(reflect.Append(elemValue, itemValue))
	return arr
}
func GetColumnIndexes(modelType reflect.Type) (map[string]int, error) {
	ma := make(map[string]int, 0)
	if modelType.Kind() != reflect.Struct {
		return ma, errors.New("bad type")
	}
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		ormTag := field.Tag.Get("gorm")
		column, ok := FindTag(ormTag, "column")
		column = strings.ToLower(column)
		if ok {
			ma[column] = i
		}
	}
	return ma, nil
}
func FindTag(tag string, key string) (string, bool) {
	if has := strings.Contains(tag, key); has {
		str1 := strings.Split(tag, ";")
		num := len(str1)
		for i := 0; i < num; i++ {
			str2 := strings.Split(str1[i], ":")
			for j := 0; j < len(str2); j++ {
				if str2[j] == key {
					return str2[j+1], true
				}
			}
		}
	}
	return "", false
}
func Scan(iter *gocql.Iter, modelType reflect.Type, options...map[string]int) (t []interface{}, err error) {
	var fieldsIndex map[string]int
	if len(options) > 0 && options[0] != nil {
		fieldsIndex = options[0]
	} else {
		fieldsIndex, err = GetColumnIndexes(modelType)
	}
	if err != nil {
		return
	}
	columns := GetColumns(iter.Columns())
	for {
		initModel := reflect.New(modelType).Interface()
		r := StructScan(initModel, columns, fieldsIndex, -1)
		if !iter.Scan(r...) {
			return
		} else {
			t = append(t, initModel)
		}
	}
}
func StructScan(s interface{}, columns []string, fieldsIndex map[string]int, indexIgnore int) (r []interface{}) {
	if s != nil {
		modelType := reflect.TypeOf(s).Elem()
		maps := reflect.Indirect(reflect.ValueOf(s))
		if columns == nil {
			for i := 0; i < maps.NumField(); i++ {
				r = append(r, maps.Field(i).Addr().Interface())
			}
			return
		}
		for i, columnsName := range columns {
			if i == indexIgnore {
				continue
			}
			var index int
			var ok bool
			var valueField reflect.Value
			if fieldsIndex == nil {
				if _, ok = modelType.FieldByName(columnsName); !ok {
					var t interface{}
					r = append(r, &t)
					continue
				}
				valueField = maps.FieldByName(columnsName)
			} else {
				if index, ok = fieldsIndex[columnsName]; !ok {
					var t interface{}
					r = append(r, &t)
					continue
				}
				valueField = maps.Field(index)
			}
			x := valueField.Addr().Interface()
			r = append(r, x)
		}
	}
	return
}
func GetColumns(cols []gocql.ColumnInfo) []string {
	c2 := make([]string, 0)
	if cols == nil {
		return c2
	}
	for _, c := range cols {
		s := strings.ToLower(c.Name)
		c2 = append(c2, s)
	}
	return c2
}
