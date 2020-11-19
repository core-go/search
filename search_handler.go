package search

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type SearchHandler struct {
	searchService             SearchService
	searchModelType           reflect.Type
	LogError                  func(context.Context, string)
	LogWriter                 SearchLogWriter
	Config                    SearchResultConfig
	quickSearch               bool
	isExtendedSearchModelType bool
	Resource                  string
	Action                    string
	embedField                string
	userId                    string

	// Search by GET
	paramIndex            map[string]int
	searchModelParamIndex map[string]int
	searchModelIndex      int
}

const (
	PageSizeDefault    = 10
	MaxPageSizeDefault = 10000
	UserId             = "userId"
	Uid                = "uid"
	Username           = "username"
	Search             = "search"
)

func BuildResourceName(s string) string {
	s2 := strings.ToLower(s)
	s3 := ""
	for i := range s {
		if s2[i] != s[i] {
			s3 += "-" + string(s2[i])
		} else {
			s3 += string(s2[i])
		}
	}
	if string(s3[0]) == "-" || string(s3[0]) == "_" {
		return s3[1:]
	}
	return s3
}
func NewSearchHandler(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), logWriter SearchLogWriter, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(searchService, searchModelType, logError, logWriter, true, options...)
}
func NewJSONSearchHandler(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), logWriter SearchLogWriter, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(searchService, searchModelType, logError, logWriter, true, options...)
}
func NewSearchHandlerWithQuickSearch(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), logWriter SearchLogWriter, quickSearch bool, options ...string) *SearchHandler {
	var resource, action, user string
	if len(options) >= 1 {
		user = options[0]
	} else {
		user = UserId
	}
	if len(options) >= 2 {
		resource = options[1]
	} else {
		name := searchModelType.Name()
		if len(name) >= 3 && strings.HasSuffix(name, "SM") {
			name = name[0 : len(name)-2]
		}
		resource = BuildResourceName(name)
	}
	if len(options) >= 3 {
		action = options[2]
	} else {
		action = Search
	}
	return NewSearchHandlerWithConfig(searchService, searchModelType, logError, nil, logWriter, quickSearch, resource, action, user, "")
}
func NewSearchHandlerWithUserId(searchService SearchService, searchModelType reflect.Type, userId string, logError func(context.Context, string), logWriter SearchLogWriter, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(searchService, searchModelType, userId, logError, logWriter, true, options...)
}
func NewJSONSearchHandlerWithUserId(searchService SearchService, searchModelType reflect.Type, userId string, logError func(context.Context, string), logWriter SearchLogWriter, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(searchService, searchModelType, userId, logError, logWriter, false, options...)
}
func NewSearchHandlerWithUserIdAndQuickSearch(searchService SearchService, searchModelType reflect.Type, userId string, logError func(context.Context, string), logWriter SearchLogWriter, quickSearch bool, options ...string) *SearchHandler {
	var resource, action string
	if len(options) >= 1 {
		resource = options[0]
	} else {
		name := searchModelType.Name()
		if len(name) >= 3 && strings.HasSuffix(name, "SM") {
			name = name[0 : len(name)-2]
		}
		resource = BuildResourceName(name)
	}
	if len(options) >= 2 {
		action = options[1]
	} else {
		action = Search
	}
	return NewSearchHandlerWithConfig(searchService, searchModelType, logError, nil, logWriter, quickSearch, resource, action, userId, "")
}
func NewDefaultSearchHandler(searchService SearchService, searchModelType reflect.Type, resource string, logError func(context.Context, string), userId string, quickSearch bool, logWriter SearchLogWriter) *SearchHandler {
	return NewSearchHandlerWithConfig(searchService, searchModelType, logError, nil, logWriter, quickSearch, resource, Search, userId, "")
}
func NewSearchHandlerWithConfig(searchService SearchService, searchModelType reflect.Type, logError func(context.Context, string), config *SearchResultConfig, logWriter SearchLogWriter, quickSearch bool, resource string, action string, userId string, embedField string) *SearchHandler {
	var c SearchResultConfig
	if len(action) == 0 {
		action = Search
	}
	if config != nil {
		c = *config
	} else {
		c.LastPage = "last"
		c.Results = "results"
		c.Total = "total"
	}
	isExtendedSearchModelType := IsExtendedFromSearchModel(searchModelType)
	if isExtendedSearchModelType == false {
		panic(errors.New(searchModelType.Name() + " isn't SearchModel struct nor extended from SearchModel struct!"))
	}

	paramIndex := BuildParamIndex(searchModelType)
	searchModelParamIndex := BuildParamIndex(reflect.TypeOf(SearchModel{}))
	searchModelIndex := FindSearchModelIndex(searchModelType)

	return &SearchHandler{searchService: searchService, searchModelType: searchModelType, Config: c, LogWriter: logWriter, quickSearch: quickSearch, isExtendedSearchModelType: isExtendedSearchModelType, Resource: resource, Action: action, paramIndex: paramIndex, searchModelIndex: searchModelIndex, searchModelParamIndex: searchModelParamIndex, userId: userId, embedField: embedField, LogError: logError}
}

