package search

import (
	"errors"
	"net/http"
	"reflect"
)

func (c *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	searchModel, x, err := BuildSearchModel(r, c.searchModelType, c.isExtendedSearchModelType, c.userId, c.searchModelParamIndex, c.searchModelIndex, c.paramIndex)
	if err != nil {
		http.Error(w, "cannot decode search model: "+err.Error(), http.StatusBadRequest)
		return
	}
	models, count, err := c.searchService.Search(r.Context(), searchModel)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, InternalServerError, c.LogError, c.Resource, "search", err, c.LogWriter)
		return
	}
	pageIndex, pageSize, firstPageSize, fs, err := ExtractSearch2(searchModel)
	result, isLastPage := BuildResultMap(models, count, pageIndex, pageSize, firstPageSize, c.Config)
	if x == -1 {
		succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
	} else if c.quickSearch && x == 1 {
		result1, ok := ResultToCsv(fs, models, count, isLastPage, c.embedField)
		if ok {
			succeed(w, r, http.StatusOK, result1, c.LogWriter, c.Resource, c.Action)
		} else {
			succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
		}
	} else {
		succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
	}
}
func ExtractSearch2(m interface{}) (int64, int64, int64, []string, error) {
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
