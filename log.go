package search

import (
	"context"
	"encoding/json"
	"net/http"
)
const InternalServerError = "Internal Server Error"
type SearchLogWriter interface {
	Write(ctx context.Context, resource string, action string, success bool, desc string) error
}
func respondError(w http.ResponseWriter, r *http.Request, code int, result interface{}, logError func(context.Context, string), resource string, action string, err error, logWriter SearchLogWriter) {
	if logError != nil {
		logError(r.Context(), err.Error())
	}
	respond(w, r, code, result, logWriter, resource, action, false, err.Error())
}
func respond(w http.ResponseWriter, r *http.Request, code int, result interface{}, logWriter SearchLogWriter, resource string, action string, success bool, desc string) {
	response, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	if logWriter != nil {
		logWriter.Write(r.Context(), resource, action, success, desc)
	}
}

func succeed(w http.ResponseWriter, r *http.Request, code int, result interface{}, logWriter SearchLogWriter, resource string, action string) {
	respond(w, r, code, result, logWriter, resource, action, true, "")
}
