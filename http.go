package search

import (
	"encoding/json"
	"log"
	"net/http"
)
const internalServerError = "Internal Server Error"
func marshal(v interface{}) ([]byte, error) {
	b, ok1 := v.([]byte)
	if ok1 {
		return b, nil
	}
	s, ok2 := v.(string)
	if ok2 {
		return []byte(s), nil
	}
	return json.Marshal(v)
}
func write(w http.ResponseWriter, code int, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if result == nil {
		w.Write([]byte("null"))
		return
	}
	response, err := marshal(result)
	if err != nil {
		log.Println("cannot marshal of result: " + err.Error())
		http.Error(w, internalServerError, http.StatusInternalServerError)
	} else {
		w.Write(response)
	}
}
func writeString(w http.ResponseWriter, code int, v string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(v))
}
