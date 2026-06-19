package response

import (
	"encoding/json"
	"net/http"

	"github.com/mohit838/learn-go-with-project/internal/constants"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type body struct {
	Data  any        `json:"data,omitempty"`
	Error *errorBody `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, body{Data: data})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	write(w, status, body{Error: &errorBody{Code: code, Message: message}})
}

func write(w http.ResponseWriter, status int, payload body) {
	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
