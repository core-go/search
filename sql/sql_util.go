package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
)

type Executor interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
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
func Count(ctx context.Context, db Executor, sql string, values ...interface{}) (int64, error) {
	var total int64
	row := db.QueryRowContext(ctx, sql, values...)
	err2 := row.Scan(&total)
	if err2 != nil {
		return total, err2
	}
	return total, nil
}
func QueryWithArray(ctx context.Context, db Executor, fieldsIndex map[string]int, results interface{}, toArray func(interface{}) interface {
	driver.Valuer
	sql.Scanner
}, sql string, values ...interface{}) error {
	rows, er1 := db.QueryContext(ctx, sql, values...)
	if er1 != nil {
		return er1
	}
	defer rows.Close()
	modelType := reflect.TypeOf(results).Elem().Elem()
	tb, er3 := Scan(rows, modelType, fieldsIndex, toArray)
	if er3 != nil {
		return er3
	}
	for _, element := range tb {
		appendToArray(results, element)
	}
	er4 := rows.Close()
	if er4 != nil {
		return er4
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if er5 := rows.Err(); er5 != nil {
		return er5
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

func QueryAndCount(ctx context.Context, db Executor, fieldsIndex map[string]int, results interface{}, toArray func(interface{}) interface {
	driver.Valuer
	sql.Scanner
}, count *int64, sql string, values ...interface{}) error {
	rows, er1 := db.QueryContext(ctx, sql, values...)
	if er1 != nil {
		return er1
	}
	defer rows.Close()
	modelType := reflect.TypeOf(results).Elem().Elem()

	if fieldsIndex == nil {
		fieldsIndex, er1 = GetColumnIndexes(modelType)
		if er1 != nil {
			return er1
		}
	}

	tb, c, er3 := ScanAndCount(rows, modelType, fieldsIndex, toArray)
	*count = c
	if er3 != nil {
		return er3
	}
	for _, element := range tb {
		appendToArray(results, element)
	}
	er4 := rows.Close()
	if er4 != nil {
		return er4
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if er5 := rows.Err(); er5 != nil {
		return er5
	}
	return nil
}

func GetColumnIndexes(modelType reflect.Type) (map[string]int, error) {
	ma := make(map[string]int)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return ma, errors.New("bad type")
	}
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		ormTag := field.Tag.Get("gorm")
		if ormTag != "-" {
			column, ok := FindTag(ormTag, "column")
			column = strings.ToLower(column)
			if ok {
				ma[column] = i
			}
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
func GetColumnsSelect(modelType reflect.Type) []string {
	numField := modelType.NumField()
	columnNameKeys := make([]string, 0)
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		ormTag := field.Tag.Get("gorm")
		if has := strings.Contains(ormTag, "column"); has {
			str1 := strings.Split(ormTag, ";")
			num := len(str1)
			for i := 0; i < num; i++ {
				str2 := strings.Split(str1[i], ":")
				for j := 0; j < len(str2); j++ {
					if str2[j] == "column" {
						columnName := strings.ToLower(str2[j+1])
						columnNameKeys = append(columnNameKeys, columnName)
					}
				}
			}
		}
	}
	return columnNameKeys
}
func GetColumns(cols []string, err error) ([]string, error) {
	if cols == nil || err != nil {
		return cols, err
	}
	c2 := make([]string, 0)
	for _, c := range cols {
		s := strings.ToLower(c)
		c2 = append(c2, s)
	}
	return c2, nil
}
func Scan(rows *sql.Rows, modelType reflect.Type, fieldsIndex map[string]int, options ...func(interface{}) interface {
	driver.Valuer
	sql.Scanner
}) (t []interface{}, err error) {
	if fieldsIndex == nil {
		fieldsIndex, err = GetColumnIndexes(modelType)
		if err != nil {
			return
		}
	}
	var toArray func(interface{}) interface {
		driver.Valuer
		sql.Scanner
	}
	if len(options) > 0 {
		toArray = options[0]
	}
	columns, er0 := GetColumns(rows.Columns())
	if er0 != nil {
		return nil, er0
	}
	for rows.Next() {
		initModel := reflect.New(modelType).Interface()
		r, swapValues := StructScan(initModel, columns, fieldsIndex, toArray)
		if err = rows.Scan(r...); err == nil {
			SwapValuesToBool(initModel, &swapValues)
			t = append(t, initModel)
		}
	}
	return
}
func StructScan(s interface{}, columns []string, fieldsIndex map[string]int, options ...func(interface{}) interface {
	driver.Valuer
	sql.Scanner
}) (r []interface{}, swapValues map[int]interface{}) {
	var toArray func(interface{}) interface {
		driver.Valuer
		sql.Scanner
	}
	if len(options) > 0 {
		toArray = options[0]
	}
	return StructScanAndIgnore(s, columns, fieldsIndex, toArray, -1)
}
func StructScanAndIgnore(s interface{}, columns []string, fieldsIndex map[string]int, toArray func(interface{}) interface {
	driver.Valuer
	sql.Scanner
}, indexIgnore int) (r []interface{}, swapValues map[int]interface{}) {
	if s != nil {
		modelType := reflect.TypeOf(s).Elem()
		swapValues = make(map[int]interface{}, 0)
		maps := reflect.Indirect(reflect.ValueOf(s))

		if columns == nil {
			for i := 0; i < maps.NumField(); i++ {
				tagBool := modelType.Field(i).Tag.Get("true")
				if tagBool == "" {
					r = append(r, maps.Field(i).Addr().Interface())
				} else {
					var str string
					swapValues[i] = reflect.New(reflect.TypeOf(str)).Elem().Addr().Interface()
					r = append(r, swapValues[i])
				}
			}
			return
		}

		for i, columnsName := range columns {
			if i == indexIgnore {
				continue
			}
			var index int
			var ok bool
			var modelField reflect.StructField
			var valueField reflect.Value
			if fieldsIndex == nil {
				if modelField, ok = modelType.FieldByName(columnsName); !ok {
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
				modelField = modelType.Field(index)
				valueField = maps.Field(index)
			}
			x := valueField.Addr().Interface()
			tagBool := modelField.Tag.Get("true")
			if tagBool == "" {
				if toArray != nil && valueField.Kind() == reflect.Slice {
					x = toArray(x)
				}
				r = append(r, x)
			} else {
				var str string
				y := reflect.New(reflect.TypeOf(str))
				swapValues[index] = y.Elem().Addr().Interface()
				r = append(r, swapValues[index])
			}
		}
	}
	return
}
func SwapValuesToBool(s interface{}, swap *map[int]interface{}) {
	if s != nil {
		modelType := reflect.TypeOf(s).Elem()
		maps := reflect.Indirect(reflect.ValueOf(s))
		for index, element := range *swap {
			dbValue2, ok2 := element.(*bool)
			if ok2 {
				if maps.Field(index).Kind() == reflect.Ptr {
					maps.Field(index).Set(reflect.ValueOf(dbValue2))
				} else {
					maps.Field(index).SetBool(*dbValue2)
				}
			} else {
				dbValue, ok := element.(*string)
				if ok {
					var isBool bool
					if *dbValue == "true" {
						isBool = true
					} else if *dbValue == "false" {
						isBool = false
					} else {
						boolStr := modelType.Field(index).Tag.Get("true")
						isBool = *dbValue == boolStr
					}
					if maps.Field(index).Kind() == reflect.Ptr {
						maps.Field(index).Set(reflect.ValueOf(&isBool))
					} else {
						maps.Field(index).SetBool(isBool)
					}
				}
			}
		}
	}
}
func ScanAndCount(rows *sql.Rows, modelType reflect.Type, fieldsIndex map[string]int, toArray func(interface{}) interface {
	driver.Valuer
	sql.Scanner
}) ([]interface{}, int64, error) {
	var t []interface{}
	columns, er0 := GetColumns(rows.Columns())
	if er0 != nil {
		return nil, 0, er0
	}
	if fieldsIndex == nil {
		fieldsIndex, er0 = GetColumnIndexes(modelType)
		if er0 != nil {
			return nil, 0, er0
		}
	}
	var count int64
	for rows.Next() {
		initModel := reflect.New(modelType).Interface()
		var c []interface{}
		c = append(c, &count)
		r, swapValues := StructScanAndIgnore(initModel, columns, fieldsIndex, toArray, 0)
		c = append(c, r...)
		if err := rows.Scan(c...); err == nil {
			SwapValuesToBool(initModel, &swapValues)
			t = append(t, initModel)
		}
	}
	return t, count, nil
}
