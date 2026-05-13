package httpapi

import "net/http"

func Healthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, Response{
		Data: map[string]string{
			"status": "ok",
		},
	})
}
