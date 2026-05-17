package subscriptions

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
	logger        *slog.Logger
	subscriptions storage.SubscriptionsRepository
	adminOnly     func(http.Handler) http.Handler
	audit         audit.Recorder
}

func NewHandler(logger *slog.Logger, subscriptions storage.SubscriptionsRepository, adminOnly func(http.Handler) http.Handler) *Handler {
	return &Handler{logger: logger, subscriptions: subscriptions, adminOnly: adminOnly, audit: audit.NoopRecorder{}}
}

func (h *Handler) WithAudit(recorder audit.Recorder) *Handler {
	if recorder != nil {
		h.audit = recorder
	}
	return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/subscriptions", h.adminOnly(http.HandlerFunc(h.List)))
	mux.Handle("POST /api/v1/subscriptions", h.adminOnly(http.HandlerFunc(h.Create)))
	mux.Handle("GET /api/v1/subscriptions/{id}", h.adminOnly(http.HandlerFunc(h.Get)))
	mux.Handle("PATCH /api/v1/subscriptions/{id}", h.adminOnly(http.HandlerFunc(h.Update)))
	mux.Handle("POST /api/v1/subscriptions/{id}/renew", h.adminOnly(http.HandlerFunc(h.Renew)))
	mux.Handle("GET /api/v1/subscriptions/{id}/access", h.adminOnly(http.HandlerFunc(h.Access)))
	mux.Handle("POST /api/v1/subscriptions/{id}/access-token", h.adminOnly(http.HandlerFunc(h.CreateAccessToken)))
	mux.HandleFunc("GET /api/v1/client/subscription-access", h.ClientAccess)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.subscriptions.List(r.Context())
	if err != nil {
		httpapi.WriteStorageError(w)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: subscriptions})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID          string  `json:"user_id"`
		PlanID          string  `json:"plan_id"`
		PreferredRegion *string `json:"preferred_region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}
	if strings.TrimSpace(request.UserID) == "" {
		httpapi.WriteBadRequest(w, "user_id is required")
		return
	}
	if strings.TrimSpace(request.PlanID) == "" {
		httpapi.WriteBadRequest(w, "plan_id is required")
		return
	}
	var preferredRegion *string
	if request.PreferredRegion != nil {
		value := strings.TrimSpace(*request.PreferredRegion)
		if value != "" {
			preferredRegion = &value
		}
	}

	subscription, err := h.subscriptions.Create(r.Context(), storage.CreateSubscriptionInput{
		UserID:          strings.TrimSpace(request.UserID),
		PlanID:          strings.TrimSpace(request.PlanID),
		PreferredRegion: preferredRegion,
	})
	if err != nil {
		h.record(r, audit.ActionSubscriptionCreate, "", audit.OutcomeFailure, errorReason(err))
		writeResourceError(w, err, "subscription")
		return
	}

	h.record(r, audit.ActionSubscriptionCreate, subscription.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusCreated, httpapi.Response{Data: subscription})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	subscription, err := h.subscriptions.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeResourceError(w, err, "subscription")
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: subscription})
}

func (h *Handler) Access(w http.ResponseWriter, r *http.Request) {
	access, err := h.subscriptions.Access(r.Context(), r.PathValue("id"))
	if err != nil {
		writeSubscriptionAccessError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: access})
}

func (h *Handler) CreateAccessToken(w http.ResponseWriter, r *http.Request) {
	token, err := h.subscriptions.CreateAccessToken(r.Context(), r.PathValue("id"))
	if err != nil {
		writeSubscriptionAccessError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, httpapi.Response{Data: token})
}

func (h *Handler) ClientAccess(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "subscription access token is missing or invalid")
		return
	}

	access, err := h.subscriptions.AccessByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidSubscriptionAccessToken) {
			httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "subscription access token is missing or invalid")
			return
		}
		writeSubscriptionAccessError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: clientAccessResponse(access)})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Status               *string `json:"status"`
		TrafficLimitBytes    *int64  `json:"traffic_limit_bytes"`
		ClearTrafficLimit    bool    `json:"clear_traffic_limit"`
		DeviceLimit          *int    `json:"device_limit"`
		PreferredRegion      *string `json:"preferred_region"`
		ClearPreferredRegion bool    `json:"clear_preferred_region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}

	input := storage.UpdateSubscriptionInput{
		TrafficLimitBytes:    request.TrafficLimitBytes,
		ClearTrafficLimit:    request.ClearTrafficLimit,
		DeviceLimit:          request.DeviceLimit,
		ClearPreferredRegion: request.ClearPreferredRegion,
	}
	if request.Status != nil {
		status := strings.TrimSpace(*request.Status)
		if status != "active" && status != "expired" && status != "suspended" {
			httpapi.WriteBadRequest(w, "status must be active, expired, or suspended")
			return
		}
		input.Status = &status
	}
	if request.TrafficLimitBytes != nil && *request.TrafficLimitBytes <= 0 {
		httpapi.WriteBadRequest(w, "traffic_limit_bytes must be greater than zero when provided")
		return
	}
	if request.DeviceLimit != nil && *request.DeviceLimit <= 0 {
		httpapi.WriteBadRequest(w, "device_limit must be greater than zero")
		return
	}
	if request.PreferredRegion != nil {
		preferredRegion := strings.TrimSpace(*request.PreferredRegion)
		input.PreferredRegion = &preferredRegion
	}

	subscription, err := h.subscriptions.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		h.record(r, audit.ActionSubscriptionUpdate, r.PathValue("id"), audit.OutcomeFailure, errorReason(err))
		writeResourceError(w, err, "subscription")
		return
	}

	h.record(r, audit.ActionSubscriptionUpdate, subscription.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: subscription})
}

