package search

import "reflect"

type SearchModel struct {
	PageIndex     int                      `json:"pageIndex,omitempty" bson:"pageIndex,omitempty" gorm:"column:pageindex"`
	PageSize      int                      `json:"pageSize,omitempty" bson:"pageSize,omitempty" gorm:"column:pagesize"`
	InitPageSize  int                      `json:"initPageSize,omitempty" bson:"initPageSize,omitempty" gorm:"column:initpagesize"`
	Fields        []string                 `json:"fields,omitempty" bson:"fields,omitempty" gorm:"column:fields"`
	SortField     string                   `json:"sortField,omitempty" bson:"sortField,omitempty" gorm:"column:sortfield"`
	SortType      string                   `json:"sortType,omitempty" bson:"sortType,omitempty" gorm:"column:sorttype"`
	CurrentUserId string                   `json:"currentUserId,omitempty" bson:"currentUserId,omitempty" gorm:"column:currentuserid"`
	Keyword       string                   `json:"keyword,omitempty" bson:"keyword,omitempty" gorm:"column:keyword"`
	Excluding     map[string][]interface{} `json:"excluding,omitempty" bson:"excluding,omitempty" gorm:"column:excluding"`
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
