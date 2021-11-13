package query

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	d "github.com/core-go/dynamodb"
	"github.com/core-go/search"
	"log"
	"reflect"
	"strings"
)

func UseQuery(tableName string, resultModelType reflect.Type, index d.SecondaryIndex) func(interface{}) (dynamodb.ScanInput, error) {
	b := NewBuilder(tableName, resultModelType, index)
	return b.BuildQuery
}

type Builder struct {
	TableName string
	ModelType reflect.Type
	Index     d.SecondaryIndex
}

func NewBuilder(tableName string, resultModelType reflect.Type, index d.SecondaryIndex) *Builder {
	return &Builder{TableName: tableName, ModelType: resultModelType, Index: index}
}

func (b *Builder) BuildQuery(sm interface{}) (dynamodb.ScanInput, error) {
	return Build(sm, b.ModelType, b.TableName)
}

func Build(sm interface{}, modelType reflect.Type, tableName string) (dynamodb.ScanInput, error) {
	query := dynamodb.ScanInput{}
	if _, ok := sm.(*search.Filter); ok {
		return query, nil
	}
	var conditionBuilders []*expression.ConditionBuilder
	var projectionBuilder *expression.ProjectionBuilder
	var keyword string
	value := reflect.Indirect(reflect.ValueOf(sm))
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		kind := field.Kind()
		x := field.Interface()
		ps := false
		var psv string
		if kind == reflect.Ptr {
			if field.IsNil() {
				continue
			}
			s0, ok0 := x.(*string)
			if ok0 {
				if s0 == nil || len(*s0) == 0 {
					continue
				}
				ps = true
				psv = *s0
			}
			field = field.Elem()
			kind = field.Kind()
		}
		s0, ok0 := x.(string)
		if ok0 {
			if len(s0) == 0 {
				continue
			}
			psv = s0
		}
		ks := kind.String()
		if v, ok := x.(*search.Filter); ok {
			if v.Excluding != nil && len(v.Excluding) > 0 {
				if _, _, name, ok := getFieldByBson(modelType, "_id"); ok {
					c := expression.Not(expression.Name(name).In(expression.Value(v.Excluding)))
					conditionBuilders = append(conditionBuilders, &c)
				}
			}
			if len(v.Fields) > 0 {
				var proj expression.ProjectionBuilder
				for idx := range v.Fields {
					name := expression.Name(v.Fields[idx])
					proj = proj.AddNames(name)
					projectionBuilder = &proj
				}
			}
			continue
		} else if ps || ks == "string" {
			if _, name, ok := d.GetFieldByName(modelType, value.Type().Field(i).Name); ok {
				var condition expression.ConditionBuilder
				if !field.IsNil() {
					if key, ok := value.Type().Field(i).Tag.Lookup("match"); ok {
						if key == d.PREFIX {
							condition = expression.Name(name).BeginsWith(psv)
						} else if key == d.CONTAIN {
							condition = expression.Name(name).Contains(psv)
						} else if key == d.EQUAL {
							condition = expression.Name(name).Equal(expression.Value(psv))
						} else {
							log.Panicf("match not support \"%v\" format\n", key)
						}
					}
				} else if len(keyword) > 0 {
					if key, ok := value.Type().Field(i).Tag.Lookup("keyword"); ok {
						if key == d.PREFIX {
							condition = expression.Name(name).BeginsWith(psv)
						} else if key == d.CONTAIN {
							condition = expression.Name(name).Contains(psv)
						} else if key == d.EQUAL {
							condition = expression.Name(name).Equal(expression.Value(psv))
						} else {
							log.Panicf("match not support \"%v\" format\n", key)
						}
					}
				}
				conditionBuilders = append(conditionBuilders, &condition)
			}
		} else if rangeTime, ok := x.(*search.TimeRange); ok && rangeTime != nil {
			if _, name, ok := d.GetFieldByName(modelType, value.Type().Field(i).Name); ok {
				var arr []*expression.ConditionBuilder
				if rangeTime.Min != nil {
					gte := expression.Name(name).GreaterThanEqual(expression.Value(rangeTime.Min))
					arr = append(arr, &gte)
				}
				if rangeTime.Max != nil {
					lte := expression.Name(name).LessThanEqual(expression.Value(rangeTime.Max))
					arr = append(arr, &lte)
				} else if rangeTime.Top != nil {
					lt := expression.Name(name).LessThan(expression.Value(rangeTime.Top))
					arr = append(arr, &lt)
				}
				var f *expression.ConditionBuilder
				for idx := range arr {
					if f == nil {
						f = arr[idx]
					} else {
						f.And(*arr[idx])
					}
				}
				conditionBuilders = append(conditionBuilders, f)
			}
		} else if rangeTime, ok := x.(search.TimeRange); ok {
			if _, name, ok := d.GetFieldByName(modelType, value.Type().Field(i).Name); ok {
				var arr []*expression.ConditionBuilder
				if rangeTime.Min != nil {
					gte := expression.Name(name).GreaterThanEqual(expression.Value(rangeTime.Min))
					arr = append(arr, &gte)
				}
				if rangeTime.Max != nil {
					lte := expression.Name(name).LessThanEqual(expression.Value(rangeTime.Max))
					arr = append(arr, &lte)
				} else if rangeTime.Top != nil {
					lt := expression.Name(name).LessThan(expression.Value(rangeTime.Top))
					arr = append(arr, &lt)
				}
				var f *expression.ConditionBuilder
				for idx := range arr {
					if f == nil {
						f = arr[idx]
					} else {
						f.And(*arr[idx])
					}
				}
				conditionBuilders = append(conditionBuilders, f)
			}
		} else if numberRange, ok := x.(*search.NumberRange); ok && numberRange != nil {
			if _, name, ok := d.GetFieldByName(modelType, value.Type().Field(i).Name); ok {
				var arr []*expression.ConditionBuilder
				if numberRange.Min != nil {
					gte := expression.Name(name).GreaterThanEqual(expression.Value(numberRange.Min))
					arr = append(arr, &gte)
				} else if numberRange.Lower != nil {
					gt := expression.Name(name).GreaterThan(expression.Value(numberRange.Lower))
					arr = append(arr, &gt)
				}
				if numberRange.Max != nil {
					lte := expression.Name(name).LessThanEqual(expression.Value(numberRange.Max))
					arr = append(arr, &lte)
				} else if numberRange.Top != nil {
					lt := expression.Name(name).LessThan(expression.Value(numberRange.Top))
					arr = append(arr, &lt)
				}

				var f *expression.ConditionBuilder
				for idx := range arr {
					if f == nil {
						f = arr[idx]
					} else {
						f.And(*arr[idx])
					}
				}
				conditionBuilders = append(conditionBuilders, f)
			}
		} else if numberRange, ok := x.(search.NumberRange); ok {
			if _, name, ok := d.GetFieldByName(modelType, value.Type().Field(i).Name); ok {
				var arr []*expression.ConditionBuilder
				if numberRange.Min != nil {
					gte := expression.Name(name).GreaterThanEqual(expression.Value(numberRange.Min))
					arr = append(arr, &gte)
				} else if numberRange.Lower != nil {
					gt := expression.Name(name).GreaterThan(expression.Value(numberRange.Lower))
					arr = append(arr, &gt)
				}
				if numberRange.Max != nil {
					lte := expression.Name(name).LessThanEqual(expression.Value(numberRange.Max))
					arr = append(arr, &lte)
				} else if numberRange.Top != nil {
					lt := expression.Name(name).LessThan(expression.Value(numberRange.Top))
					arr = append(arr, &lt)
				}

				var f *expression.ConditionBuilder
				for idx := range arr {
					if f == nil {
						f = arr[idx]
					} else {
						f.And(*arr[idx])
					}
				}
				conditionBuilders = append(conditionBuilders, f)
			}
		} else if kind == reflect.Slice {
			if _, name, ok := d.GetFieldByName(modelType, value.Type().Field(i).Name); ok {
				condition := expression.Name(name).In(expression.Value(x))
				conditionBuilders = append(conditionBuilders, &condition)
			}
		} else {
			t := kind.String()
			if _, ok := x.(*search.Filter); t == "bool" || (strings.Contains(t, "int") && x != 0) || (strings.Contains(t, "float") && x != 0) || (!ok && t == "string" && field.Len() > 0) || (!ok && t == "ptr" &&
				field.Pointer() != 0) {
				if _, name, ok := d.GetFieldByName(modelType, value.Type().Field(i).Name); ok {
					c := expression.Name(name).Equal(expression.Value(x))
					conditionBuilders = append(conditionBuilders, &c)
				}
			}
		}
	}

	//test := conditionBuilders[0]
	var filter *expression.ConditionBuilder

	if conditionBuilders == nil && projectionBuilder == nil {
		query = dynamodb.ScanInput{
			TableName: aws.String(tableName),
			Select:    aws.String(dynamodb.SelectAllAttributes),
		}
		return query, nil
	} else if conditionBuilders == nil && projectionBuilder != nil {
		expr, _ := expression.NewBuilder().WithProjection(*projectionBuilder).Build()
		query = dynamodb.ScanInput{
			TableName:                aws.String(tableName),
			ProjectionExpression:     expr.Projection(),
			ExpressionAttributeNames: expr.Names(),
			Select:                   aws.String(dynamodb.SelectSpecificAttributes),
		}
		return query, nil
	} else if conditionBuilders != nil {
		for i := range conditionBuilders {
			if filter == nil {
				filter = conditionBuilders[i]
			} else {
				filt := filter.And(*conditionBuilders[i])
				filter = &filt
			}
		}

		if projectionBuilder == nil {
			expr, _ := expression.NewBuilder().WithFilter(*filter).Build()
			query = dynamodb.ScanInput{
				TableName:                 aws.String(tableName),
				FilterExpression:          expr.Filter(),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				Select:                    aws.String(dynamodb.SelectAllAttributes),
			}
			return query, nil
		}

		if projectionBuilder != nil {
			expr, _ := expression.NewBuilder().WithFilter(*filter).WithProjection(*projectionBuilder).Build()
			query = dynamodb.ScanInput{
				TableName:                 aws.String(tableName),
				ProjectionExpression:      expr.Projection(),
				FilterExpression:          expr.Filter(),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				Select:                    aws.String(dynamodb.SelectSpecificAttributes),
			}
			return query, nil
		}

	}
	return query, nil
}
func getFieldByBson(modelType reflect.Type, bsonName string) (int, string, string, bool) {
	numField := modelType.NumField()
	for index := 0; index < numField; index++ {
		field := modelType.Field(index)
		tag1, ok1 := field.Tag.Lookup("bson")
		if ok1 && strings.Split(tag1, ",")[0] == bsonName {
			if dynamodbTag, ok := field.Tag.Lookup("dynamodbav"); ok {
				name := strings.Split(dynamodbTag, ",")[0]
				return index, field.Name, name, true
			}
			if jsonTag, ok := field.Tag.Lookup("json"); ok {
				name := strings.Split(jsonTag, ",")[0]
				return index, field.Name, name, true
			}
		}

	}
	return -1, "", "", false
}