func BuildSearchModel(r *http.Request, searchModelType reflect.Type, isExtendedSearchModelType bool, userIdName string, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) (interface{}, int, error) {
	var searchModel = CreateSearchModelObject(searchModelType, isExtendedSearchModelType)
	method := r.Method
	x := 1
	if method == http.MethodGet {
		ps := r.URL.Query()
		fs := ps.Get("fields")
		if len(fs) == 0 {
			x = -1
		}
		MapParamsToSearchModel(searchModel, ps, searchModelParamIndex, searchModelIndex, paramIndex)
	} else if method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&searchModel); err != nil {
			return nil, x, err
		}
	}
	userId := ""
	if len(userId) == 0 {
		u := r.Context().Value(userIdName)
		if u != nil {
			u2, ok2 := u.(string)
			if ok2 {
				userId = u2
			}
		}
	}
	ProcessSearchModel(searchModel, userId)
	return searchModel, x, nil
}
func BuildResultMap(models interface{}, count int64, pageIndex int64, pageSize int64, firstPageSize int64, config SearchResultConfig) (map[string]interface{}, bool) {
	result := make(map[string]interface{})
	isLastPage := IsLastPage(models, count, pageIndex, pageSize, firstPageSize)

	result[config.Total] = count
	if isLastPage {
		result[config.LastPage] = isLastPage
	}
	result[config.Results] = models
	return result, isLastPage
}
func ResultToCsv(fields []string, models interface{}, count int64, isLastPage bool, embedField string) (string, bool) {
	if len(fields) > 0 {
		result1 := ToCsv(fields, models, count, isLastPage, embedField)
		return result1, true
	} else {
		return "", false
	}
}
func BuildParamIndex(searchModelType reflect.Type) map[string]int {
	params := map[string]int{}
	numField := searchModelType.NumField()
	for i := 0; i < numField; i++ {
		field := searchModelType.Field(i)
		fullJsonTag := field.Tag.Get("json")
		tagDetails := strings.Split(fullJsonTag, ",")
		if len(tagDetails) > 0 && len(tagDetails[0]) > 0 {
			params[tagDetails[0]] = i
		}
	}
	return params
}

func FindSearchModelIndex(searchModelType reflect.Type) int {
	numField := searchModelType.NumField()
	for i := 0; i < numField; i++ {
		if searchModelType.Field(i).Type == reflect.TypeOf(&SearchModel{}) {
			return i
		}
	}
	return -1
}

// Check valid and change value of pagination to correct
func RepairSearchModel(searchModel *SearchModel, currentUserId string) {
	searchModel.CurrentUserId = currentUserId

	pageSize := searchModel.PageSize
	if pageSize > MaxPageSizeDefault || pageSize < 1 {
		pageSize = PageSizeDefault
	}
	pageIndex := searchModel.PageIndex
	if searchModel.PageIndex < 1 {
		pageIndex = 1
	}

	if searchModel.PageSize != pageSize {
		searchModel.PageSize = pageSize
	}

	if searchModel.PageIndex != pageIndex {
		searchModel.PageIndex = pageIndex
	}
}

func ProcessSearchModel(sm interface{}, currentUserId string) {
	if s, ok := sm.(*SearchModel); ok { // Is SearchModel struct
		RepairSearchModel(s, currentUserId)
	} else { // Is extended from SearchModel struct
		value := reflect.Indirect(reflect.ValueOf(sm))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find SearchModel field of extended struct
			if s, ok := value.Field(i).Interface().(*SearchModel); ok {
				RepairSearchModel(s, currentUserId)
				break
			}
		}
	}
}
func CreateSearchModelObject(searchModelType reflect.Type, isExtendedSearchModelType bool) interface{} {
	var searchModel = reflect.New(searchModelType).Interface()
	if isExtendedSearchModelType {
		value := reflect.Indirect(reflect.ValueOf(searchModel))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			// Find SearchModel field of extended struct
			if _, ok := value.Field(i).Interface().(*SearchModel); ok {
				// Init SearchModel to avoid nil value
				value.Field(i).Set(reflect.ValueOf(&SearchModel{}))
				break
			}
		}
	}
	return searchModel
}
func MapParamsToSearchModel(searchModel interface{}, params url.Values, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) interface{} {
	value := reflect.Indirect(reflect.ValueOf(searchModel))
	if value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	for paramKey, valueArr := range params {
		paramValue := ""
		if len(valueArr) > 0 {
			paramValue = valueArr[0]
		}
		if err, field := FindField(value, paramKey, searchModelParamIndex, searchModelIndex, paramIndex); err == nil {
			kind := field.Kind()

			var v interface{}
			// Need handle more case of kind
			if kind == reflect.Int {
				v, _ = strconv.Atoi(paramValue)
			} else if kind == reflect.Int64 {
				v, _ = strconv.ParseInt(paramValue, 10, 64)
			} else if kind == reflect.String {
				v = paramValue
			} else if kind == reflect.Slice {
				sliceKind := reflect.TypeOf(field.Interface()).Elem().Kind()
				if sliceKind == reflect.String {
					v = strings.Split(paramValue, ",")
				} else {
					log.Println("Unhandled slice kind:", kind)
					continue
				}
			} else if kind == reflect.Struct {
				newModel := reflect.New(reflect.Indirect(field).Type()).Interface()
				if errDecode := json.Unmarshal([]byte(paramValue), newModel); errDecode != nil {
					panic(errDecode)
				}
				v = newModel
			} else {
				log.Println("Unhandled kind:", kind)
				continue
			}
			field.Set(reflect.Indirect(reflect.ValueOf(v)))
		} else {
			log.Println(err)
		}
	}
	return searchModel
}

func FindField(value reflect.Value, paramKey string, searchModelParamIndex map[string]int, searchModelIndex int, paramIndex map[string]int) (error, reflect.Value) {
	if index, ok := searchModelParamIndex[paramKey]; ok {
		searchModelField := value.Field(searchModelIndex)
		if searchModelField.Kind() == reflect.Ptr {
			searchModelField = reflect.Indirect(searchModelField)
		}
		return nil, searchModelField.Field(index)
	} else if index, ok := paramIndex[paramKey]; ok {
		return nil, value.Field(index)
	}
	return errors.New("can't find field " + paramKey), value
}
