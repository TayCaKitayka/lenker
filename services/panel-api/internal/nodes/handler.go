package nodes

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/audit"
	"github.com/lenker/lenker/services/panel-api/internal/auth"
	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

type Handler struct {
	logger    *slog.Logger
	nodes     storage.NodesRepository
	adminOnly func(http.Handler) http.Handler
	audit     audit.Recorder
}

func NewHandler(logger *slog.Logger, nodes storage.NodesRepository, adminOnly func(http.Handler) http.Handler) *Handler {
	return &Handler{logger: logger, nodes: nodes, adminOnly: adminOnly, audit: audit.NoopRecorder{}}
}

func (h *Handler) WithAudit(recorder audit.Recorder) *Handler {
	if recorder != nil {
		h.audit = recorder
	}
	return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	if h.adminOnly != nil {
		mux.Handle("POST /api/v1/nodes/bootstrap-token", h.adminOnly(http.HandlerFunc(h.CreateBootstrapToken)))
	}
	mux.HandleFunc("POST /api/v1/nodes/register", h.Register)
	mux.HandleFunc("POST /api/v1/nodes/{id}/heartbeat", h.Heartbeat)
}

func (h *Handler) CreateBootstrapToken(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name             string `json:"name"`
		Region           string `json:"region"`
		CountryCode      string `json:"country_code"`
		Hostname         string `json:"hostname"`
		ExpiresInMinutes int    `json:"expires_in_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}

	expiresIn := request.ExpiresInMinutes
	if expiresIn == 0 {
		expiresIn = 30
	}
	if expiresIn < 1 || expiresIn > 10080 {
		httpapi.WriteError(w, http.StatusBadRequest, "validation_error", "expires_in_minutes must be between 1 and 10080")
		return
	}

	admin, _ := auth.AdminFromContext(r.Context())
	token, err := h.nodes.CreateBootstrapToken(r.Context(), storage.CreateBootstrapTokenInput{
		Name:             strings.TrimSpace(request.Name),
		Region:           strings.TrimSpace(request.Region),
		CountryCode:      strings.ToUpper(strings.TrimSpace(request.CountryCode)),
		Hostname:         strings.TrimSpace(request.Hostname),
		ExpiresAt:        time.Now().UTC().Add(time.Duration(expiresIn) * time.Minute),
		CreatedByAdminID: admin.ID,
	})
	if err != nil {
		h.recordAdmin(r, audit.ActionNodeBootstrapToken, "", audit.OutcomeFailure, "storage_error")
		if h.logger != nil {
			h.logger.Error("node bootstrap token creation failed", "error", err)
		}
		httpapi.WriteStorageError(w)
		return
	}

	h.recordAdmin(r, audit.ActionNodeBootstrapToken, token.NodeID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusCreated, httpapi.Response{Data: map[string]any{
		"id":              token.ID,
		"node_id":         token.NodeID,
		"bootstrap_token": token.Token,
		"expires_at":      token.ExpiresAt,
	}})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request struct {
		NodeID         string `json:"node_id"`
		BootstrapToken string `json:"bootstrap_token"`
		AgentVersion   string `json:"agent_version"`
		Hostname       string `json:"hostname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}

	input := storage.RegisterNodeInput{
		NodeID:         strings.TrimSpace(request.NodeID),
		BootstrapToken: strings.TrimSpace(request.BootstrapToken),
		AgentVersion:   strings.TrimSpace(request.AgentVersion),
		Hostname:       strings.TrimSpace(request.Hostname),
	}
	if input.BootstrapToken == "" {
		httpapi.WriteBadRequest(w, "bootstrap_token is required")
		return
	}
	if input.AgentVersion == "" {
		httpapi.WriteBadRequest(w, "agent_version is required")
		return
	}

	result, err := h.nodes.Register(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidBootstrapToken):
			h.recordNode(r, audit.ActionNodeRegister, input.NodeID, audit.OutcomeFailure, "invalid_bootstrap_token")
			httpapi.WriteError(w, http.StatusUnauthorized, "invalid_bootstrap_token", "bootstrap token is invalid")
			return
		case errors.Is(err, storage.ErrExpiredBootstrapToken):
			h.recordNode(r, audit.ActionNodeRegister, input.NodeID, audit.OutcomeFailure, "expired_bootstrap_token")
			httpapi.WriteError(w, http.StatusUnauthorized, "expired_bootstrap_token", "bootstrap token is expired")
			return
		case errors.Is(err, storage.ErrBootstrapTokenUsed):
			h.recordNode(r, audit.ActionNodeRegister, input.NodeID, audit.OutcomeFailure, "bootstrap_token_used")
			httpapi.WriteError(w, http.StatusUnauthorized, "bootstrap_token_used", "bootstrap token was already used")
			return
		}
		if h.logger != nil {
			h.logger.Error("node registration failed", "error", err)
		}
		h.recordNode(r, audit.ActionNodeRegister, input.NodeID, audit.OutcomeFailure, "internal_error")
		httpapi.WriteStorageError(w)
		return
	}

	h.recordNode(r, audit.ActionNodeRegister, result.Node.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusCreated, httpapi.Response{Data: map[string]any{
		"node_id":       result.Node.ID,
		"node_token":    result.NodeToken,
		"status":        result.Node.Status,
		"drain_state":   result.Node.DrainState,
		"registered_at": result.Node.RegisteredAt,
	}})
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	nodeToken, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		httpapi.WriteUnauthorized(w)
		return
	}

	var request struct {
		NodeID         string    `json:"node_id"`
		AgentVersion   string    `json:"agent_version"`
		Status         string    `json:"status"`
		ActiveRevision int       `json:"active_revision"`
		SentAt         time.Time `json:"sent_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteBadRequest(w, "invalid JSON request body")
		return
	}

	nodeID := strings.TrimSpace(r.PathValue("id"))
	if nodeID == "" {
		httpapi.WriteBadRequest(w, "node id is required")
		return
	}
	if strings.TrimSpace(request.NodeID) != "" && strings.TrimSpace(request.NodeID) != nodeID {
		httpapi.WriteBadRequest(w, "node_id must match path id")
		return
	}
	status := strings.TrimSpace(request.Status)
	if !validNodeStatus(status) {
		httpapi.WriteError(w, http.StatusBadRequest, "validation_error", "status must be pending, active, unhealthy, drained, or disabled")
		return
	}
	if strings.TrimSpace(request.AgentVersion) == "" {
		httpapi.WriteBadRequest(w, "agent_version is required")
		return
	}
	if request.SentAt.IsZero() {
		request.SentAt = time.Now().UTC()
	}

	node, err := h.nodes.RecordHeartbeat(r.Context(), storage.HeartbeatInput{
		NodeID:         nodeID,
		NodeToken:      nodeToken,
		AgentVersion:   strings.TrimSpace(request.AgentVersion),
		Status:         status,
		ActiveRevision: request.ActiveRevision,
		SentAt:         request.SentAt.UTC(),
	})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.recordNode(r, audit.ActionNodeHeartbeat, nodeID, audit.OutcomeFailure, "not_found")
			httpapi.WriteNotFound(w, "node")
			return
		}
		if errors.Is(err, storage.ErrInvalidNodeStatus) {
			h.recordNode(r, audit.ActionNodeHeartbeat, nodeID, audit.OutcomeFailure, "validation_error")
			httpapi.WriteError(w, http.StatusBadRequest, "validation_error", "status must be pending, active, unhealthy, drained, or disabled")
			return
		}
		if h.logger != nil {
			h.logger.Error("node heartbeat failed", "error", err)
		}
		h.recordNode(r, audit.ActionNodeHeartbeat, nodeID, audit.OutcomeFailure, "internal_error")
		httpapi.WriteStorageError(w)
		return
	}

	h.recordNode(r, audit.ActionNodeHeartbeat, node.ID, audit.OutcomeSuccess, "")
	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: map[string]any{
		"node_id":         node.ID,
		"status":          node.Status,
		"drain_state":     node.DrainState,
		"active_revision": node.ActiveRevision,
		"last_seen_at":    node.LastSeenAt,
	}})
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

func validNodeStatus(status string) bool {
	switch status {
	case "pending", "active", "unhealthy", "drained", "disabled":
		return true
	default:
		return false
	}
}

func (h *Handler) recordAdmin(r *http.Request, action string, resourceID string, outcome string, reason string) {
	admin, _ := auth.AdminFromContext(r.Context())
	_ = h.audit.Record(r.Context(), audit.Event{
		ActorType:    "admin",
		ActorID:      admin.ID,
		Action:       action,
		ResourceType: "node",
		ResourceID:   resourceID,
		Outcome:      outcome,
		Reason:       reason,
	})
}

func (h *Handler) recordNode(r *http.Request, action string, resourceID string, outcome string, reason string) {
	_ = h.audit.Record(r.Context(), audit.Event{
		ActorType:    "node",
		ActorID:      resourceID,
		Action:       action,
		ResourceType: "node",
		ResourceID:   resourceID,
		Outcome:      outcome,
		Reason:       reason,
	})
}
