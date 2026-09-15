package workflow_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zatrano/packages/workflow"
)

func TestParallelMultipleFailuresPreserveParts(t *testing.T) {
	a := workflow.Named("a", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		env.Output = "A"
		return env, errors.New("fail-a")
	})
	b := workflow.Named("b", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		env.Output = "B"
		return env, errors.New("fail-b")
	})
	c := echo("c", "C")
	res, err := workflow.Sequential("p", workflow.Parallel("fan", nil, a, b, c)).Run(context.Background(), workflow.Envelope{Input: "in"})
	var pe *workflow.PartialError
	if !errors.As(err, &pe) {
		t.Fatalf("err=%v", err)
	}
	if pe.Failed["a"] == nil || pe.Failed["b"] == nil {
		t.Fatalf("failed=%v", pe.Failed)
	}
	if pe.Failed["c"] != nil {
		t.Fatalf("c should succeed: %v", pe.Failed["c"])
	}
	if pe.Parts["c"].Output == "" {
		t.Fatalf("c part lost: %+v", pe.Parts)
	}
	if pe.Error() != "workflow: parallel failed: a, b" {
		t.Fatalf("nondeterministic error string %q", pe.Error())
	}
	if res.Status != workflow.StatusFailed {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestParallelCancelNeverStartedNotSuccess(t *testing.T) {
	started := make(chan struct{})
	block := workflow.Named("block", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		close(started)
		<-ctx.Done()
		return env, ctx.Err()
	})
	idle := make([]workflow.Executor, 6)
	for i := range idle {
		idle[i] = echo("idle", "x")
	}
	parts := append([]workflow.Executor{block}, idle...)
	g := workflow.Sequential("p", workflow.Parallel("fan", nil, parts...))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := g.Run(ctx, workflow.Envelope{Input: "in"})
		done <- err
	}()
	<-started
	cancel()
	err := <-done
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *workflow.PartialError
	if errors.As(err, &pe) {
		for k, ferr := range pe.Failed {
			if ferr == nil {
				t.Fatalf("%s nil failure", k)
			}
		}
		for k, part := range pe.Parts {
			if _, failed := pe.Failed[k]; !failed && strings.TrimSpace(part.Output) == "" {
				t.Fatalf("unstarted %s treated as empty success", k)
			}
		}
		return
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestParallelTimeout(t *testing.T) {
	slow := workflow.Named("slow", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		<-ctx.Done()
		return env, ctx.Err()
	})
	fast := echo("fast", "ok")
	_, err := workflow.Sequential("p", workflow.Parallel("fan", nil, slow, fast)).Run(
		context.Background(),
		workflow.Envelope{Input: "in"},
		workflow.WithTimeout(30*time.Millisecond),
	)
	if err == nil {
		t.Fatal("expected timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, workflow.ErrNeverStarted) {
		var pe *workflow.PartialError
		if !errors.As(err, &pe) {
			t.Fatalf("err=%v", err)
		}
		found := false
		for _, ferr := range pe.Failed {
			if errors.Is(ferr, context.DeadlineExceeded) || errors.Is(ferr, workflow.ErrNeverStarted) {
				found = true
			}
		}
		if !found {
			t.Fatalf("failed=%v", pe.Failed)
		}
	}
}

func TestParallelDuplicateNames(t *testing.T) {
	res, err := workflow.Sequential("p", workflow.Parallel("fan", nil, echo("same", "1"), echo("same", "2"))).Run(
		context.Background(), workflow.Envelope{Input: "in"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Envelope.Output, "1") || !strings.Contains(res.Envelope.Output, "2") {
		t.Fatalf("%q", res.Envelope.Output)
	}
}

func TestParallelNestedWaitRejected(t *testing.T) {
	_, err := workflow.Sequential("p", workflow.Parallel("fan", nil, workflow.Approval("gate", "no"), echo("ok", "x"))).Run(
		context.Background(), workflow.Envelope{Input: "in"},
	)
	if !errors.Is(err, workflow.ErrNestedWait) {
		t.Fatalf("err=%v", err)
	}
	if _, ok := workflow.IsWait(err); ok {
		t.Fatal("nested wait must not be IsWait")
	}
}

func TestParallelAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := workflow.Sequential("p", workflow.Parallel("fan", nil, echo("a", "A"))).Run(ctx, workflow.Envelope{Input: "in"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestParallelWaitForChildrenOnCancel(t *testing.T) {
	started := make(chan struct{})
	block := workflow.Named("block", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		close(started)
		<-ctx.Done()
		return env, ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_, _ = workflow.Sequential("p", workflow.Parallel("fan", nil, block)).Run(ctx, workflow.Envelope{Input: "in"})
		close(done)
	}()
	<-started
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("parallel did not return after cancel")
	}
}
