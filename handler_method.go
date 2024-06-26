package search

import (
	"net/http"
	"reflect"
)

const internalServerError = "Internal Server Error"

func (c *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	filter, x, er0 := BuildFilter(r, c.filterType, c.ParamIndex, c.userId, c.FilterIndex)
	if er0 != nil {
		http.Error(w, "cannot decode filter: "+er0.Error(), http.StatusBadRequest)
		return
	}
	limit, offset, fs, _, _, er1 := Extract(filter)
	if er1 != nil {
		RespondError(w, r, http.StatusInternalServerError, internalServerError, c.LogError, c.ResourceName, c.Activity, er1, c.WriteLog)
		return
	}
	modelsType := reflect.Zero(reflect.SliceOf(c.modelType)).Type()
	models := reflect.New(modelsType).Interface()
	count, er2 := c.Find(r.Context(), filter, models, limit, offset)
	if er2 != nil {
		RespondError(w, r, http.StatusInternalServerError, internalServerError, c.LogError, c.ResourceName, c.Activity, er2, c.WriteLog)
		return
	}

	result := BuildResultMap(models, count, c.List, c.Total)
	if x == -1 {
		succeed(w, r, http.StatusOK, result, c.WriteLog, c.ResourceName, c.Activity)
	} else if c.CSV && x == 1 {
		result1, ok := ResultToCsv(fs, models, count, c.embedField, c.JsonMap, c.SecondaryJsonMap)
		if ok {
			succeed(w, r, http.StatusOK, result1, c.WriteLog, c.ResourceName, c.Activity)
		} else {
			succeed(w, r, http.StatusOK, result, c.WriteLog, c.ResourceName, c.Activity)
		}
	} else {
		succeed(w, r, http.StatusOK, result, c.WriteLog, c.ResourceName, c.Activity)
	}
}

func (c *NextSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	filter, x, er0 := BuildFilter(r, c.filterType, c.ParamIndex, c.userId, c.FilterIndex)
	if er0 != nil {
		http.Error(w, "cannot decode filter: "+er0.Error(), http.StatusBadRequest)
		return
	}
	limit, _, fs, _, nextPageToken, er1 := Extract(filter)
	if er1 != nil {
		RespondError(w, r, http.StatusInternalServerError, internalServerError, c.LogError, c.ResourceName, c.Activity, er1, c.WriteLog)
		return
	}
	modelsType := reflect.Zero(reflect.SliceOf(c.modelType)).Type()
	models := reflect.New(modelsType).Interface()
	nx, er2 := c.Find(r.Context(), filter, models, limit, nextPageToken)
	if er2 != nil {
		RespondError(w, r, http.StatusInternalServerError, internalServerError, c.LogError, c.ResourceName, c.Activity, er2, c.WriteLog)
		return
	}

	result := BuildNextResultMap(models, nx, c.List, c.Next)
	if x == -1 {
		succeed(w, r, http.StatusOK, result, c.WriteLog, c.ResourceName, c.Activity)
	} else if c.CSV && x == 1 {
		result1, ok := ResultToNextCsv(fs, models, nx, c.embedField, c.JsonMap, c.SecondaryJsonMap)
		if ok {
			succeed(w, r, http.StatusOK, result1, c.WriteLog, c.ResourceName, c.Activity)
		} else {
			succeed(w, r, http.StatusOK, result, c.WriteLog, c.ResourceName, c.Activity)
		}
	} else {
		succeed(w, r, http.StatusOK, result, c.WriteLog, c.ResourceName, c.Activity)
	}
}
