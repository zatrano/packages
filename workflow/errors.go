package workflow

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrInvalid           = errors.New("workflow: invalid argument")
	ErrNotWaiting        = errors.New("workflow: execution is not waiting")
	ErrUnknownExecution  = errors.New("workflow: unknown execution")
	ErrMaxHops           = errors.New("workflow: max hops reached")
	ErrUnknownNode       = errors.New("workflow: unknown node")
	ErrUnknownRoute      = errors.New("workflow: unknown route")
	ErrRejected          = errors.New("workflow: approval rejected")
	ErrNeverStarted      = errors.New("workflow: part never started")
	ErrNestedWait        = errors.New("workflow: nested human wait is not supported")
	ErrMissingCheckpoint = errors.New("workflow: Resume requires WithCheckpointer")
)

// WaitError stops a graph in StatusWaiting (in-process pause). It is not a failure.
type WaitError struct {
	Wait Wait
}

func (e *WaitError) Error() string {
	if e == nil {
		return "workflow: waiting"
	}
	if e.Wait.Reason != "" {
		return "workflow: waiting: " + e.Wait.Reason
	}
	if e.Wait.Kind != "" {
		return "workflow: waiting: " + e.Wait.Kind
	}
	return "workflow: waiting"
}

// IsWait reports a top-level in-process wait. NestedWaitError is not a wait.
func IsWait(err error) (*Wait, bool) {
	if err == nil || errors.Is(err, ErrNestedWait) {
		return nil, false
	}
	var w *WaitError
	if errors.As(err, &w) && w != nil {
		cp := w.Wait
		return &cp, true
	}
	return nil, false
}

// NestedWaitError is returned when a child executor pauses for a human.
// Nested HITL is not supported; this is a failure, not StatusWaiting.
type NestedWaitError struct {
	Name string
}

func (e *NestedWaitError) Error() string {
	name := ""
	if e != nil {
		name = e.Name
	}
	if name == "" {
		return ErrNestedWait.Error()
	}
	return fmt.Sprintf("%s: %s", ErrNestedWait.Error(), name)
}

func (e *NestedWaitError) Unwrap() error { return ErrNestedWait }

func rejectNestedWait(name string, err error) error {
	if _, ok := IsWait(err); ok {
		return &NestedWaitError{Name: name}
	}
	return err
}

// PartialError is returned when one or more Parallel children fail.
// Successful siblings are retained in Parts; they are not discarded.
type PartialError struct {
	Failed map[string]error
	Parts  map[string]Envelope
}

func (e *PartialError) Error() string {
	if e == nil || len(e.Failed) == 0 {
		return "workflow: partial failure"
	}
	names := make([]string, 0, len(e.Failed))
	for k := range e.Failed {
		names = append(names, k)
	}
	sort.Strings(names)
	return fmt.Sprintf("workflow: parallel failed: %s", strings.Join(names, ", "))
}

func (e *PartialError) Unwrap() []error {
	if e == nil {
		return nil
	}
	names := make([]string, 0, len(e.Failed))
	for k := range e.Failed {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]error, 0, len(names))
	for _, k := range names {
		if err := e.Failed[k]; err != nil {
			out = append(out, err)
		}
	}
	return out
}
