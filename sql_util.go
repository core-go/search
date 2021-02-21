package search

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
)

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
func GetColumnName(modelType reflect.Type, fieldName string) (col string, colExist bool) {
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

func Count(db *sql.DB, sql string, values ...interface{}) (int64, error) {
	var total int64
	row := db.QueryRow(sql, values...)
	err2 := row.Scan(&total)
	if err2 != nil {
		return total, err2
	}
	return total, nil
}

func Query(db *sql.DB, results interface{}, fieldsIndex map[string]int, sql string, values ...interface{}) error {
	rows, er1 := db.Query(sql, values...)
	if er1 != nil {
		return er1
	}
	defer rows.Close()
	modelType := reflect.TypeOf(results).Elem().Elem()

	if fieldsIndex == nil {
		tb, er2 := ScanSearchType(rows, modelType)
		if er2 != nil {
			return er2
		}
		if tb != nil {
			reflect.ValueOf(results).Elem().Set(reflect.ValueOf(tb).Elem())
		}
	} else {
		columns, er3 := rows.Columns()
		if er3 != nil {
			return er3
		}
		fieldsIndexSelected := make([]int, 0)
		for _, columnsName := range columns {
			if index, ok := fieldsIndex[columnsName]; ok {
				fieldsIndexSelected = append(fieldsIndexSelected, index)
			}
		}
		tb, er4 := ScanType(rows, modelType, fieldsIndexSelected)
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

func QueryAndCount(db *sql.DB, results interface{}, count *int64, driverName string, sql string, values ...interface{}) error {
	rows, er1 := db.Query(sql, values...)
	if er1 != nil {
		return er1
	}
	defer rows.Close()
	modelType := reflect.TypeOf(results).Elem().Elem()

	fieldsIndex, er0 := GetColumnIndexes(modelType, driverName)
	if er0 != nil {
		return er0
	}

	tb, c, er2 := ScansSearchAndCount(rows, modelType, fieldsIndex)
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

// StructScan : transfer struct to slice for scan
func StructScan(s interface{}, indexColumns []int) (r []interface{}) {
	if s != nil {
		maps := reflect.Indirect(reflect.ValueOf(s))
		for _, index := range indexColumns {
			r = append(r, maps.Field(index).Addr().Interface())
		}
	}
	return
}

func GetColumnIndexes(modelType reflect.Type, driver string) (map[string]int, error) {
	mapp := make(map[string]int, 0)
	if modelType.Kind() != reflect.Struct {
		return mapp, errors.New("bad type")
	}
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		ormTag := field.Tag.Get("gorm")
		column, ok := FindTag(ormTag, "column")
		if ok {
			if driver == DriverOracle {
				column = strings.ToUpper(column)
			}
			mapp[column] = i
		}
	}
	return mapp, nil
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

func ScanType(rows *sql.Rows, modelType reflect.Type, indexes []int) (t []interface{}, err error) {
	for rows.Next() {
		initModel := reflect.New(modelType).Interface()
		if err = rows.Scan(StructScan(initModel, indexes)...); err == nil {
			t = append(t, initModel)
		}
	}
	return
}

func ScanSearchType(rows *sql.Rows, modelType reflect.Type) (t []interface{}, err error) {
	for rows.Next() {
		gTb := reflect.New(modelType).Interface()
		if err = rows.Scan(StructSearchScan(gTb)...); err == nil {
			t = append(t, gTb)
		}
	}

	return
}

func ScansSearchAndCount(rows *sql.Rows, modelType reflect.Type, fieldsIndex map[string]int) ([]interface{}, int64, error) {
	var t []interface{}
	columns, er0 := rows.Columns()
	if er0 != nil {
		return nil, 0, er0
	}
	var count int64
	for rows.Next() {
		initModel := reflect.New(modelType).Interface()
		var c []interface{}
		c = append(c, &count)
		c = append(c, StructScanWithIgnore(initModel, fieldsIndex, columns, 0)...)
		if err := rows.Scan(c...); err == nil {
			t = append(t, initModel)
		}
	}
	return t, count, nil
}

// StructScan : transfer struct to slice for scan
func StructScanWithIgnore(s interface{}, fieldsIndex map[string]int, columns []string, indexIgnore int) (r []interface{}) {
	if s != nil {
		maps := reflect.Indirect(reflect.ValueOf(s))
		fieldsIndexSelected := make([]int, 0)
		for i, columnsName := range columns {
			if i == indexIgnore {
				continue
			}
			if index, ok := fieldsIndex[columnsName]; ok {
				fieldsIndexSelected = append(fieldsIndexSelected, index)
				r = append(r, maps.Field(index).Addr().Interface())
			} else {
				var t interface{}
				r = append(r, &t)
			}
		}
	}
	return
}

func StructSearchScan(s interface{}) (r []interface{}) {
	if s != nil {
		vals := reflect.Indirect(reflect.ValueOf(s))
		for i := 0; i < vals.NumField(); i++ {
			r = append(r, vals.Field(i).Addr().Interface())
		}
	}

	return
}
func GetColumnNameForSearch(modelType reflect.Type, sortField string) string {
	sortField = strings.TrimSpace(sortField)
	i, _, column := GetFieldByJson(modelType, sortField)
	if i > -1 {
		return column
	}
	return sortField // injection
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
						columnName := str2[j+1]
						columnNameKeys = append(columnNameKeys, columnName)
					}
				}
			}
		}
	}
	return columnNameKeys
}

func GetSortType(sortType string) string {
	if sortType == "-" {
		return desc
	} else {
		return asc
	}
}