func (h *Handler) Renew(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ExtendDays int `json:"extend_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}
	if request.ExtendDays <= 0 {
		httpapi.WriteBadRequest(w, "extend_days must be greater than zero")
		return
	}

	subscription, err := h.subscriptions.Renew(r.Context(), r.PathValue("id"), request.ExtendDays)
	if err != nil {
		h.record(r, audit.ActionSubscriptionRenew, r.PathValue("id"), audit.OutcomeFailure, errorReason(err))
		writeResourceError(w, err, "subscription")
		return
	}

	h.record(r, audit.ActionSubscriptionRenew, subscription.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: subscription})
}

func (h *Handler) record(r *http.Request, action string, resourceID string, outcome string, reason string) {
	admin, _ := auth.AdminFromContext(r.Context())
	_ = h.audit.Record(r.Context(), audit.Event{
		ActorType:    "admin",
		ActorID:      admin.ID,
		Action:       action,
		ResourceType: "subscription",
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

func writeSubscriptionAccessError(w http.ResponseWriter, err error) {
	if errors.Is(err, storage.ErrNotFound) {
		httpapi.WriteNotFound(w, "subscription")
		return
	}
	if errors.Is(err, storage.ErrSubscriptionAccessUnavailable) {
		httpapi.WriteError(w, http.StatusConflict, "access_unavailable", "subscription access export is unavailable")
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

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

func clientAccessResponse(access storage.SubscriptionAccess) map[string]any {
	return map[string]any{
		"export_kind":     access.ExportKind,
		"subscription_id": access.SubscriptionID,
		"status":          access.Status,
		"protocol":        access.Protocol,
		"protocol_path":   access.ProtocolPath,
		"plan_name":       access.PlanName,
		"node":            access.Node,
		"endpoint":        access.Endpoint,
		"client": map[string]any{
			"id":    access.Client.ID,
			"email": access.Client.Email,
			"flow":  access.Client.Flow,
			"level": access.Client.Level,
		},
		"display_name": access.DisplayName,
		"uri":          access.URI,
	}
}
