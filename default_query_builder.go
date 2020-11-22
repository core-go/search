package search

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

type DefaultQueryBuilder struct {
	TableName  string
	ModelType  reflect.Type
	DriverName string
}

func NewQueryBuilder(db *sql.DB, tableName string, modelType reflect.Type) *DefaultQueryBuilder {
	driverName := GetDriver(db)
	return NewDefaultQueryBuilder(tableName, modelType, driverName)
}
func NewDefaultQueryBuilder(tableName string, modelType reflect.Type, driverName string) *DefaultQueryBuilder {
	return &DefaultQueryBuilder{TableName: tableName, ModelType: modelType, DriverName: driverName}
}

const (
	Exact            = "="
	Like             = "like"
	GreaterEqualThan = ">="
	GreaterThan      = ">"
	LighterEqualThan = "<="
	LighterThan      = "<"
	In               = "in"
)

func GetColumnNameFromSqlBuilderTag(typeOfField reflect.StructField) *string {
	tag := typeOfField.Tag
	properties := strings.Split(tag.Get("sql_builder"), ";")
	for _, property := range properties {
		if strings.HasPrefix(property, "column:") {
			column := property[7:]
			return &column
		}
	}
	return nil
}
func (b *DefaultQueryBuilder) BuildQuery(sm interface{}) (string, []interface{}) {
	return BuildQuery(sm, b.TableName, b.ModelType, b.DriverName)
}
func BuildQuery(sm interface{}, tableName string, modelType reflect.Type, driverName string) (string, []interface{}) {
	s1 := ""
	rawConditions := make([]string, 0)
	queryValues := make([]interface{}, 0)
	sortString := ""
	fields := make([]string, 0)
	var keyword string
	var keywordFormat map[string]string
	keywordFormat = map[string]string{
		"prefix":  "?%",
		"contain": "%?%",
		"equal":   "?",
	}

	value := reflect.Indirect(reflect.ValueOf(sm))
	typeOfValue := value.Type()
	numField := value.NumField()
	marker := 0

	for i := 0; i < numField; i++ {
		field := value.Field(i)
		kind := field.Kind()
		x := field.Interface()
		typeOfField := value.Type().Field(i)
		param := BuildParam(marker+1, driverName)

		if v, ok := x.(*SearchModel); ok {
			if len(v.Fields) > 0 {
				for _, key := range v.Fields {
					i, _, columnName := GetFieldByJson(modelType, key)
					if len(columnName) < 0 {
						fields = fields[len(fields):]
						break
					} else if i == -1 {
						columnName = strings.ToLower(key) // injection
					}
					fields = append(fields, columnName)
				}
			}
			if len(fields) > 0 {
				s1 = `select ` + strings.Join(fields, ",") + ` from ` + tableName
			} else {
				columns := GetColumnsSelect(modelType)
				if len(columns) > 0 {
					s1 = `select  ` + strings.Join(columns, ",") + ` from ` + tableName
				} else {
					s1 = `select * from ` + tableName
				}
			}
			if len(v.Sort) > 0 {
				sortString = BuildSort(v.Sort, modelType)
			}
		}

		columnName, existCol := GetColumnName(value.Type(), typeOfField.Name)
		if !existCol {
			columnName, _ = GetColumnName(modelType, typeOfField.Name)
		}
		columnNameFromSqlBuilderTag := GetColumnNameFromSqlBuilderTag(typeOfField)
		if columnNameFromSqlBuilderTag != nil {
			columnName = *columnNameFromSqlBuilderTag
		}
		if kind == reflect.Ptr && field.IsNil() {
			continue
		}
		if kind == reflect.Ptr {
			field = field.Elem()
			kind = field.Kind()
		}
		if v, ok := x.(*SearchModel); ok {
			if len(v.Excluding) > 0 {
				for key, val := range v.Excluding {
					index, _, columnName := GetFieldByJson(value.Type(), key)
					if index == -1 || columnName == "" {
						log.Panic("column name not found")
					}
					if len(val) > 0 {
						format := fmt.Sprintf("(%s)", BuildParametersFrom(marker, len(val), driverName))
						marker += len(val) - 1
						rawConditions = append(rawConditions, fmt.Sprintf("%s NOT IN %s", columnName, format))
						queryValues = ExtractArray(queryValues, val)
					}
				}
			} else if len(v.Keyword) > 0 {
				keyword = strings.TrimSpace(v.Keyword)
			}
			continue
		} else if dateRange, ok := x.(DateRange); ok {
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, GreaterEqualThan, param))
			queryValues = append(queryValues, dateRange.StartDate)
			var eDate = dateRange.EndDate.Add(time.Hour * 24)
			dateRange.EndDate = &eDate
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, LighterThan, param))
			queryValues = append(queryValues, dateRange.EndDate)
			marker += 2
		} else if dateRange, ok := x.(*DateRange); ok && dateRange != nil {
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, GreaterEqualThan, param))
			queryValues = append(queryValues, dateRange.StartDate)
			var eDate = dateRange.EndDate.Add(time.Hour * 24)
			dateRange.EndDate = &eDate
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, LighterThan, param))
			queryValues = append(queryValues, dateRange.EndDate)
			marker += 2
		} else if dateTime, ok := x.(TimeRange); ok {
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, GreaterEqualThan, param))
			queryValues = append(queryValues, dateTime.StartTime)
			var eDate = dateTime.EndTime.Add(time.Hour * 24)
			dateTime.EndTime = &eDate
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, LighterThan, param))
			queryValues = append(queryValues, dateTime.EndTime)
			marker += 2
		} else if dateTime, ok := x.(*TimeRange); ok && dateTime != nil {
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, GreaterEqualThan, param))
			queryValues = append(queryValues, dateTime.StartTime)
			var eDate = dateTime.EndTime.Add(time.Hour * 24)
			dateTime.EndTime = &eDate
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, LighterThan, param))
			queryValues = append(queryValues, dateTime.EndTime)
			marker += 2
		} else if numberRange, ok := x.(NumberRange); ok {
			if numberRange.Min != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, GreaterEqualThan, param))
				queryValues = append(queryValues, numberRange.Min)
				marker++
			} else if numberRange.Lower != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, GreaterThan, param))
				queryValues = append(queryValues, numberRange.Lower)
				marker++
			}
			if numberRange.Max != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, LighterEqualThan, param))
				queryValues = append(queryValues, numberRange.Max)
				marker++
			} else if numberRange.Upper != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, LighterThan, param))
				queryValues = append(queryValues, numberRange.Upper)
				marker++
			}
		} else if numberRange, ok := x.(*NumberRange); ok && numberRange != nil {
			if numberRange.Min != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, GreaterEqualThan, param))
				queryValues = append(queryValues, numberRange.Min)
				marker++
			} else if numberRange.Lower != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, GreaterThan, param))
				queryValues = append(queryValues, numberRange.Lower)
				marker++
			}
			if numberRange.Max != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, LighterEqualThan, param))
				queryValues = append(queryValues, numberRange.Max)
				marker++
			} else if numberRange.Upper != nil {
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, LighterThan, param))
				queryValues = append(queryValues, numberRange.Upper)
				marker++
			}
		} else if kind == reflect.String {
			var searchValue bool
			if field.Len() > 0 {
				const defaultKey = "contain"
				if key, ok := typeOfValue.Field(i).Tag.Lookup("match"); ok {
					if format, exist := keywordFormat[key]; exist {
						searchValue = true
						value2, valid := x.(string)
						if !valid {
							log.Panicf("invalid data \"%v\" \n", x)
						}
						//if sql == "mysql" {
						//	value2 = EscapeString(value2)
						//} else if sql == "postgres" || sql == "mssql" {
						//	value2 = EscapeStringForSelect(value2)
						//}
						value2 = func(format, s string) string {
							return strings.Replace(format, "?", s, -1)
						}(format, value2)
						//value2 = value2 + `%`
						//rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, Like))
						queryValues = append(queryValues, value2)
					} else {
						log.Panicf("match not support \"%v\" format\n", key)
					}
				} else if format, exist := keywordFormat[defaultKey]; exist {
					searchValue = true
					value2, valid := x.(string)
					if !valid {
						log.Panicf("invalid data \"%v\" \n", x)
					}
					//if sql == "mysql" {
					//	value2 = EscapeString(value2)
					//} else if sql == "postgres" || sql == "mssql" {
					//	value2 = EscapeStringForSelect(value2)
					//}
					//value2 = `%` + value2 + `%`
					value2 = func(format, s string) string {
						return strings.Replace(format, "?", s, -1)
					}(format, value2)
					//rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, Like))
					queryValues = append(queryValues, value2)
				} else {
					searchValue = true
					value2, valid := x.(string)
					if !valid {
						log.Panicf("invalid data \"%v\" \n", x)
					}
					value2 = value2 + `%`
					//rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, Like))
					queryValues = append(queryValues, value2)
				}
			} else if len(keyword) > 0 {
				if key, ok := typeOfValue.Field(i).Tag.Lookup("keyword"); ok {
					if format, exist := keywordFormat[key]; exist {
						//if sql == "mysql" {
						//	keyword = EscapeString(keyword)
						//} else if sql == "postgres" || sql == "mssql" {
						//	keyword = EscapeStringForSelect(keyword)
						//}
						if format == `?%` {
							keyword = keyword + `%`
						} else if format == `%?%` {
							keyword = `%` + keyword + `%`
						} else {
							log.Panicf("keyword not support \"%v\" format\n", key)
						}

						queryValues = append(queryValues, keyword)
					} else {
						log.Panicf("keyword not support \"%v\" format\n", key)
					}
				}
			}
			if searchValue {
				if driverName == DriverPostgres { // "postgres"
					rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, `ilike`, param))
				} else {
					rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, Like, param))
				}
				marker++
			}
		} else if kind == reflect.Slice {
			if field.Len() > 0 {
				format := fmt.Sprintf("(%s)", BuildParametersFrom(marker, field.Len(), driverName))
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, In, format))
				queryValues = ExtractArray(queryValues, x)
				marker += field.Len()
			}
		} else {
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, Exact, param))
			queryValues = append(queryValues, x)
		}
	}
	if len(rawConditions) > 0 {
		s2 := s1 + ` where ` + strings.Join(rawConditions, " AND ") + sortString
		return s2, queryValues
	}
	s3 := s1 + sortString
	return s3, queryValues
}

