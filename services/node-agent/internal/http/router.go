package httpapi

import (
	"net/http"

	"github.com/lenker/lenker/services/node-agent/internal/agent"
)

type RouterDeps struct {
	Agent *agent.Service
}

func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", Healthz)
	mux.HandleFunc("GET /status", Status(deps.Agent))
	return mux
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, Response{Data: map[string]string{"status": "ok"}})
}

func Status(service *agent.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", "agent service is not configured")
			return
		}
		WriteJSON(w, http.StatusOK, Response{Data: service.Status()})
	}
}
