package search

import (
	"reflect"
	"strconv"
	"strings"
)

func ToCsv(fields []string, r interface{}, total int64, last bool, embedField string) (out string) {
	val := reflect.ValueOf(r)
	models := reflect.Indirect(val)

	if models.Len() == 0 {
		return "0"
	}

	lastPage := ""
	if last {
		lastPage = "1"
	}
	var rows []string
	rows = append(rows, strconv.FormatInt(total, 10)+","+lastPage)
	rows = BuildCsv(rows, fields, models, embedField)
	return strings.Join(rows, "\n")
	return out
}
