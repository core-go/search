package gin

import (
	"context"
	"errors"
	s "github.com/core-go/search"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strings"
)

type SearchHandler struct {
	search                    func(ctx context.Context, searchModel interface{}, results interface{}, limit int64, options ...int64) (int64, string, error)
	modelType                 reflect.Type
	searchModelType           reflect.Type
	Error                     func(context.Context, string)
	Config                    s.SearchResultConfig
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

func NewSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(search, modelType, searchModelType, logError, writeLog, true, options...)
}
func NewJSONSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithQuickSearch(search, modelType, searchModelType, logError, writeLog, false, options...)
}
func NewSearchHandlerWithQuickSearch(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, searchModelType reflect.Type, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
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
	return NewSearchHandlerWithConfig(search, modelType, searchModelType, logError, nil, writeLog, quickSearch, resource, action, user, "")
}
func NewSearchHandlerWithUserId(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(search, modelType, searchModelType, userId, logError, writeLog, true, options...)
}
func NewJSONSearchHandlerWithUserId(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, options ...string) *SearchHandler {
	return NewSearchHandlerWithUserIdAndQuickSearch(search, modelType, searchModelType, userId, logError, writeLog, false, options...)
}
func NewSearchHandlerWithUserIdAndQuickSearch(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, searchModelType reflect.Type, userId string, logError func(context.Context, string), writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, options ...string) *SearchHandler {
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
	return NewSearchHandlerWithConfig(search, modelType, searchModelType, logError, nil, writeLog, quickSearch, resource, action, userId, "")
}
func NewDefaultSearchHandler(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, searchModelType reflect.Type, resource string, logError func(context.Context, string), userId string, quickSearch bool, writeLog func(context.Context, string, string, bool, string) error) *SearchHandler {
	return NewSearchHandlerWithConfig(search, modelType, searchModelType, logError, nil, writeLog, quickSearch, resource, s.Search, userId, "")
}
func NewSearchHandlerWithConfig(search func(context.Context, interface{}, interface{}, int64, ...int64) (int64, string, error), modelType reflect.Type, searchModelType reflect.Type, logError func(context.Context, string), config *s.SearchResultConfig, writeLog func(context.Context, string, string, bool, string) error, quickSearch bool, resource string, action string, userId string, embedField string) *SearchHandler {
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
	isExtendedSearchModelType := s.IsExtendedFromSearchModel(searchModelType)
	if isExtendedSearchModelType == false {
		panic(errors.New(searchModelType.Name() + " isn't SearchModel struct nor extended from SearchModel struct!"))
	}

	paramIndex := s.BuildParamIndex(searchModelType)
	searchModelParamIndex := s.BuildParamIndex(reflect.TypeOf(s.SearchModel{}))
	searchModelIndex := s.FindSearchModelIndex(searchModelType)

	return &SearchHandler{search: search, modelType: modelType, searchModelType: searchModelType, Config: c, Log: writeLog, quickSearch: quickSearch, isExtendedSearchModelType: isExtendedSearchModelType, Resource: resource, Action: action, paramIndex: paramIndex, searchModelIndex: searchModelIndex, searchModelParamIndex: searchModelParamIndex, userId: userId, embedField: embedField, Error: logError}
}

const internalServerError = "Internal Server Error"

func (c *SearchHandler) Search(ctx *gin.Context) {
	r := ctx.Request
	searchModel, x, er0 := s.BuildSearchModel(r, c.searchModelType, c.isExtendedSearchModelType, c.userId, c.searchModelParamIndex, c.searchModelIndex, c.paramIndex)
	if er0 != nil {
		ctx.String(http.StatusBadRequest, "cannot parse form: "+"cannot decode search model: "+er0.Error())
		return
	}
	limit, offset, fs, _, _, er1 := s.Extract(searchModel)
	if er1 != nil {
		respondError(ctx, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er1, c.Log)
		return
	}
	modelsType := reflect.Zero(reflect.SliceOf(c.modelType)).Type()
	models := reflect.New(modelsType).Interface()
	count, nextPageToken, er2 := c.search(r.Context(), searchModel, models, limit, offset)
	if er2 != nil {
		respondError(ctx, http.StatusInternalServerError, internalServerError, c.Error, c.Resource, "search", er2, c.Log)
		return
	}

	result := s.BuildResultMap(models, count, nextPageToken, c.Config)
	if x == -1 {
		succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
	} else if c.quickSearch && x == 1 {
		result1, ok := s.ResultToCsv(fs, models, count, nextPageToken, c.embedField)
		if ok {
			succeed(ctx, http.StatusOK, result1, c.Log, c.Resource, c.Action)
		} else {
			succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
		}
	} else {
		succeed(ctx, http.StatusOK, result, c.Log, c.Resource, c.Action)
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
func respondError(ctx *gin.Context, code int, result interface{}, logError func(context.Context, string), resource string, action string, err error, writeLog func(ctx context.Context, resource string, action string, success bool, desc string) error) {
	if logError != nil {
		logError(ctx.Request.Context(), err.Error())
	}
	respond(ctx, code, result, writeLog, resource, action, false, err.Error())
}
func respond(ctx *gin.Context, code int, result interface{}, writeLog func(ctx context.Context, resource string, action string, success bool, desc string) error, resource string, action string, success bool, desc string) {
	ctx.JSON(code, result)
	if writeLog != nil {
		writeLog(ctx.Request.Context(), resource, action, success, desc)
	}
}

func succeed(ctx *gin.Context, code int, result interface{}, writeLog func(ctx context.Context, resource string, action string, success bool, desc string) error, resource string, action string) {
	respond(ctx, code, result, writeLog, resource, action, true, "")
}
