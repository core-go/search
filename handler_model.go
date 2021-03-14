package search

import (
	"errors"
	"reflect"
)

func SetUserId(sm interface{}, currentUserId string) {
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
func CreateSearchModel(searchModelType reflect.Type, isExtendedSearchModelType bool) interface{} {
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

	if searchModel.PageIndex != 0 && searchModel.Page == 0 {
		searchModel.Page = searchModel.PageIndex
	}
	if searchModel.PageSize != 0 && searchModel.Limit == 0 {
		searchModel.Limit = searchModel.PageSize
	}
	if searchModel.FirstPageSize !=0 && searchModel.FirstLimit == 0 {
		searchModel.FirstLimit = searchModel.FirstPageSize
	}

	pageSize := searchModel.Limit
	if pageSize > MaxPageSizeDefault {
		pageSize = PageSizeDefault
	}

	pageIndex := searchModel.Page
	if searchModel.Page < 1 {
		pageIndex = 1
	}

	if searchModel.Limit != pageSize {
		searchModel.Limit = pageSize
	}

	if searchModel.Page != pageIndex {
		searchModel.Page = pageIndex
	}
}

func ExtractFullSearch(m interface{}) (int64, int64, int64, []string, error) {
	if sModel, ok := m.(*SearchModel); ok {
		return sModel.Page, sModel.Limit, sModel.FirstLimit, sModel.Fields, nil
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*SearchModel); ok {
				return sModel1.Page, sModel1.Limit, sModel1.FirstLimit, sModel1.Fields, nil
			}
		}
		return 0, 0, 0, nil, errors.New("cannot extract sort, pageIndex, pageSize, firstPageSize from model")
	}
}
