package search

import (
	"context"
	"errors"
	"reflect"
	"strings"
)

type SearchHandler struct {
	search                    func(ctx context.Context, searchModel interface{}, results interface{}, pageIndex int64, pageSize int64, options ...int64) (int64, error)
	modelType                 reflect.Type
	searchModelType           reflect.Type
	Error                     func(context.Context, string)
	Config                    SearchResultConfig
	quickSearch               bool
	isExtendedSearchModelType bool
	Log                       func(ctx context.Context, resource string, action string, success bool, desc string) error
	Resource                  string
	Action                    string
	embedField                string
	userId                    string

	// search by GET
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

func NewSearchHandler(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error), modelType reflect.Type, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(search, modelType, searchModelType, logError, writeLog, true, options...)
}
func NewJSONSearchHandler(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error), modelType reflect.Type, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(search, modelType, searchModelType, logError, writeLog, false, options...)
}
func NewSearchHandlerWithQuickSearch(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error), modelType reflect.Type, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
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
		resource = buildResourceName(name)
	}
	if len(options) >= 3 {
		action = options[2]
	} else {
		action = Search
	}
	return NewSearchHandlerWithConfig(search, modelType, searchModelType, logError, nil, writeLog, quickSearch, resource, action, user, "")
}
func NewSearchHandlerWithUserId(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error), modelType reflect.Type, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(search, modelType, searchModelType, userId, logError, writeLog, true, options...)
}
func NewJSONSearchHandlerWithUserId(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error), modelType reflect.Type, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(search, modelType, searchModelType, userId, logError, writeLog, false, options...)
}
func NewSearchHandlerWithUserIdAndQuickSearch(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error), modelType reflect.Type, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
	var resource, action string
	if len(options) >= 1 {
		resource = options[0]
	} else {
		name := searchModelType.Name()
		if len(name) >= 3 && strings.HasSuffix(name, "SM") {
			name = name[0 : len(name)-2]
		}
		resource = buildResourceName(name)
	}
	if len(options) >= 2 {
		action = options[1]
	} else {
		action = Search
	}
	return NewSearchHandlerWithConfig(search, modelType, searchModelType, logError, nil, writeLog, quickSearch, resource, action, userId, "")
}
func NewDefaultSearchHandler(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error), modelType reflect.Type, searchModelType reflect.Type, resource string, logError func(context.Context, string), userId string, quickSearch bool, writeLog func(context.Context, string, string, bool, string) error) *SearchHandler {
	return NewSearchHandlerWithConfig(search, modelType, searchModelType, logError, nil, writeLog, quickSearch, resource, Search, userId, "")
}
func NewSearchHandlerWithConfig(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error), modelType reflect.Type, searchModelType reflect.Type, logError func(context.Context, string), config *SearchResultConfig, writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, resource string, action string, userId string, embedField string) *SearchHandler {
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

	return &SearchHandler{search: search, modelType: modelType, searchModelType: searchModelType, Config: c, Log: writeLog, quickSearch: quickSearch, isExtendedSearchModelType: isExtendedSearchModelType, Resource: resource, Action: action, paramIndex: paramIndex, searchModelIndex: searchModelIndex, searchModelParamIndex: searchModelParamIndex, userId: userId, embedField: embedField, Error: logError}
}
