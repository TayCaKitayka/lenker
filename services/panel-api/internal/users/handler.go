package users

import (
	"log/slog"
	"net/http"

	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

type Handler struct {
	logger *slog.Logger
	users  storage.UsersRepository
}

func NewHandler(logger *slog.Logger, users storage.UsersRepository) *Handler {
	return &Handler{logger: logger, users: users}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/users", h.List)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		httpapi.WriteStorageError(w)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: users})
}
