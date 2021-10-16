package echo

import (
	"context"
	"errors"
	s "github.com/core-go/search"
	h "github.com/core-go/search/handler"
	"github.com/labstack/echo"
	"net/http"
	"reflect"
)
type SearchHandler struct {
	search           func(ctx context.Context, filter interface{}, results interface{}, limit int64, options ...int64) (int64, string, error)
	modelType        reflect.Type
	filterType       reflect.Type
	Error            func(context.Context, string)
	Config           s.SearchResultConfig
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

func NewSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(search, modelType, filterType, logError, writeLog, true, options...)
}
func NewJSONSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(search, modelType, filterType, logError, writeLog, false, options...)
}
func NewSearchHandlerWithQuickSearch(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
	var resource, action, user string
	if len(options) > 0 && len(options[0]) > 0 {
		user = options[0]
	} else {
		user = h.UserId
	}
	if len(options) > 1 && len(options[1]) > 0 {
		resource = options[1]
	} else {
		name := modelType.Name()
		resource = h.BuildResourceName(name)
	}
	if len(options) > 2 && len(options[2]) > 0 {
		action = options[2]
	} else {
		action = h.Search
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
	if len(options) > 0 && len(options[0]) > 0 {
		resource = options[0]
	} else {
		name := modelType.Name()
		resource = h.BuildResourceName(name)
	}
	if len(options) > 1 && len(options[1]) > 0 {
		action = options[1]
	} else {
		action = h.Search
	}
	return NewSearchHandlerWithConfig(search, modelType, filterType, logError, nil, writeLog, quickSearch, resource, action, userId, "")
}
func NewDefaultSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, resource string, logError func(context.Context, string), userId string, quickSearch bool, writeLog func(context.Context, string, string, bool, string) error) *SearchHandler {
	return NewSearchHandlerWithConfig(search, modelType, filterType, logError, nil, writeLog, quickSearch, resource, h.Search, userId, "")
}
func NewSearchHandlerWithConfig(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), config *s.SearchResultConfig, writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, resource string, action string, userId string, embedField string) *SearchHandler {
	var c s.SearchResultConfig
	if len(action) == 0 {
		action = h.Search
	}
	if config != nil {
		c = *config
	} else {
		c.LastPage = "last"
		c.Results = "results"
		c.Total = "total"
	}
	isExtendedFilter := s.IsExtendedFromFilter(filterType)
	if isExtendedFilter == false {
		panic(errors.New(filterType.Name() + " isn't Filter struct nor extended from Filter struct!"))
	}

	paramIndex := h.BuildParamIndex(filterType)
	filterParamIndex := h.BuildParamIndex(reflect.TypeOf(s.Filter{}))
	filterIndex := h.FindFilterIndex(filterType)

	return &SearchHandler{search: search, modelType: modelType, filterType: filterType, Config: c, Log: writeLog, quickSearch: quickSearch, isExtendedFilter: isExtendedFilter, Resource: resource, Action: action, paramIndex: paramIndex, filterIndex: filterIndex, filterParamIndex: filterParamIndex, userId: userId, embedField: embedField, Error: logError}
}

const internalServerError = "Internal Server Error"

func (c *SearchHandler) Search(ctx echo.Context) error {
	r := ctx.Request()
	filter, x, er0 := h.BuildFilter(r, c.filterType, c.isExtendedFilter, c.userId, c.filterParamIndex, c.filterIndex, c.paramIndex)
	if er0 != nil {
		return ctx.String(http.StatusBadRequest, "cannot parse form: "+"cannot decode filter: "+er0.Error())
	}
	limit, offset, fs, _, _, er1 := h.Extract(filter)
	if er1 != nil {
		return respondError(ctx, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er1, c.Log)
	}
	modelsType := reflect.Zero(reflect.SliceOf(c.modelType)).Type()
	models := reflect.New(modelsType).Interface()
	count, nextPageToken, er2 := c.search(r.Context(), filter, models, limit, offset)
	if er2 != nil {
		return respondError(ctx, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er2, c.Log)
	}

	result := h.BuildResultMap(models, count, nextPageToken, c.Config)
	if x == -1 {
		return succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
	} else if c.quickSearch && x == 1 {
		result1, ok := h.ResultToCsv(fs, models, count, nextPageToken, c.embedField)
		if ok {
			return succeed(ctx, http.StatusOK, result1, c.Log, c.Resource, c.Action)
		} else {
			return succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
		}
	} else {
		return succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
	}
}
func respondError(ctx echo.Context, code int, result interface{}, logError func(context.Context, string), resource string, action string, err error, writeLog func(ctx context.Context, resource string, action string, success bool, desc string) error) error {
	if logError != nil {
		logError(ctx.Request().Context(), err.Error())
	}
	return respond(ctx, code, result, writeLog, resource, action, false, err.Error())
}
func respond(ctx echo.Context, code int, result interface{}, writeLog func(ctx context.Context, resource string, action string, success bool, desc string) error, resource string, action string, success bool, desc string) error {
	err := ctx.JSON(code, result)
	if writeLog != nil {
		writeLog(ctx.Request().Context(), resource, action, success, desc)
	}
	return err
}
func succeed(ctx echo.Context, code int, result interface{}, writeLog func(ctx context.Context, resource string, action string, success bool, desc string) error, resource string, action string) error {
	return respond(ctx, code, result, writeLog, resource, action, true, "")
}
