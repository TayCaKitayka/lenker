package plans

import (
	"log/slog"
	"net/http"

	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

type Handler struct {
	logger *slog.Logger
	plans  storage.PlansRepository
}

func NewHandler(logger *slog.Logger, plans storage.PlansRepository) *Handler {
	return &Handler{logger: logger, plans: plans}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/plans", h.List)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	plans, err := h.plans.List(r.Context())
	if err != nil {
		httpapi.WriteStorageError(w)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: plans})
}
