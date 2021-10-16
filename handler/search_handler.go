package handler

import (
	"context"
	"errors"
	"reflect"

	. "github.com/core-go/search"
)

type SearchHandler struct {
	search           func(ctx context.Context, filter interface{}, results interface{}, limit int64, options ...int64) (int64, string, error)
	modelType        reflect.Type
	filterType       reflect.Type
	Error            func(context.Context, string)
	Config           SearchResultConfig
	quickSearch      bool
	isExtendedFilter bool
	Log              func(ctx context.Context, resource string, action string, success bool, desc string) error
	Resource         string
	Action           string
	embedField       string
	userId           string

	// search by GET
	paramIndex       map[string]int
	filterParamIndex map[string]int
	filterIndex      int
}

const (
	PageSizeDefault    = 10
	MaxPageSizeDefault = 10000
	UserId             = "userId"
	Uid                = "uid"
	Username           = "username"
	Search             = "search"
)

func NewSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(search, modelType, filterType, logError, writeLog, true, options...)
}
func NewJSONSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(search, modelType, filterType, logError, writeLog, false, options...)
}
func NewSearchHandlerWithQuickSearch(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
	var resource, action, user string
	if len(options) >= 1 {
		user = options[0]
	} else {
		user = UserId
	}
	if len(options) >= 2 {
		resource = options[1]
	} else {
		name := modelType.Name()
		resource = BuildResourceName(name)
	}
	if len(options) >= 3 {
		action = options[2]
	} else {
		action = Search
	}
	return NewSearchHandlerWithConfig(search, modelType, filterType, logError, nil, writeLog, quickSearch, resource, action, user, "")
}
func NewSearchHandlerWithUserId(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(search, modelType, filterType, userId, logError, writeLog, true, options...)
}
func NewJSONSearchHandlerWithUserId(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(search, modelType, filterType, userId, logError, writeLog, false, options...)
}
func NewSearchHandlerWithUserIdAndQuickSearch(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
	var resource, action string
	if len(options) >= 1 {
		resource = options[0]
	} else {
		name := modelType.Name()
		resource = BuildResourceName(name)
	}
	if len(options) >= 2 {
		action = options[1]
	} else {
		action = Search
	}
	return NewSearchHandlerWithConfig(search, modelType, filterType, logError, nil, writeLog, quickSearch, resource, action, userId, "")
}
func NewDefaultSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, resource string, logError func(context.Context, string), userId string, quickSearch bool, writeLog func(context.Context, string, string, bool, string) error) *SearchHandler {
	return NewSearchHandlerWithConfig(search, modelType, filterType, logError, nil, writeLog, quickSearch, resource, Search, userId, "")
}
func NewSearchHandlerWithConfig(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), config *SearchResultConfig, writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, resource string, action string, userId string, embedField string) *SearchHandler {
	var c SearchResultConfig
	if len(action) == 0 {
		action = Search
	}
	if config != nil {
		c = *config
	} else {
		c.LastPage = "last"
		c.Results = "list"
		c.Total = "total"
		c.NextPageToken = "nextPageToken"
	}
	isExtendedFilter := IsExtendedFromFilter(filterType)
	if isExtendedFilter == false {
		panic(errors.New(filterType.Name() + " isn't Filter struct nor extended from Filter struct!"))
	}

	paramIndex := BuildParamIndex(filterType)
	filterParamIndex := BuildParamIndex(reflect.TypeOf(Filter{}))
	filterIndex := FindFilterIndex(filterType)

	return &SearchHandler{search: search, modelType: modelType, filterType: filterType, Config: c, Log: writeLog, quickSearch: quickSearch, isExtendedFilter: isExtendedFilter, Resource: resource, Action: action, paramIndex: paramIndex, filterIndex: filterIndex, filterParamIndex: filterParamIndex, userId: userId, embedField: embedField, Error: logError}
}
