package search

import (
	"errors"
	"reflect"
)

func ProcessSearchModel(sm interface{}, currentUserId string) {
	if s, ok := sm.(*SearchModel); ok { // Is SearchModel struct
		RepairSearchModel(s, currentUserId)
	} else { // Is extended from SearchModel struct
		value := reflect.Indirect(reflect.ValueOf(sm))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find SearchModel field of extended struct
			if s, ok := value.Field(i).Interface().(*SearchModel); ok {
				RepairSearchModel(s, currentUserId)
				break
			}
		}
	}
}
func CreateSearchModelObject(searchModelType reflect.Type, isExtendedSearchModelType bool) interface{} {
	var searchModel = reflect.New(searchModelType).Interface()
	if isExtendedSearchModelType {
		value := reflect.Indirect(reflect.ValueOf(searchModel))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find SearchModel field of extended struct
			if _, ok := value.Field(i).Interface().(*SearchModel); ok {
				// Init SearchModel to avoid nil value
				value.Field(i).Set(reflect.ValueOf(&SearchModel{}))
				break
			}
		}
	}
	return searchModel
}

func FindSearchModelIndex(searchModelType reflect.Type) int {
	numField := searchModelType.NumField()
	for i := 0; i < numField; i++ {
		if searchModelType.Field(i).Type == reflect.TypeOf(&SearchModel{}) {
			return i
		}
	}
	return -1
}

// Check valid and change value of pagination to correct
func RepairSearchModel(searchModel *SearchModel, currentUserId string) {
	searchModel.CurrentUserId = currentUserId

	pageSize := searchModel.PageSize
	if pageSize > MaxPageSizeDefault || pageSize < 1 {
		pageSize = PageSizeDefault
	}
	pageIndex := searchModel.PageIndex
	if searchModel.PageIndex < 1 {
		pageIndex = 1
	}

	if searchModel.PageSize != pageSize {
		searchModel.PageSize = pageSize
	}

	if searchModel.PageIndex != pageIndex {
		searchModel.PageIndex = pageIndex
	}
}

func ExtractFullSearch(m interface{}) (int64, int64, int64, []string, error) {
	if sModel, ok := m.(*SearchModel); ok {
		return sModel.PageIndex, sModel.PageSize, sModel.FirstPageSize, sModel.Fields, nil
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*SearchModel); ok {
				return sModel1.PageIndex, sModel1.PageSize, sModel1.FirstPageSize, sModel1.Fields, nil
			}
		}
		return 0, 0, 0, nil, errors.New("cannot extract sort, pageIndex, pageSize, firstPageSize from model")
	}
}
