package nodes

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

type Handler struct {
	logger *slog.Logger
	nodes  storage.NodesRepository
}

func NewHandler(logger *slog.Logger, nodes storage.NodesRepository) *Handler {
	return &Handler{logger: logger, nodes: nodes}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/nodes/register", h.Register)
	mux.HandleFunc("POST /api/v1/nodes/{id}/heartbeat", h.Heartbeat)
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
		if h.logger != nil {
			h.logger.Error("node registration failed", "error", err)
		}
		httpapi.WriteStorageError(w)
		return
	}

	httpapi.WriteJSON(w, http.StatusCreated, httpapi.Response{Data: map[string]any{
		"node_id":    result.Node.ID,
		"node_token": result.NodeToken,
		"status":     result.Node.Status,
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
	if status != "healthy" && status != "degraded" && status != "registered" && status != "offline" {
		httpapi.WriteBadRequest(w, "status must be registered, healthy, degraded, or offline")
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
			httpapi.WriteNotFound(w, "node")
			return
		}
		if h.logger != nil {
			h.logger.Error("node heartbeat failed", "error", err)
		}
		httpapi.WriteStorageError(w)
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, httpapi.Response{Data: map[string]any{
		"node_id":         node.ID,
		"status":          node.Status,
		"drain_state":     node.DrainState,
		"active_revision": node.ActiveRevision,
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
