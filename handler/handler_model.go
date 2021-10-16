package handler

import (
	. "github.com/core-go/search"
	"reflect"
)

func BuildResultMap(models interface{}, count int64, nextPageToken string, config SearchResultConfig) map[string]interface{} {
	result := make(map[string]interface{})

	result[config.Total] = count
	result[config.Results] = models
	if len(nextPageToken) > 0 {
		result[config.NextPageToken] = nextPageToken
	}
	return result
}
func SetUserId(sm interface{}, currentUserId string) {
	if s, ok := sm.(*Filter); ok { // Is Filter struct
		RepairFilter(s, currentUserId)
	} else { // Is extended from Filter struct
		value := reflect.Indirect(reflect.ValueOf(sm))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find Filter field of extended struct
			if s, ok := value.Field(i).Interface().(*Filter); ok {
				RepairFilter(s, currentUserId)
				break
			}
		}
	}
}
func CreateFilter(filterType reflect.Type, isExtendedFilter bool) interface{} {
	var filter = reflect.New(filterType).Interface()
	if isExtendedFilter {
		value := reflect.Indirect(reflect.ValueOf(filter))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find Filter field of extended struct
			if _, ok := value.Field(i).Interface().(*Filter); ok {
				// Init Filter to avoid nil value
				value.Field(i).Set(reflect.ValueOf(&Filter{}))
				break
			}
		}
	}
	return filter
}

func FindFilterIndex(filterType reflect.Type) int {
	numField := filterType.NumField()
	for i := 0; i < numField; i++ {
		if filterType.Field(i).Type == reflect.TypeOf(&Filter{}) {
			return i
		}
	}
	return -1
}

// Check valid and change value of pagination to correct
func RepairFilter(filter *Filter, currentUserId string) {
	filter.CurrentUserId = currentUserId

	if filter.PageIndex != 0 && filter.Page == 0 {
		filter.Page = filter.PageIndex
	}
	if filter.PageSize != 0 && filter.Limit == 0 {
		filter.Limit = filter.PageSize
	}
	if filter.FirstPageSize != 0 && filter.FirstLimit == 0 {
		filter.FirstLimit = filter.FirstPageSize
	}

	pageSize := filter.Limit
	if pageSize > MaxPageSizeDefault {
		pageSize = PageSizeDefault
	}

	pageIndex := filter.Page
	if filter.Page < 1 {
		pageIndex = 1
	}

	if filter.Limit != pageSize {
		filter.Limit = pageSize
	}

	if filter.Page != pageIndex {
		filter.Page = pageIndex
	}
}
