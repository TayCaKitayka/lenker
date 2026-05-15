package agent

import (
	"errors"
	"os"
	"time"
)

const AgentVersion = "0.1.0-dev"

var (
	ErrBootstrapTokenRequired = errors.New("bootstrap token is required")
	ErrNodeIDRequired         = errors.New("node id is required")
)

type Service struct {
	identity Identity
	status   Status
}

func NewService(identity Identity) *Service {
	registered := identity.NodeID != ""
	status := StatusBootstrapping
	if registered {
		status = StatusActive
	}

	return &Service{
		identity: identity,
		status: Status{
			NodeID:     identity.NodeID,
			Status:     status,
			Registered: registered,
			PanelURL:   identity.PanelURL,
		},
	}
}

func (s *Service) Status() Status {
	return s.status
}

func (s *Service) BuildRegistrationPayload() (RegistrationPayload, error) {
	if s.identity.BootstrapToken == "" {
		return RegistrationPayload{}, ErrBootstrapTokenRequired
	}

	hostname, _ := os.Hostname()
	return RegistrationPayload{
		NodeID:         s.identity.NodeID,
		BootstrapToken: s.identity.BootstrapToken,
		AgentVersion:   AgentVersion,
		Hostname:       hostname,
	}, nil
}

func (s *Service) BuildHeartbeatPayload(now time.Time) (HeartbeatPayload, error) {
	if s.identity.NodeID == "" {
		return HeartbeatPayload{}, ErrNodeIDRequired
	}

	return HeartbeatPayload{
		NodeID:         s.identity.NodeID,
		AgentVersion:   AgentVersion,
		Status:         s.status.Status,
		ActiveRevision: s.status.ActiveRevision,
		SentAt:         now.UTC(),
	}, nil
}

func (s *Service) MarkHeartbeatSent(at time.Time) {
	s.status.LastHeartbeatAt = at.UTC()
	if s.status.Status == StatusPending {
		s.status.Status = StatusActive
	}
}

func (s *Service) TrackAppliedRevision(revision ConfigRevision) {
	s.status.ActiveRevision = revision.RevisionNumber
	s.status.LastAppliedRevision = revision.RevisionNumber
}

func (s *Service) PlanRollback(toRevision int, reason string) RollbackPlan {
	return RollbackPlan{
		FromRevision: s.status.ActiveRevision,
		ToRevision:   toRevision,
		Reason:       reason,
	}
}
