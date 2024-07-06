package query

import (
	"fmt"
	s "github.com/core-go/search"
	"reflect"
	"strings"
	"time"
)

const (
	desc = "desc"
	asc  = "asc"
)

type Builder[T any, F any] struct {
	TableName  string
	ModelType  reflect.Type
	BuildParam func(int) string
}

func UseQuery[T any, F any](tableName string, options ...func(int) string) func(F) (string, []interface{}) {
	b := NewBuilder[T, F](tableName, options...)
	return b.BuildQuery
}
func NewBuilder[T any, F any](tableName string, options ...func(int) string) *Builder[T, F] {
	var build func(int) string
	if len(options) > 0 {
		build = options[0]
	} else {
		build = BuildParam
	}
	var t T
	resultModelType := reflect.TypeOf(t)
	if resultModelType.Kind() == reflect.Ptr {
		resultModelType = resultModelType.Elem()
	}
	return &Builder[T, F]{TableName: tableName, ModelType: resultModelType, BuildParam: build}
}

const (
	like             = "like"
	greaterEqualThan = ">="
	greaterThan      = ">"
	lessEqualThan    = "<="
	lessThan         = "<"
	in               = "in"
)

func getStringFromTag(typeOfField reflect.StructField, tagName string, key string) *string {
	tag := typeOfField.Tag
	properties := strings.Split(tag.Get(tagName), ";")
	for _, property := range properties {
		if strings.HasPrefix(property, key) {
			column := property[len(key):]
			return &column
		}
	}
	return nil
}

func getJoinFromSqlBuilderTag(typeOfField reflect.StructField) *string {
	return getStringFromTag(typeOfField, "sql_builder", "join:")
}

