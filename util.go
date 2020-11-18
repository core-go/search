package search

import "reflect"

func IsLastPage(models interface{}, count int64, pageIndex int64, pageSize int64, initPageSize int64) bool {
	lengthModels := int64(reflect.Indirect(reflect.ValueOf(models)).Len())
	var receivedItems int64

	if initPageSize > 0 {
		if pageIndex == 1 {
			receivedItems = initPageSize
		} else if pageIndex > 1 {
			receivedItems = pageSize*(pageIndex-2) + initPageSize + lengthModels
		}
	} else {
		receivedItems = pageSize*(pageIndex-1) + lengthModels
	}
	return receivedItems >= count
}
