package search

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
)

func getFieldByJson(modelType reflect.Type, jsonName string) (int, string, string) {
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
func getColumnName(modelType reflect.Type, fieldName string) (col string, colExist bool) {
	field, ok := modelType.FieldByName(fieldName)
	if !ok {
		return fieldName, false
	}
	tag2, ok2 := field.Tag.Lookup("gorm")
	if !ok2 {
		return "", true
	}

	if has := strings.Contains(tag2, "column"); has {
		str1 := strings.Split(tag2, ";")
		num := len(str1)
		for i := 0; i < num; i++ {
			str2 := strings.Split(str1[i], ":")
			for j := 0; j < len(str2); j++ {
				if str2[j] == "column" {
					return str2[j+1], true
				}
			}
		}
	}
	//return gorm.ToColumnName(fieldName), false
	return fieldName, false
}

func Count(ctx context.Context, db *sql.DB, sql string, values ...interface{}) (int64, error) {
	var total int64
	row := db.QueryRowContext(ctx, sql, values...)
	err2 := row.Scan(&total)
	if err2 != nil {
		return total, err2
	}
	return total, nil
}

func Query(ctx context.Context, db *sql.DB, results interface{}, fieldsIndex map[string]int, sql string, values ...interface{}) error {
	rows, er1 := db.QueryContext(ctx, sql, values...)
	if er1 != nil {
		return er1
	}

	defer rows.Close()
	modelType := reflect.TypeOf(results).Elem().Elem()

	if fieldsIndex == nil {
		tb, er2 := scanSearchType(rows, modelType)
		if er2 != nil {
			return er2
		}
		if tb != nil {
			reflect.ValueOf(results).Elem().Set(reflect.ValueOf(tb).Elem())
		}
	} else {
		tb, er4 := scanType(rows, modelType, fieldsIndex)
		if er4 != nil {
			return er4
		}
		for _, element := range tb {
			appendToArray(results, element)
		}
	}
	er5 := rows.Close()
	if er5 != nil {
		return er5
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if er6 := rows.Err(); er6 != nil {
		return er6
	}
	return nil
}

func QueryAndCount(ctx context.Context, db *sql.DB, results interface{}, count *int64, driver string, sql string, values ...interface{}) error {
	rows, er1 := db.QueryContext(ctx, sql, values...)
	if er1 != nil {
		return er1
	}
	defer rows.Close()
	modelType := reflect.TypeOf(results).Elem().Elem()

	fieldsIndex, er0 := getColumnIndexes(modelType, driver)
	if er0 != nil {
		return er0
	}

	tb, c, er2 := scansSearchAndCount(rows, modelType, fieldsIndex)
	*count = c
	if er2 != nil {
		return er2
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

func swapValuesToBool(s interface{}, swap *map[int]interface{})  {
	if s != nil {
		maps := reflect.Indirect(reflect.ValueOf(s))
		modelType := reflect.TypeOf(s).Elem()
		for index, element := range (*swap){
			var isBool bool
			boolStr := modelType.Field(index).Tag.Get("true")
			var dbValue = element.(*string)
			isBool = *dbValue == boolStr
			if maps.Field(index).Kind() == reflect.Ptr {
				maps.Field(index).Set(reflect.ValueOf(&isBool))
			} else {
				maps.Field(index).SetBool(isBool)
			}
		}
	}
}

func getColumnIndexes(modelType reflect.Type, driver string) (map[string]int, error) {
	mapp := make(map[string]int, 0)
	if modelType.Kind() != reflect.Struct {
		return mapp, errors.New("bad type")
	}
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		ormTag := field.Tag.Get("gorm")
		column, ok := findTag(ormTag, "column")
		if ok {
			if driver == DriverOracle {
				column = strings.ToUpper(column)
			}
			mapp[column] = i
		}
	}
	return mapp, nil
}

func findTag(tag string, key string) (string, bool) {
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

func scanType(rows *sql.Rows, modelType reflect.Type, fieldsIndex map[string]int) (t []interface{}, err error) {
	columns, er3 := rows.Columns()
	if er3 != nil {
		return nil, er3
	}
	for rows.Next() {
		initModel := reflect.New(modelType).Interface()
		r, swap := structScan(initModel, columns, fieldsIndex, -1)
		if err = rows.Scan(r...); err == nil {
			swapValuesToBool(initModel, &swap)
			t = append(t, initModel)
		}
	}
	return
}

func scanSearchType(rows *sql.Rows, modelType reflect.Type) (t []interface{}, err error) {
	for rows.Next() {
		gTb := reflect.New(modelType).Interface()
		r, swp := structScan(gTb, nil, nil,-1)
		if err = rows.Scan(r...); err == nil {
			swapValuesToBool(gTb, &swp)
			t = append(t, gTb)
		}
	}

	return
}

func scansSearchAndCount(rows *sql.Rows, modelType reflect.Type, fieldsIndex map[string]int) ([]interface{}, int64, error) {
	columns, er0 := rows.Columns()
	if er0 != nil {
		return nil, 0, er0
	}
	var t []interface{}
	var count int64
	for rows.Next() {
		initModel := reflect.New(modelType).Interface()
		var c []interface{}
		c = append(c, &count)
		r, swap := structScan(initModel, columns, fieldsIndex, 0)
		c = append(c, r...)
		if err := rows.Scan(c...); err == nil {
			swapValuesToBool(initModel, &swap)
			t = append(t, initModel)
		}
	}
	return t, count, nil
}

// StructScan : transfer struct to slice for scan
func structScan(s interface{}, columns []string, fieldsIndex map[string]int, indexIgnore int) (r []interface{}, swapValues map[int]interface{}) {
	if s != nil {
		maps := reflect.Indirect(reflect.ValueOf(s))
		swapValues = make(map[int]interface{}, 0)
		modelType := reflect.TypeOf(s).Elem()

		if columns == nil {
			for i := 0; i < maps.NumField(); i++ {
				tagBool := modelType.Field(i).Tag.Get("true")
				if tagBool == ""{
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
				valueField =maps.Field(index)
			}
			tagBool := modelField.Tag.Get("true")
			if tagBool == ""{
				r = append(r, valueField.Addr().Interface())
			} else {
				var str string
				swapValues[index] = reflect.New(reflect.TypeOf(str)).Elem().Addr().Interface()
				r = append(r, swapValues[index])
			}

		}
	}
	return
}
func getColumnNameForSearch(modelType reflect.Type, sortField string) string {
	sortField = strings.TrimSpace(sortField)
	i, _, column := getFieldByJson(modelType, sortField)
	if i > -1 {
		return column
	}
	return sortField // injection
}

func getColumnsSelect(modelType reflect.Type) []string {
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
						columnName := str2[j+1]
						columnNameTag := getColumnNameFromSqlBuilderTag(field)
						if columnNameTag != nil {
							columnName = *columnNameTag
						}
						columnNameKeys = append(columnNameKeys, columnName)
					}
				}
			}
		}
	}
	return columnNameKeys
}

func getSortType(sortType string) string {
	if sortType == "-" {
		return desc
	} else {
		return asc
	}
}
