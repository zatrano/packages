package workflow

import "time"

type options struct {
	timeout  time.Duration
	cp       Checkpointer
	execID   string
	observer Observer
	maxHops  int
}

// Option configures a single Run or Resume.
type Option func(*options)

// WithTimeout sets a workflow-level deadline (distinct from per-step Timeout).
func WithTimeout(d time.Duration) Option {
	return func(o *options) { o.timeout = d }
}

// WithCheckpointer persists hop/wait state.
func WithCheckpointer(c Checkpointer) Option {
	return func(o *options) { o.cp = c }
}

// WithExecutionID sets a stable execution identity (required to Resume).
func WithExecutionID(id string) Option {
	return func(o *options) { o.execID = id }
}

// WithObserver receives hop traces.
func WithObserver(obs Observer) Option {
	return func(o *options) { o.observer = obs }
}

func applyOptions(opt []Option) options {
	var o options
	for _, fn := range opt {
		if fn != nil {
			fn(&o)
		}
	}
	return o
}
