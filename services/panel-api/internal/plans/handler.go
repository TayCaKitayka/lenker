package plans

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
	plans     storage.PlansRepository
	adminOnly func(http.Handler) http.Handler
	audit     audit.Recorder
}

func NewHandler(logger *slog.Logger, plans storage.PlansRepository, adminOnly func(http.Handler) http.Handler) *Handler {
	return &Handler{logger: logger, plans: plans, adminOnly: adminOnly, audit: audit.NoopRecorder{}}
}

func (h *Handler) WithAudit(recorder audit.Recorder) *Handler {
	if recorder != nil {
		h.audit = recorder
	}
	return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/plans", h.adminOnly(http.HandlerFunc(h.List)))
	mux.Handle("POST /api/v1/plans", h.adminOnly(http.HandlerFunc(h.Create)))
	mux.Handle("GET /api/v1/plans/{id}", h.adminOnly(http.HandlerFunc(h.Get)))
	mux.Handle("PATCH /api/v1/plans/{id}", h.adminOnly(http.HandlerFunc(h.Update)))
	mux.Handle("POST /api/v1/plans/{id}/archive", h.adminOnly(http.HandlerFunc(h.Archive)))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	plans, err := h.plans.List(r.Context())
	if err != nil {
		httpapi.WriteStorageError(w)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: plans})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name              string `json:"name"`
		DurationDays      int    `json:"duration_days"`
		TrafficLimitBytes *int64 `json:"traffic_limit_bytes"`
		DeviceLimit       int    `json:"device_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}
	if strings.TrimSpace(request.Name) == "" {
		httpapi.WriteBadRequest(w, "name is required")
		return
	}
	if request.DurationDays <= 0 {
		httpapi.WriteBadRequest(w, "duration_days must be greater than zero")
		return
	}
	if request.DeviceLimit <= 0 {
		httpapi.WriteBadRequest(w, "device_limit must be greater than zero")
		return
	}
	if request.TrafficLimitBytes != nil && *request.TrafficLimitBytes <= 0 {
		httpapi.WriteBadRequest(w, "traffic_limit_bytes must be greater than zero when provided")
		return
	}

	plan, err := h.plans.Create(r.Context(), storage.CreatePlanInput{
		Name:              strings.TrimSpace(request.Name),
		DurationDays:      request.DurationDays,
		TrafficLimitBytes: request.TrafficLimitBytes,
		DeviceLimit:       request.DeviceLimit,
	})
	if err != nil {
		h.record(r, audit.ActionPlanCreate, "", audit.OutcomeFailure, "storage_error")
		httpapi.WriteStorageError(w)
		return
	}

	h.record(r, audit.ActionPlanCreate, plan.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusCreated, httpapi.Response{Data: plan})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	plan, err := h.plans.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeResourceError(w, err, "plan")
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: plan})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name              *string `json:"name"`
		DurationDays      *int    `json:"duration_days"`
		TrafficLimitBytes *int64  `json:"traffic_limit_bytes"`
		ClearTrafficLimit bool    `json:"clear_traffic_limit"`
		DeviceLimit       *int    `json:"device_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}

	input := storage.UpdatePlanInput{
		TrafficLimitBytes: request.TrafficLimitBytes,
		ClearTrafficLimit: request.ClearTrafficLimit,
		DurationDays:      request.DurationDays,
		DeviceLimit:       request.DeviceLimit,
	}
	if request.Name != nil {
		name := strings.TrimSpace(*request.Name)
		if name == "" {
			httpapi.WriteBadRequest(w, "name cannot be empty")
			return
		}
		input.Name = &name
	}
	if request.DurationDays != nil && *request.DurationDays <= 0 {
		httpapi.WriteBadRequest(w, "duration_days must be greater than zero")
		return
	}
	if request.DeviceLimit != nil && *request.DeviceLimit <= 0 {
		httpapi.WriteBadRequest(w, "device_limit must be greater than zero")
		return
	}
	if request.TrafficLimitBytes != nil && *request.TrafficLimitBytes <= 0 {
		httpapi.WriteBadRequest(w, "traffic_limit_bytes must be greater than zero when provided")
		return
	}

	plan, err := h.plans.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		h.record(r, audit.ActionPlanUpdate, r.PathValue("id"), audit.OutcomeFailure, errorReason(err))
		writeResourceError(w, err, "plan")
		return
	}

	h.record(r, audit.ActionPlanUpdate, plan.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: plan})
}

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	plan, err := h.plans.Archive(r.Context(), r.PathValue("id"))
	if err != nil {
		h.record(r, audit.ActionPlanArchive, r.PathValue("id"), audit.OutcomeFailure, errorReason(err))
		writeResourceError(w, err, "plan")
		return
	}

	h.record(r, audit.ActionPlanArchive, plan.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: plan})
}

func (h *Handler) record(r *http.Request, action string, resourceID string, outcome string, reason string) {
	admin, _ := auth.AdminFromContext(r.Context())
	_ = h.audit.Record(r.Context(), audit.Event{
		ActorType:    "admin",
		ActorID:      admin.ID,
		Action:       action,
		ResourceType: "plan",
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
