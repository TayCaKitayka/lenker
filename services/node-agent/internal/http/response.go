package httpapi

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func WriteError(w http.ResponseWriter, status int, code string, message string) {
	WriteJSON(w, status, Response{
		Error: &APIError{Code: code, Message: message},
	})
}
