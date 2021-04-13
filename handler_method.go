package search

import (
	"net/http"
	"reflect"
)

func (c *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	searchModel, x, er0 := BuildSearchModel(r, c.searchModelType, c.isExtendedSearchModelType, c.userId, c.searchModelParamIndex, c.searchModelIndex, c.paramIndex)
	if er0 != nil {
		http.Error(w, "cannot decode search model: "+er0.Error(), http.StatusBadRequest)
		return
	}
	pageIndex, pageSize, firstPageSize, fs, er1 := ExtractFullSearch(searchModel)
	if er1 != nil {
		respondError(w, r, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er1, c.Log)
		return
	}
	modelsType := reflect.Zero(reflect.SliceOf(c.modelType)).Type()
	models := reflect.New(modelsType).Interface()
	count, er2 := c.search(r.Context(), searchModel, models, pageIndex, pageSize, firstPageSize)
	if er2 != nil {
		respondError(w, r, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er2, c.Log)
		return
	}

	result, isLastPage := BuildResultMap(models, count, pageIndex, pageSize, firstPageSize, c.Config)
	if x == -1 {
		succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action)
	} else if c.quickSearch && x == 1 {
		result1, ok := ResultToCsv(fs, models, count, isLastPage, c.embedField)
		if ok {
			succeed(w, r, http.StatusOK, result1, c.Log, c.Resource, c.Action)
		} else {
			succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action)
		}
	} else {
		succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action)
	}
}
