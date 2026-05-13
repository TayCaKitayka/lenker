package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
	"github.com/lenker/lenker/services/panel-api/internal/audit"
	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
)

type SessionMiddleware struct {
	logger *slog.Logger
	admins admins.Repository
	audit  audit.Recorder
}

func NewSessionMiddleware(logger *slog.Logger, admins admins.Repository) *SessionMiddleware {
	return &SessionMiddleware{
		logger: logger,
		admins: admins,
		audit:  audit.NoopRecorder{},
	}
}

func (m *SessionMiddleware) WithAudit(recorder audit.Recorder) *SessionMiddleware {
	if recorder != nil {
		m.audit = recorder
	}
	return m
}

func (m *SessionMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			m.recordSessionValidation(r, "", audit.OutcomeFailure, "missing_token")
			httpapi.WriteUnauthorized(w)
			return
		}

		admin, err := m.admins.FindByActiveSessionTokenHash(r.Context(), HashSessionToken(token), time.Now().UTC())
		if err != nil {
			if errors.Is(err, admins.ErrNotFound) {
				m.recordSessionValidation(r, "", audit.OutcomeFailure, "invalid_session")
				httpapi.WriteUnauthorized(w)
				return
			}
			m.recordSessionValidation(r, "", audit.OutcomeFailure, "internal_error")
			m.logger.Error("admin session validation failed", "error", err)
			httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
			return
		}

		m.recordSessionValidation(r, admin.ID, audit.OutcomeSuccess, "")
		next.ServeHTTP(w, r.WithContext(WithAdmin(r.Context(), admin)))
	})
}

func (m *SessionMiddleware) recordSessionValidation(r *http.Request, adminID string, outcome string, reason string) {
	_ = m.audit.Record(r.Context(), audit.Event{
		ActorType:    "admin",
		ActorID:      adminID,
		Action:       audit.ActionAdminSessionValidation,
		ResourceType: "admin_session",
		Outcome:      outcome,
		Reason:       reason,
	})
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
