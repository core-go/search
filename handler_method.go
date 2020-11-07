package search

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func (c *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	var searchModel = CreateSearchModelObject(c.searchModelType, c.isExtendedSearchModelType)

	method := r.Method
	x := 1
	if method == http.MethodGet {
		ps := r.URL.Query()
		fs := ps.Get("fields")
		if len(fs) == 0 {
			x = -1
		}
		MapParamsToSearchModel(searchModel, ps, c.searchModelParamIndex, c.searchModelIndex, c.paramIndex)
	} else if method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&searchModel); err != nil {
			http.Error(w, "cannot decode search model: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	userId := ""
	if len(c.userId) == 0 {
		u := r.Context().Value(c.userId)
		if u != nil {
			u2, ok2 := u.(string)
			if ok2 {
				userId = u2
			}
		}
	}
	ProcessSearchModel(searchModel, userId)

	models, count, err := c.searchService.Search(r.Context(), searchModel)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, InternalServerError, c.LogError, c.Resource, "search", err, c.LogWriter)
	} else {
		result := make(map[string]interface{})
		m := GetSearchModel(searchModel)
		isLastPage := IsLastPage(models, count, m.PageIndex, m.PageSize, m.FirstPageSize)
		if isLastPage {
			result[c.Config.LastPage] = isLastPage
		}
		result[c.Config.Results] = models
		result[c.Config.Total] = count
		if x == -1 {
			succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
		} else if c.quickSearch && x == 1 {
			value := reflect.Indirect(reflect.ValueOf(searchModel))
			numField := value.NumField()
			for i := 0; i < numField; i++ {
				field := value.Field(i)
				interfaceOfField := field.Interface()
				if v, ok := interfaceOfField.(*SearchModel); ok {
					if len(v.Fields) > 0 {
						result1 := ToCsv(*m, models, count, isLastPage, c.embedField)
						succeed(w, r, http.StatusOK, result1, c.LogWriter, c.Resource, c.Action)
						return
					}
				}
			}
			succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
		} else {
			succeed(w, r, http.StatusOK, result, c.LogWriter, c.Resource, c.Action)
		}
	}
}
