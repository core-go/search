package handler

import (
	"net/http"
	"reflect"
)

const internalServerError = "Internal Server Error"

func (c *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	filter, x, er0 := BuildFilter(r, c.filterType, c.isExtendedFilter, c.userId, c.filterParamIndex, c.filterIndex, c.paramIndex)
	if er0 != nil {
		http.Error(w, "cannot decode filter: "+er0.Error(), http.StatusBadRequest)
		return
	}
	limit, offset, fs, _, _, er1 := Extract(filter)
	if er1 != nil {
		respondError(w, r, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er1, c.Log)
		return
	}
	modelsType := reflect.Zero(reflect.SliceOf(c.modelType)).Type()
	models := reflect.New(modelsType).Interface()
	count, nextPageToken, er2 := c.search(r.Context(), filter, models, limit, offset)
	if er2 != nil {
		respondError(w, r, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er2, c.Log)
		return
	}

	result := BuildResultMap(models, count, nextPageToken, c.Config)
	if x == -1 {
		succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action)
	} else if c.quickSearch && x == 1 {
		result1, ok := ResultToCsv(fs, models, count, nextPageToken, c.embedField)
		if ok {
			succeed(w, r, http.StatusOK, result1, c.Log, c.Resource, c.Action)
		} else {
			succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action)
		}
	} else {
		succeed(w, r, http.StatusOK, result, c.Log, c.Resource, c.Action)
	}
}
