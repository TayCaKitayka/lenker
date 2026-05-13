package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
)

type Handler struct {
	logger  *slog.Logger
	service *Service
}

func NewHandler(logger *slog.Logger, service *Service) *Handler {
	return &Handler{logger: logger, service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/admin/login", h.AdminLogin)
}

func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON request body")
		return
	}

	result, err := h.service.Login(r.Context(), LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			httpapi.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "invalid admin credentials")
		case errors.Is(err, ErrInactiveAdmin):
			httpapi.WriteError(w, http.StatusForbidden, "inactive_admin", "admin account is inactive")
		default:
			h.logger.Error("admin login failed", "error", err)
			httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		}
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: result})
}
