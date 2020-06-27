package search

import "reflect"

type SearchModel struct {
	PageIndex     int64                    `json:"pageIndex,omitempty" gorm:"column:pageindex" bson:"pageIndex,omitempty" dynamodbav:"pageIndex,omitempty" firestore:"pageIndex,omitempty"`
	PageSize      int64                    `json:"pageSize,omitempty" gorm:"column:pagesize" bson:"pageSize,omitempty" dynamodbav:"pageSize,omitempty" firestore:"pageSize,omitempty"`
	InitPageSize  int64                    `json:"initPageSize,omitempty" gorm:"column:initpagesize" bson:"initPageSize,omitempty" dynamodbav:"initpagesize,omitempty" firestore:"initpagesize,omitempty"`
	Fields        []string                 `json:"fields,omitempty" gorm:"column:fields" bson:"fields,omitempty" dynamodbav:"fields,omitempty" firestore:"fields,omitempty"`
	SortField     string                   `json:"sortField,omitempty" gorm:"column:sortfield" bson:"sortField,omitempty" dynamodbav:"sortField,omitempty" firestore:"sortField,omitempty"`
	SortType      string                   `json:"sortType,omitempty" gorm:"column:sorttype" bson:"sortType,omitempty" dynamodbav:"sorttype,omitempty" firestore:"sorttype,omitempty"`
	CurrentUserId string                   `json:"currentUserId,omitempty" gorm:"column:currentuserid" bson:"currentUserId,omitempty" dynamodbav:"currentUserId,omitempty" firestore:"currentUserId,omitempty"`
	Keyword       string                   `json:"keyword,omitempty" gorm:"column:keyword" bson:"keyword,omitempty" dynamodbav:"keyword,omitempty" firestore:"keyword,omitempty"`
	Excluding     map[string][]interface{} `json:"excluding,omitempty" gorm:"column:excluding" bson:"excluding,omitempty" dynamodbav:"excluding,omitempty" firestore:"excluding,omitempty"`
	RefId         string                   `json:"refId,omitempty" gorm:"column:refid" bson:"refId,omitempty" dynamodbav:"refId,omitempty" firestore:"refId,omitempty"`
}

func IsExtendedFromSearchModel(searchModelType reflect.Type) bool {
	var searchModel = reflect.New(searchModelType).Interface()
	if _, ok := searchModel.(*SearchModel); ok {
		return false
	} else {
		value := reflect.Indirect(reflect.ValueOf(searchModel))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if _, ok := value.Field(i).Interface().(*SearchModel); ok {
				return true
			}
		}
	}
	return false
}
