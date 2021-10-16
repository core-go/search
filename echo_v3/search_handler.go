package echo

import (
	"context"
	"errors"
	s "github.com/core-go/search"
	"github.com/labstack/echo"
	"net/http"
	"reflect"
	"strings"
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
	if len(options) >= 1 {
		user = options[0]
	} else {
		user = s.UserId
	}
	if len(options) >= 2 {
		resource = options[1]
	} else {
		name := modelType.Name()
		resource = buildResourceName(name)
	}
	if len(options) >= 3 {
		action = options[2]
	} else {
		action = s.Search
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
		resource = buildResourceName(name)
	}
	if len(options) >= 2 {
		action = options[1]
	} else {
		action = s.Search
	}
	return NewSearchHandlerWithConfig(search, modelType, filterType, logError, nil, writeLog, quickSearch, resource, action, userId, "")
}
func NewDefaultSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, resource string, logError func(context.Context, string), userId string, quickSearch bool, writeLog func(context.Context, string, string, bool, string) error) *SearchHandler {
	return NewSearchHandlerWithConfig(search, modelType, filterType, logError, nil, writeLog, quickSearch, resource, s.Search, userId, "")
}
func NewSearchHandlerWithConfig(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, filterType reflect.Type, logError func(context.Context, string), config *s.SearchResultConfig, writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, resource string, action string, userId string, embedField string) *SearchHandler {
	var c s.SearchResultConfig
	if len(action) == 0 {
		action = s.Search
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

	paramIndex := s.BuildParamIndex(filterType)
	filterParamIndex := s.BuildParamIndex(reflect.TypeOf(s.Filter{}))
	filterIndex := s.FindFilterIndex(filterType)

	return &SearchHandler{search: search, modelType: modelType, filterType: filterType, Config: c, Log: writeLog, quickSearch: quickSearch, isExtendedFilter: isExtendedFilter, Resource: resource, Action: action, paramIndex: paramIndex, filterIndex: filterIndex, filterParamIndex: filterParamIndex, userId: userId, embedField: embedField, Error: logError}
}

const internalServerError = "Internal Server Error"

func (c *SearchHandler) Search(ctx echo.Context) error {
	r := ctx.Request()
	filter, x, er0 := s.BuildFilter(r, c.filterType, c.isExtendedFilter, c.userId, c.filterParamIndex, c.filterIndex, c.paramIndex)
	if er0 != nil {
		return ctx.String(http.StatusBadRequest, "cannot parse form: "+"cannot decode filter: "+er0.Error())
	}
	limit, offset, fs, _, _, er1 := s.Extract(filter)
	if er1 != nil {
		return respondError(ctx, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er1, c.Log)
	}
	modelsType := reflect.Zero(reflect.SliceOf(c.modelType)).Type()
	models := reflect.New(modelsType).Interface()
	count, nextPageToken, er2 := c.search(r.Context(), filter, models, limit, offset)
	if er2 != nil {
		return respondError(ctx, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er2, c.Log)
	}

	result := s.BuildResultMap(models, count, nextPageToken, c.Config)
	if x == -1 {
		return succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
	} else if c.quickSearch && x == 1 {
		result1, ok := s.ResultToCsv(fs, models, count, nextPageToken, c.embedField)
		if ok {
			return succeed(ctx, http.StatusOK, result1, c.Log, c.Resource, c.Action)
		} else {
			return succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
		}
	} else {
		return succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
	}
}

func buildResourceName(s string) string {
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
