package workflow

import "time"

// Status is the execution lifecycle of one Run/Resume.
type Status string

const (
	StatusCreated   Status = "created"
	StatusRunning   Status = "running"
	StatusWaiting   Status = "waiting"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Result is the outcome of Graph.Run or Resume.
type Result struct {
	ExecutionID string
	WorkflowID  string
	Status      Status
	Envelope    Envelope
	Traces      []Trace
	Wait        *Wait
}

// Trace records one hop.
type Trace struct {
	StepID   string
	Name     string
	Duration time.Duration
	Err      error
}

// Wait describes a human-in-the-loop pause.
type Wait struct {
	Kind   string // "approval" or "input"
	StepID string
	Reason string
}

// Decision resumes a waiting execution.
type Decision struct {
	Approved bool
	Input    string
	Comment  string
}
