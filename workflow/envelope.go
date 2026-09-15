// Package workflow is an import-only library for deterministic executor graphs.
//
// A workflow may contain functions, HTTP or database work, RAG operations, agents,
// and human waits. It does not import ai, rag, or agent.
//
// Human approval is in-process pause/resume (StatusWaiting + Resume). It is not
// crash-safe durable execution. Nested waits inside Parallel, Branch, or
// Supervisor are rejected.
package workflow

import "maps"

// Envelope is structured context transferred between steps.
// It is not a conversation transcript and not a generic map[string]any bag.
// Metadata is auxiliary only — not an execution-control protocol.
type Envelope struct {
	Task        string
	Goal        string
	Input       string
	Previous    string
	Output      string
	Required    string
	Route       string // supervisor / router next-hop hint; not user-visible output
	Artifacts   map[string]string
	Constraints []string
	Permissions []string
	Metadata    map[string]string // auxiliary; not HITL or routing control
	Decision    *Decision         // HITL resume payload; nil on ordinary hops
}

func (e Envelope) clone() Envelope {
	out := e
	if e.Artifacts != nil {
		out.Artifacts = maps.Clone(e.Artifacts)
	}
	if e.Constraints != nil {
		out.Constraints = append([]string(nil), e.Constraints...)
	}
	if e.Permissions != nil {
		out.Permissions = append([]string(nil), e.Permissions...)
	}
	if e.Metadata != nil {
		out.Metadata = maps.Clone(e.Metadata)
	}
	if e.Decision != nil {
		d := *e.Decision
		out.Decision = &d
	}
	return out
}
