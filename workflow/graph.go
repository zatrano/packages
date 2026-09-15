package workflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Node is one named hop in a Graph.
type Node struct {
	Exec    Executor
	Timeout time.Duration // 0 = no step timeout
	Route   func(ctx context.Context, env Envelope) (next string, err error)
}

// Graph is a deterministic executor graph. Nodes are not required to be agents.
type Graph struct {
	ID      string
	Start   string
	Nodes   map[string]Node
	MaxHops int // default 32
}

// Run walks the graph from Start. opt may set timeout, checkpointer, execution id.
func (g *Graph) Run(ctx context.Context, env Envelope, opt ...Option) (*Result, error) {
	if err := g.validate(); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	o := applyOptions(opt)
	execID := strings.TrimSpace(o.execID)
	if execID == "" {
		execID = newExecutionID()
	}
	ctx, cancel := wrapTimeout(ctx, o.timeout)
	if cancel != nil {
		defer cancel()
	}
	res := &Result{
		ExecutionID: execID,
		WorkflowID:  strings.TrimSpace(g.ID),
		Status:      StatusRunning,
		Envelope:    env.clone(),
	}
	return g.walk(ctx, res, g.Start, o)
}

// Resume continues an in-process waiting execution after a human Decision.
// It does not recover executions after process crash. Nested waits are not resumable.
func (g *Graph) Resume(ctx context.Context, executionID string, decision Decision, opt ...Option) (*Result, error) {
	if err := g.validate(); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	o := applyOptions(opt)
	if o.cp == nil {
		return nil, ErrMissingCheckpoint
	}
	executionID = strings.TrimSpace(executionID)
	if executionID == "" {
		return nil, fmt.Errorf("%w: execution id is required", ErrInvalid)
	}
	cp, err := o.cp.ClaimWaiting(ctx, executionID)
	if err != nil {
		return nil, err
	}
	o.execID = executionID
	ctx, cancel := wrapTimeout(ctx, o.timeout)
	if cancel != nil {
		defer cancel()
	}
	env := applyDecision(cp.Envelope, decision)
	res := &Result{
		ExecutionID: executionID,
		WorkflowID:  strings.TrimSpace(g.ID),
		Status:      StatusRunning,
		Envelope:    env,
	}
	return g.walk(ctx, res, cp.StepID, o)
}

func (g *Graph) validate() error {
	if g == nil || len(g.Nodes) == 0 {
		return fmt.Errorf("%w: graph has no nodes", ErrInvalid)
	}
	start := strings.TrimSpace(g.Start)
	if start == "" {
		return fmt.Errorf("%w: graph Start is required", ErrInvalid)
	}
	if _, ok := g.Nodes[start]; !ok {
		return fmt.Errorf("%w: %q", ErrUnknownNode, start)
	}
	return nil
}

func (g *Graph) walk(ctx context.Context, res *Result, start string, o options) (*Result, error) {
	maxHops := g.MaxHops
	if o.maxHops > 0 {
		maxHops = o.maxHops
	}
	if maxHops <= 0 {
		maxHops = 32
	}
	cur := strings.TrimSpace(start)
	for hop := 0; hop < maxHops; hop++ {
		if err := ctx.Err(); err != nil {
			return finish(ctx, o, res, cur, statusForCtx(err), err)
		}
		node, ok := g.Nodes[cur]
		if !ok {
			return finish(ctx, o, res, cur, StatusFailed, fmt.Errorf("%w: %q", ErrUnknownNode, cur))
		}
		if node.Exec == nil {
			return finish(ctx, o, res, cur, StatusFailed, fmt.Errorf("%w: node %q missing Executor", ErrInvalid, cur))
		}

		startAt := time.Now()
		stepCtx, stepCancel := wrapTimeout(ctx, node.Timeout)
		out, err := node.Exec.Execute(stepCtx, res.Envelope.clone())
		if stepCancel != nil {
			stepCancel()
		}
		res.Envelope = out
		tr := Trace{StepID: cur, Name: node.Exec.Name(), Duration: time.Since(startAt), Err: err}
		res.Traces = append(res.Traces, tr)
		if o.observer != nil {
			o.observer.OnHop(ctx, tr)
		}

		if w, ok := IsWait(err); ok {
			wait := *w
			wait.StepID = cur
			res.Wait = &wait
			return finish(ctx, o, res, cur, StatusWaiting, err)
		}
		if err != nil {
			if ctx.Err() != nil {
				return finish(ctx, o, res, cur, statusForCtx(ctx.Err()), err)
			}
			if errors.Is(err, context.DeadlineExceeded) {
				return finish(ctx, o, res, cur, StatusFailed, err)
			}
			return finish(ctx, o, res, cur, StatusFailed, err)
		}
		if saveErr := saveCP(ctx, o, res, cur, StatusRunning); saveErr != nil {
			return finish(ctx, o, res, cur, StatusFailed, saveErr)
		}
		if node.Route == nil {
			return finish(ctx, o, res, cur, StatusCompleted, nil)
		}
		next, rerr := node.Route(ctx, res.Envelope)
		if rerr != nil {
			return finish(ctx, o, res, cur, StatusFailed, rerr)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return finish(ctx, o, res, cur, StatusCompleted, nil)
		}
		cur = next
	}
	return finish(ctx, o, res, cur, StatusFailed, fmt.Errorf("%w (%d)", ErrMaxHops, maxHops))
}

func saveCP(_ context.Context, o options, res *Result, step string, status Status) error {
	if o.cp == nil {
		return nil
	}
	path := make([]string, 0, len(res.Traces))
	for _, t := range res.Traces {
		path = append(path, t.StepID)
	}
	return o.cp.Save(context.Background(), Checkpoint{
		ExecutionID: res.ExecutionID,
		WorkflowID:  res.WorkflowID,
		StepID:      step,
		Status:      status,
		Envelope:    res.Envelope.clone(),
		Path:        path,
	})
}

func finish(ctx context.Context, o options, res *Result, step string, status Status, err error) (*Result, error) {
	res.Status = status
	if saveErr := saveCP(ctx, o, res, step, status); saveErr != nil && err == nil {
		return res, saveErr
	}
	return res, err
}

func statusForCtx(err error) Status {
	if err == nil {
		return StatusFailed
	}
	if errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return StatusCancelled
	}
	return StatusFailed
}

func wrapTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return ctx, nil
	}
	return context.WithTimeout(ctx, d)
}

func applyDecision(env Envelope, d Decision) Envelope {
	env = env.clone()
	cp := d
	env.Decision = &cp
	if strings.TrimSpace(d.Input) != "" {
		env.Input = d.Input
	}
	return env
}

func newExecutionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
