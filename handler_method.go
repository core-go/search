package search

import "net/http"

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
	m := GetSearchModel(searchModel)
	result, isLastPage := BuildResultMap(models, count, m, c.Config)
	if x == -1 {
		succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
	} else if c.quickSearch && x == 1 {
		result1, ok := ResultToCsv(searchModel, m, models, count, isLastPage, c.embedField)
		if ok {
			succeed(w, r, http.StatusOK, result1, c.LogWriter, c.Resource, c.Action)
		} else {
			succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
		}
	} else {
		succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
	}
}
