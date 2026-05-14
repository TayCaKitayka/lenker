package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lenker/lenker/services/panel-api/internal/audit"
	"github.com/lenker/lenker/services/panel-api/internal/auth"
	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

type Handler struct {
	logger    *slog.Logger
	users     storage.UsersRepository
	adminOnly func(http.Handler) http.Handler
	audit     audit.Recorder
}

func NewHandler(logger *slog.Logger, users storage.UsersRepository, adminOnly func(http.Handler) http.Handler) *Handler {
	return &Handler{logger: logger, users: users, adminOnly: adminOnly, audit: audit.NoopRecorder{}}
}

func (h *Handler) WithAudit(recorder audit.Recorder) *Handler {
	if recorder != nil {
		h.audit = recorder
	}
	return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/users", h.adminOnly(http.HandlerFunc(h.List)))
	mux.Handle("POST /api/v1/users", h.adminOnly(http.HandlerFunc(h.Create)))
	mux.Handle("GET /api/v1/users/{id}", h.adminOnly(http.HandlerFunc(h.Get)))
	mux.Handle("PATCH /api/v1/users/{id}", h.adminOnly(http.HandlerFunc(h.Update)))
	mux.Handle("POST /api/v1/users/{id}/suspend", h.adminOnly(http.HandlerFunc(h.Suspend)))
	mux.Handle("POST /api/v1/users/{id}/activate", h.adminOnly(http.HandlerFunc(h.Activate)))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		httpapi.WriteStorageError(w)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: users})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}

	input := storage.CreateUserInput{
		Email:       strings.TrimSpace(strings.ToLower(request.Email)),
		DisplayName: strings.TrimSpace(request.DisplayName),
	}
	if input.Email == "" || !strings.Contains(input.Email, "@") {
		httpapi.WriteBadRequest(w, "email is required")
		return
	}

	user, err := h.users.Create(r.Context(), input)
	if err != nil {
		h.record(r, audit.ActionUserCreate, "", audit.OutcomeFailure, "storage_error")
		httpapi.WriteStorageError(w)
		return
	}

	h.record(r, audit.ActionUserCreate, user.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusCreated, httpapi.Response{Data: user})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := h.users.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeResourceError(w, err, "user")
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: user})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email       *string `json:"email"`
		DisplayName *string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}

	input := storage.UpdateUserInput{}
	if request.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*request.Email))
		if email == "" || !strings.Contains(email, "@") {
			httpapi.WriteBadRequest(w, "email must be valid when provided")
			return
		}
		input.Email = &email
	}
	if request.DisplayName != nil {
		displayName := strings.TrimSpace(*request.DisplayName)
		input.DisplayName = &displayName
	}

	user, err := h.users.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		h.record(r, audit.ActionUserUpdate, r.PathValue("id"), audit.OutcomeFailure, errorReason(err))
		writeResourceError(w, err, "user")
		return
	}

	h.record(r, audit.ActionUserUpdate, user.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: user})
}

func (h *Handler) Suspend(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, "suspended", audit.ActionUserSuspend)
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, "active", audit.ActionUserActivate)
}

func (h *Handler) setStatus(w http.ResponseWriter, r *http.Request, status string, action string) {
	user, err := h.users.SetStatus(r.Context(), r.PathValue("id"), status)
	if err != nil {
		h.record(r, action, r.PathValue("id"), audit.OutcomeFailure, errorReason(err))
		writeResourceError(w, err, "user")
		return
	}

	h.record(r, action, user.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: user})
}

func (h *Handler) record(r *http.Request, action string, resourceID string, outcome string, reason string) {
	admin, _ := auth.AdminFromContext(r.Context())
	_ = h.audit.Record(r.Context(), audit.Event{
		ActorType:    "admin",
		ActorID:      admin.ID,
		Action:       action,
		ResourceType: "user",
		ResourceID:   resourceID,
		Outcome:      outcome,
		Reason:       reason,
	})
}

func writeResourceError(w http.ResponseWriter, err error, resource string) {
	if errors.Is(err, storage.ErrNotFound) {
		httpapi.WriteNotFound(w, resource)
		return
	}
	httpapi.WriteStorageError(w)
}

func errorReason(err error) string {
	if errors.Is(err, storage.ErrNotFound) {
		return "not_found"
	}
	return "storage_error"
}
