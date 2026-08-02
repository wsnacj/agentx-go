package hostkit

import "context"

type TaskObservation struct {
	SemanticTool  string `json:"semantic_tool,omitempty"`
	RuntimeTool   string `json:"runtime_tool,omitempty"`
	RuntimeKind   string `json:"runtime_kind,omitempty"`
	Status        string `json:"status,omitempty"`
	AdapterStatus string `json:"adapter_status,omitempty"`
	FailureCode   string `json:"failure_code,omitempty"`
	Source        string `json:"source,omitempty"`
}

type TaskObserver interface {
	ObserveBrowserTask(context.Context, TaskObservation)
}

type TaskObserverFunc func(context.Context, TaskObservation)

func (f TaskObserverFunc) ObserveBrowserTask(ctx context.Context, observation TaskObservation) {
	if f != nil {
		f(ctx, observation)
	}
}

func MultiTaskObserver(observers ...TaskObserver) TaskObserver {
	nonNil := make([]TaskObserver, 0, len(observers))
	for _, observer := range observers {
		if observer != nil {
			nonNil = append(nonNil, observer)
		}
	}
	if len(nonNil) == 0 {
		return nil
	}
	return TaskObserverFunc(func(ctx context.Context, observation TaskObservation) {
		for _, observer := range nonNil {
			observer.ObserveBrowserTask(ctx, observation)
		}
	})
}
