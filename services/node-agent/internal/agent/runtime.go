package agent

import (
	"context"
	"strings"
	"time"
)

const (
	RuntimeModeNoProcess  = "no-process"
	RuntimeModeDryRunOnly = "dry-run-only"

	RuntimeDesiredStateConfigReady = "validated-config-ready"

	RuntimeStateNotPrepared       = "not_prepared"
	RuntimeStateActiveConfigReady = "active_config_ready"
	RuntimeStateValidationFailed  = "validation_failed"
	RuntimeStatePrepareFailed     = "prepare_failed"

	RuntimeAttemptSkipped = "skipped"
	RuntimeAttemptReady   = "ready"
	RuntimeAttemptFailed  = "failed"

	DryRunStatusNotConfigured = "not_configured"
	DryRunStatusPassed        = "passed"
	DryRunStatusFailed        = "failed"
)

type RuntimePrepareRequest struct {
	Revision     ConfigRevision
	Artifact     ConfigArtifact
	DryRunStatus string
	At           time.Time
}

type RuntimeTransition struct {
	State        string
	Attempt      string
	ErrorMessage string
	At           time.Time
}

type RuntimeSupervisor interface {
	PrepareActiveConfig(ctx context.Context, request RuntimePrepareRequest) (RuntimeTransition, error)
}

type NoProcessRuntimeSupervisor struct{}

func (NoProcessRuntimeSupervisor) PrepareActiveConfig(ctx context.Context, request RuntimePrepareRequest) (RuntimeTransition, error) {
	if err := ctx.Err(); err != nil {
		return RuntimeTransition{}, err
	}
	at := request.At
	if at.IsZero() {
		at = time.Now().UTC()
	}
	return RuntimeTransition{
		State:   RuntimeStateActiveConfigReady,
		Attempt: RuntimeAttemptSkipped,
		At:      at.UTC(),
	}, nil
}

func runtimeModeForXrayBin(xrayBin string) string {
	if strings.TrimSpace(xrayBin) != "" {
		return RuntimeModeDryRunOnly
	}
	return RuntimeModeNoProcess
}

func dryRunStatusForValidator(validator XrayDryRunValidator) string {
	if validator == nil {
		return DryRunStatusNotConfigured
	}
	return DryRunStatusPassed
}
