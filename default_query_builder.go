package search

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type DefaultQueryBuilder struct {
	TableName     string
	ModelType     reflect.Type
	QuestionParam bool
}

func NewQueryBuilder(tableName string, modelType reflect.Type) *DefaultQueryBuilder {
	return &DefaultQueryBuilder{TableName: tableName, ModelType: modelType, QuestionParam: true}
}

const (
	Exact            = "="
	Like             = "LIKE"
	GreaterEqualThan = ">="
	LighterEqualThan = "<="
	LighterThan      = "<"
	In               = "IN"
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
	for i := 0; i < numField; i++ {
		field := value.Field(i)
		kind := field.Kind()
		interfaceOfField := field.Interface()
		typeOfField := value.Type().Field(i)

		if v, ok := interfaceOfField.(*SearchModel); ok {
			if len(v.Fields) > 0 {
				for _, key := range v.Fields {
					i, _, columnName := GetFieldByJson(b.ModelType, key)
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
				s1 = `SELECT ` + strings.Join(fields, ",") + ` FROM ` + b.TableName
			} else {
				columns := GetColumnsSelect(b.ModelType)
				if len(columns) > 0 {
					s1 = `SELECT  ` + strings.Join(columns, ",") + ` FROM ` + b.TableName
				} else {
					s1 = `SELECT * FROM ` + b.TableName
				}
			}
			if len(v.Sort) > 0 {
				sortString = BuildSort(v.Sort, b.ModelType)
			}
		}

		columnName, existCol := GetColumnName(value.Type(), typeOfField.Name)
		if !existCol {
			columnName, _ = GetColumnName(b.ModelType, typeOfField.Name)
		}
		columnNameFromSqlBuilderTag := GetColumnNameFromSqlBuilderTag(typeOfField)
		if columnNameFromSqlBuilderTag != nil {
			columnName = *columnNameFromSqlBuilderTag
		}

		if kind == reflect.Ptr && field.IsNil() {
			continue
		} else if v, ok := interfaceOfField.(*SearchModel); ok {
			if len(v.Excluding) > 0 {
				r := regexp.MustCompile(`[A-Z]`)
				for key, val := range v.Excluding {
					columnName = r.ReplaceAllStringFunc(key, func(m string) string {
						out := "_" + strings.ToLower(m)
						return out
					})
					if len(val) > 0 {
						rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s (?)", columnName, "NOT", In))
						queryValues = append(queryValues, val)
					}
				}
			} else if len(v.Keyword) > 0 {
				keyword = strings.TrimSpace(v.Keyword)
			}
			continue
		} else if dateRange, ok := interfaceOfField.(DateRange); ok {
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, GreaterEqualThan))
			queryValues = append(queryValues, dateRange.StartDate)
			var eDate = dateRange.EndDate.Add(time.Hour * 24)
			dateRange.EndDate = &eDate
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, LighterThan))
			queryValues = append(queryValues, dateRange.EndDate)
		} else if dateTime, ok := interfaceOfField.(TimeRange); ok {
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, GreaterEqualThan))
			queryValues = append(queryValues, dateTime.StartTime)
			var eDate = dateTime.EndTime.Add(time.Hour * 24)
			dateTime.EndTime = &eDate
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, LighterThan))
			queryValues = append(queryValues, dateTime.EndTime)
		} else if kind == reflect.String {
			var searchValue string
			if field.Len() > 0 {
				const defaultKey = "contain"
				if key, ok := typeOfValue.Field(i).Tag.Lookup("match"); ok {
					if _, exist := keywordFormat[key]; exist {
						searchValue = `?`
						value2, valid := interfaceOfField.(string)
						if !valid {
							log.Panicf("invalid data \"%v\" \n", interfaceOfField)
						}
						//if sql == "mysql" {
						//	value2 = EscapeString(value2)
						//} else if sql == "postgres" || sql == "mssql" {
						//	value2 = EscapeStringForSelect(value2)
						//}
						value2 = value2 + `%`
						//rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, Like))
						queryValues = append(queryValues, value2)
					} else {
						log.Panicf("match not support \"%v\" format\n", key)
					}
				} else if _, exist := keywordFormat[defaultKey]; exist {
					searchValue = `?`
					value2, valid := interfaceOfField.(string)
					if !valid {
						log.Panicf("invalid data \"%v\" \n", interfaceOfField)
					}
					//if sql == "mysql" {
					//	value2 = EscapeString(value2)
					//} else if sql == "postgres" || sql == "mssql" {
					//	value2 = EscapeStringForSelect(value2)
					//}
					value2 = `%` + value2 + `%`
					//rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, Like))
					queryValues = append(queryValues, value2)
				} else {
					searchValue = `?`
					value2, valid := interfaceOfField.(string)
					if !valid {
						log.Panicf("invalid data \"%v\" \n", interfaceOfField)
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
						searchValue = `?`
						queryValues = append(queryValues, keyword)
					} else {
						log.Panicf("keyword not support \"%v\" format\n", key)
					}
				}
			}
			if len(searchValue) > 0 {
				if !b.QuestionParam { // "postgres"
					rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, `ILIKE`, searchValue))
				} else {
					rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, Like, searchValue))
				}
			}
		} else if kind == reflect.Slice {
			if field.Len() > 0 {
				format := "(?"
				for i := 0; i < field.Len()-1; i++ {
					format += ", ?"
				}
				format += ")"
				rawConditions = append(rawConditions, fmt.Sprintf("%s %s %s", columnName, In, format))
				queryValues = append(queryValues, interfaceOfField)
			}
		} else {
			rawConditions = append(rawConditions, fmt.Sprintf("%s %s ?", columnName, Exact))
			queryValues = append(queryValues, interfaceOfField)
		}
	}
	if len(rawConditions) > 0 {
		return s1 + ` WHERE ` + strings.Join(rawConditions, " AND ") + sortString, queryValues
	}
	return s1 + sortString, queryValues
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
	return ` ORDER BY ` + strings.Join(sort, ",")
}
