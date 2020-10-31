package search

import (
	"reflect"
	"strconv"
	"strings"
)

func ToCsv(m interface{}, r interface{}, embedField string) (out string) {
	if modelSearch, ok := m.(*SearchModel); ok {
		if result, ok := r.(*SearchResult); ok {
			val := reflect.ValueOf(result.Results)
			models := reflect.Indirect(val)

			if models.Len() == 0 {
				return "0"
			}

			lastPage := ""
			if result.Last {
				lastPage = "1"
			}
			var rows []string
			rows = append(rows, strconv.FormatInt(result.Total, 10)+","+lastPage)
			rows = BuildCsv(rows, modelSearch.Fields, models, embedField)
			return strings.Join(rows, "\n")
		}
	}
	return out
}
