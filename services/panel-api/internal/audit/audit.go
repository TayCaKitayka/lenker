package audit

import "context"

type Event struct {
	ActorType    string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Outcome      string
	Reason       string
}

type Recorder interface {
	Record(ctx context.Context, event Event) error
}

type NoopRecorder struct {
}

func (NoopRecorder) Record(ctx context.Context, event Event) error {
	return nil
}

const (
	ActionAdminLogin               = "admin.login"
	ActionAdminSessionValidation   = "admin.session_validation"
	ActionUserCreate               = "user.create"
	ActionUserUpdate               = "user.update"
	ActionUserSuspend              = "user.suspend"
	ActionUserActivate             = "user.activate"
	ActionPlanCreate               = "plan.create"
	ActionPlanUpdate               = "plan.update"
	ActionPlanArchive              = "plan.archive"
	ActionSubscriptionCreate       = "subscription.create"
	ActionSubscriptionUpdate       = "subscription.update"
	ActionSubscriptionRenew        = "subscription.renew"
	ActionNodeBootstrapToken       = "node.bootstrap_token.create"
	ActionNodeRegister             = "node.register"
	ActionNodeHeartbeat            = "node.heartbeat"
	ActionNodeDrain                = "node.drain"
	ActionNodeUndrain              = "node.undrain"
	ActionNodeDisable              = "node.disable"
	ActionNodeEnable               = "node.enable"
	ActionNodeConfigRevisionCreate = "node.config_revision.create"
	ActionNodeConfigRevisionFetch  = "node.config_revision.fetch"
	ActionNodeConfigRevisionReport = "node.config_revision.report"

	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
)
