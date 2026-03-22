package httpapi

import (
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	OK   bool `json:"ok"`
	Data T    `json:"data,omitempty"`
	Err  *Err `json:"error,omitempty"`
}

type Err struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteOK[T any](w http.ResponseWriter, status int, data T) {
	WriteJSON(w, status, Response[T]{OK: true, Data: data})
}

func WriteError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	WriteJSON(w, status, Response[any]{OK: false, Err: &Err{Code: code, Message: message, Details: details}})
}