func ExtractArray(values []interface{}, field interface{}) []interface{} {
	s := reflect.Indirect(reflect.ValueOf(field))
	for i := 0; i < s.Len(); i++ {
		values = append(values, s.Index(i).Interface())
	}
	return values
}
func BuildSort(sortString string, modelType reflect.Type) string {
	var sort = make([]string, 0)
	sorts := strings.Split(sortString, ",")
	for i := 0; i < len(sorts); i++ {
		sortField := strings.TrimSpace(sorts[i])
		fieldName := sortField
		c := sortField[0:1]
		if c == "-" || c == "+" {
			fieldName = sortField[1:]
		}
		columnName := GetColumnNameForSearch(modelType, fieldName)
		sortType := GetSortType(c)
		sort = append(sort, columnName+" "+sortType)
	}
	return ` order by ` + strings.Join(sort, ",")
}
func ReplaceParameters(sql string, number int, prefix string) string {
	for i := 0; i < number; i++ {
		count := i + 1
		sql = strings.Replace(sql, "?", prefix+fmt.Sprintf("%v", count), 1)
	}
	return sql
}
func BuildQueryByDriver(sql string, number int, driverName string) string {
	switch driverName {
	case DriverPostgres:
		return ReplaceParameters(sql, number, "$")
	case DriverOracle:
		return ReplaceParameters(sql, number, ":val")
	default:
		return sql
	}
}
