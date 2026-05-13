package audit

import "context"

type Event struct {
	ActorType    string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
}

type Recorder interface {
	Record(ctx context.Context, event Event) error
}

type NoopRecorder struct {
}

func (NoopRecorder) Record(ctx context.Context, event Event) error {
	return nil
}