func getColumnNameFromSqlBuilderTag(typeOfField reflect.StructField) *string {
	return getStringFromTag(typeOfField, "sql_builder", "column:")
}
func (b *Builder[T, F]) BuildQuery(fm F) (string, []interface{}) {
	return Build(fm, b.TableName, b.ModelType, b.BuildParam)
}
func Build(fm interface{}, tableName string, modelType reflect.Type, buildParam func(int) string) (string, []interface{}) {
	if buildParam == nil {
		buildParam = BuildParam
	}
	s1 := ""
	rawConditions := make([]string, 0)
	queryValues := make([]interface{}, 0)
	qQueryValues := make([]string, 0)
	qCols := make([]string, 0)
	rawJoin := make([]string, 0)
	sortString := ""
	fields := make([]string, 0)
	var excluding []string
	var keyword string
	value := reflect.Indirect(reflect.ValueOf(fm))
	typeOfValue := value.Type()
	numField := value.NumField()
	var idCol string
	marker := 0
	fCount := 0
	for i := 0; i < numField; i++ {
		field := value.Field(i)
		kind := field.Kind()
		x := field.Interface()
		typeOfField := value.Type().Field(i)
		param := buildParam(marker + 1)

		if v, ok := x.(*s.Filter); ok {
			if len(v.Fields) > 0 {
				for _, key := range v.Fields {
					i, _, columnName := getFieldByJson(modelType, key)
					if len(columnName) < 0 {
						fields = fields[len(fields):]
						break
					} else if i > -1 {
						fields = append(fields, columnName)
					}
				}
			}
			if len(fields) > 0 {
				s1 = `select ` + strings.Join(fields, ",") + ` from ` + tableName
			}
			if len(v.Sort) > 0 {
				sortString = buildSort(v.Sort, modelType)
			}
		}

		columnName, existCol := getColumnName(value.Type(), typeOfField.Name)
		if !existCol {
			columnName, _ = getColumnName(modelType, typeOfField.Name)
		}

		columnNameFromSqlBuilderTag := getColumnNameFromSqlBuilderTag(typeOfField)
		if columnNameFromSqlBuilderTag != nil {
			columnName = *columnNameFromSqlBuilderTag
		}

		joinFromSqlBuilderTag := getJoinFromSqlBuilderTag(typeOfField)
		if joinFromSqlBuilderTag != nil {
			rawJoin = append(rawJoin, *joinFromSqlBuilderTag)
		}
		ps := false
		var value2 string
		tag := typeOfValue.Field(i).Tag
		isContinue := false
		isStrPointer := false
		if kind == reflect.Ptr {
			if field.IsNil() {
				isContinue = true
				isStrPointer = true
			} else {
				s0, ok0 := x.(*string)
				if ok0 {
					if s0 == nil || len(*s0) == 0 {
						isContinue = true
						isStrPointer = true
					}
					ps = true
					value2 = *s0
				}
				field = field.Elem()
				x = field.Interface()
				kind = field.Kind()
			}
		}
		if !isStrPointer {
			s0, ok0 := x.(string)
			if ok0 {
				if len(s0) == 0 {
					isContinue = true
				}
				value2 = s0
			}
		}
		if isContinue {
			if len(keyword) > 0 {
				qMatch, isQ := tag.Lookup("q")
				if isQ {
					if qMatch == "=" {
						qQueryValues = append(qQueryValues, keyword)

					} else if qMatch == "like" {
						qQueryValues = append(qQueryValues, buildQ(keyword))
					} else {
						qQueryValues = append(qQueryValues, prefix(keyword))
					}
					qCols = append(qCols, columnName)
				}
			}
			continue
		}
		if v, ok := x.(s.Filter); ok {
			if v.Excluding != nil && len(v.Excluding) > 0 {
				index, _, columnName := getFieldByBson(value.Type(), "_id")
				if !(index == -1 || columnName == "") {
					idCol = columnName
					excluding = v.Excluding
				}
			}
			if len(v.Q) > 0 {
				keyword = strings.TrimSpace(v.Q)
			}
			continue
		} else if ps || kind == reflect.String {
			if len(value2) > 0 {
				key, ok := tag.Lookup("operator")
				if !ok {
					key, _ = tag.Lookup("q")
				}
				if key == "=" {
					rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, "=", param))
				} else {
					rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, like, param))
					fCount = fCount + 1
					if key == "like" {
						queryValues = append(queryValues, buildQ(value2))
					} else {
						queryValues = append(queryValues, prefix(value2))
					}
				}
				marker++
			}
		} else if dateTime, ok := x.(s.TimeRange); ok {
			if dateTime.Min != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, greaterEqualThan, param))
				queryValues = append(queryValues, dateTime.Min)
				marker += 1
			}
			if dateTime.Max != nil {
				param := buildParam(marker + 1)
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, lessEqualThan, param))
				queryValues = append(queryValues, dateTime.Max)
				marker += 1
			} else if dateTime.Top != nil {
				param := buildParam(marker + 1)
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, lessThan, param))
				queryValues = append(queryValues, dateTime.Top)
				marker += 1
			}
		} else if numberRange, ok := x.(s.NumberRange); ok {
			if numberRange.Min != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, greaterEqualThan, param))
				queryValues = append(queryValues, numberRange.Min)
				marker++
			} else if numberRange.Bottom != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, greaterThan, param))
				queryValues = append(queryValues, numberRange.Bottom)
				marker++
			}
			if numberRange.Max != nil {
				param := buildParam(marker + 1)
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, lessEqualThan, param))
				queryValues = append(queryValues, numberRange.Max)
				marker++
			} else if numberRange.Top != nil {
				param := buildParam(marker + 1)
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, lessThan, param))
				queryValues = append(queryValues, numberRange.Top)
				marker++
			}
		} else if numberRange, ok := x.(s.Int64Range); ok {
			if numberRange.Min != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, greaterEqualThan, numberRange.Min))
			} else if numberRange.Bottom != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, greaterThan, numberRange.Bottom))
			}
			if numberRange.Max != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, lessEqualThan, numberRange.Max))
			} else if numberRange.Top != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, lessThan, numberRange.Top))
			}
		} else if numberRange, ok := x.(s.IntRange); ok {
			if numberRange.Min != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, greaterEqualThan, numberRange.Min))
			} else if numberRange.Bottom != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, greaterThan, numberRange.Bottom))
			}
			if numberRange.Max != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, lessEqualThan, numberRange.Max))
			} else if numberRange.Top != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, lessThan, numberRange.Top))
			}
		} else if numberRange, ok := x.(s.Int32Range); ok {
			if numberRange.Min != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, greaterEqualThan, numberRange.Min))
			} else if numberRange.Bottom != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, greaterThan, numberRange.Bottom))
			}
			if numberRange.Max != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, lessEqualThan, numberRange.Max))
			} else if numberRange.Top != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %d", columnName, lessThan, numberRange.Top))
			}
		} else if dateRange, ok := x.(s.DateRange); ok {
			if dateRange.Min != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, greaterEqualThan, param))
				queryValues = append(queryValues, dateRange.Min)
				marker += 1
			}
			if dateRange.Max != nil {
				param := buildParam(marker + 1)
				var eDate = dateRange.Max.Add(time.Hour * 24)
				dateRange.Max = &eDate
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, lessThan, param))
				queryValues = append(queryValues, dateRange.Max)
				marker += 1
			}
		} else if kind == reflect.Slice {
			if field.Len() > 0 {
				format := fmt.Sprintf("(%s)", buildParametersFrom(marker, field.Len(), buildParam))
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, in, format))
				queryValues = extractArray(queryValues, x)
				marker += field.Len()
			}
		} else {
			key, ok := tag.Lookup("operator")
			if !ok {
				key = "="
			}
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, key, param))
			queryValues = append(queryValues, x)
			marker += 1
		}
	}

	if excluding != nil && len(excluding) > 0 && len(idCol) > 0 {
		format := fmt.Sprintf("(%s)", buildParametersFrom(marker, len(excluding), buildParam))
		marker += len(excluding)
		rawConditions = append(rawConditions, fmt.Sprintf("%s NOT IN %s", idCol, format))
		queryValues = extractArray(queryValues, excluding)
	}
	if len(s1) == 0 {
		columns := getColumnsSelect(modelType)
		if len(columns) > 0 {
			s1 = `select  ` + strings.Join(columns, ",") + ` from ` + tableName
		} else {
			s1 = `select * from ` + tableName
		}
	}
	if len(rawJoin) > 0 {
		s1 = s1 + " " + strings.Join(rawJoin, " ")
	}
	if len(qCols) > 0 {
		qConditions := make([]string, 0)
		for i, s := range qCols {
			param := buildParam(marker + 1)
			qConditions = append(qConditions, fmt.Sprintf("%s %s %s", s, like, param))
			fCount = fCount + 1
			queryValues = append(queryValues, qQueryValues[i])
			marker++
		}
		if len(qConditions) > 0 {
			rawConditions = append(rawConditions, " ("+strings.Join(qConditions, " or ")+") ")
		}
	}
	allowFiltering := ""
	if fCount >= 1 {
		allowFiltering = " allow filtering"
	}
	if len(rawConditions) > 0 {
		s2 := s1 + ` where ` + strings.Join(rawConditions, " AND ") + sortString + allowFiltering
		return s2, queryValues
	}
	s3 := s1 + sortString + allowFiltering
	return s3, queryValues
}
func extractArray(values []interface{}, field interface{}) []interface{} {
	s := reflect.Indirect(reflect.ValueOf(field))
	for i := 0; i < s.Len(); i++ {
		values = append(values, s.Index(i).Interface())
	}
	return values
}
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
func getFieldByBson(modelType reflect.Type, bsonName string) (int, string, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		tag1, ok1 := field.Tag.Lookup("bson")
		if ok1 && strings.Split(tag1, ",")[0] == bsonName {
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
	return -1, bsonName, bsonName
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
func buildSort(sortString string, modelType reflect.Type) string {
	var sort = make([]string, 0)
	sorts := strings.Split(sortString, ",")
	for i := 0; i < len(sorts); i++ {
		sortField := strings.TrimSpace(sorts[i])
		fieldName := sortField
		c := sortField[0:1]
		if c == "-" || c == "+" {
			fieldName = sortField[1:]
		}
		columnName := getColumnNameForSearch(modelType, fieldName)
		if len(columnName) > 0 {
			sortType := getSortType(c)
			sort = append(sort, columnName+" "+sortType)
		}
	}
	if len(sort) > 0 {
		return ` order by ` + strings.Join(sort, ",")
	} else {
		return ""
	}
}
func getColumnNameForSearch(modelType reflect.Type, sortField string) string {
	sortField = strings.TrimSpace(sortField)
	i, _, column := getFieldByJson(modelType, sortField)
	if i > -1 {
		return column
	}
	return ""
}
func getSortType(sortType string) string {
	if sortType == "-" {
		return desc
	} else {
		return asc
	}
}

func BuildParam(i int) string {
	return "?"
}
func buildParametersFrom(i int, numCol int, buildParam func(i int) string) string {
	var arrValue []string
	for j := 0; j < numCol; j++ {
		arrValue = append(arrValue, buildParam(i+j+1))
	}
	return strings.Join(arrValue, ",")
}
func buildQ(s string) string {
	if !(strings.HasPrefix(s, "%") && strings.HasSuffix(s, "%")) {
		return "%" + s + "%"
	} else if strings.HasPrefix(s, "%") {
		return s + "%"
	} else if strings.HasSuffix(s, "%") {
		return "%" + s
	}
	return s
}
func prefix(s string) string {
	if strings.HasSuffix(s, "%") {
		return s
	} else {
		return s + "%"
	}
}
