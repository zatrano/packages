package workflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zatrano/packages/workflow"
)

func TestSequentialInvalidEmpty(t *testing.T) {
	_, err := workflow.Sequential("s").Run(context.Background(), workflow.Envelope{Input: "x"})
	if !errors.Is(err, workflow.ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestSequentialStepError(t *testing.T) {
	g := workflow.Sequential("s", echo("a", "A"), workflow.Named("bad", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		return env, errors.New("step fail")
	}))
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if err == nil || err.Error() != "step fail" {
		t.Fatalf("err=%v", err)
	}
	if res.Status != workflow.StatusFailed {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestSequentialMissingExecutor(t *testing.T) {
	g := &workflow.Graph{Start: "a", Nodes: map[string]workflow.Node{"a": {}}}
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if !errors.Is(err, workflow.ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestUnknownStartNode(t *testing.T) {
	g := &workflow.Graph{Start: "missing", Nodes: map[string]workflow.Node{"a": {Exec: echo("a", "A")}}}
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if !errors.Is(err, workflow.ErrUnknownNode) {
		t.Fatalf("err=%v", err)
	}
}

func TestStepTimeoutBeforeWorkflowTimeout(t *testing.T) {
	block := workflow.Named("slow", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		<-ctx.Done()
		return env, ctx.Err()
	})
	g := &workflow.Graph{
		ID:    "t",
		Start: "slow",
		Nodes: map[string]workflow.Node{
			"slow": {Exec: block, Timeout: 15 * time.Millisecond},
		},
	}
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "x"}, workflow.WithTimeout(time.Hour))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
	if res.Status != workflow.StatusFailed {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestCompletionBeatsTimeout(t *testing.T) {
	g := workflow.Sequential("s", echo("a", "A"))
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "x"}, workflow.WithTimeout(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != workflow.StatusCompleted {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestCancelBeatsPendingTimeout(t *testing.T) {
	started := make(chan struct{})
	block := workflow.Named("block", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		close(started)
		<-ctx.Done()
		return env, ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan *workflow.Result)
	go func() {
		res, _ := workflow.Sequential("s", block).Run(ctx, workflow.Envelope{Input: "x"}, workflow.WithTimeout(time.Hour))
		done <- res
	}()
	<-started
	cancel()
	res := <-done
	if res.Status != workflow.StatusCancelled {
		t.Fatalf("status=%s", res.Status)
	}
}
