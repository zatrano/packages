package workflow

import "context"

// Observer is notified after each hop. It is not a second telemetry system;
// applications may forward traces into packages/observability.
type Observer interface {
	OnHop(ctx context.Context, t Trace)
}

// FuncObserver adapts a function to Observer.
type FuncObserver func(ctx context.Context, t Trace)

// OnHop implements Observer.
func (f FuncObserver) OnHop(ctx context.Context, t Trace) {
	if f != nil {
		f(ctx, t)
	}
}
