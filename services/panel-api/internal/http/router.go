package httpapi

import (
	"log/slog"
	"net/http"
)

type Handler interface {
	RegisterRoutes(mux *http.ServeMux)
}

type RouterDeps struct {
	Logger        *slog.Logger
	Auth          Handler
	Admins        Handler
	Users         Handler
	Plans         Handler
	Subscriptions Handler
	Nodes         Handler
}

func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", Healthz)

	for _, handler := range []Handler{
		deps.Auth,
		deps.Admins,
		deps.Users,
		deps.Plans,
		deps.Subscriptions,
		deps.Nodes,
	} {
		if handler != nil {
			handler.RegisterRoutes(mux)
		}
	}

	return requestLogger(deps.Logger, withCORS(mux))
}
