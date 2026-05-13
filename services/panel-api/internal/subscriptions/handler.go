package subscriptions

import (
	"log/slog"
	"net/http"

	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

type Handler struct {
	logger        *slog.Logger
	subscriptions storage.SubscriptionsRepository
}

func NewHandler(logger *slog.Logger, subscriptions storage.SubscriptionsRepository) *Handler {
	return &Handler{logger: logger, subscriptions: subscriptions}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/subscriptions", h.List)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.subscriptions.List(r.Context())
	if err != nil {
		httpapi.WriteStorageError(w)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: subscriptions})
}
