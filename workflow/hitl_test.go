package workflow_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/zatrano/packages/workflow"
)

func TestResumeRequiresCheckpointer(t *testing.T) {
	g := workflow.Sequential("h", workflow.Approval("gate", "x"))
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "j"})
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	_, err = g.Resume(context.Background(), "missing", workflow.Decision{Approved: true})
	if !errors.Is(err, workflow.ErrMissingCheckpoint) {
		t.Fatalf("err=%v", err)
	}
}

func TestResumeUnknownExecution(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.Approval("gate", "x"))
	_, err := g.Resume(context.Background(), "nope", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if !errors.Is(err, workflow.ErrUnknownExecution) {
		t.Fatalf("err=%v", err)
	}
}

func TestResumeRejectedWhenNotWaiting(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", echo("a", "A"))
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "j"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("done1"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != workflow.StatusCompleted {
		t.Fatalf("status=%s", res.Status)
	}
	_, err = g.Resume(context.Background(), "done1", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if !errors.Is(err, workflow.ErrNotWaiting) {
		t.Fatalf("err=%v", err)
	}
}

func TestResumeTwice(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.Approval("gate", "x"), echo("fin", "ok"))
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "j"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("e3"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	res, err := g.Resume(context.Background(), "e3", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != workflow.StatusCompleted {
		t.Fatalf("status=%s", res.Status)
	}
	_, err = g.Resume(context.Background(), "e3", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if !errors.Is(err, workflow.ErrNotWaiting) {
		t.Fatalf("err=%v", err)
	}
}

func TestResumeEmptyID(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.Approval("gate", "x"))
	_, err := g.Resume(context.Background(), "  ", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if !errors.Is(err, workflow.ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestResumeInvalidInputWait(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.InputWait("ask", "need text"))
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "j"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("in1"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	_, err = g.Resume(context.Background(), "in1", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if !errors.Is(err, workflow.ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestResumeInputWait(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.InputWait("ask", "need text"), echo("fin", "ok"))
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "j"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("in2"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	res, err := g.Resume(context.Background(), "in2", workflow.Decision{Approved: true, Input: "hello"}, workflow.WithCheckpointer(cp))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != workflow.StatusCompleted {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestResumeCanceledContextDoesNotClaim(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.Approval("gate", "x"), echo("fin", "ok"))
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "j"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("e4"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = g.Resume(ctx, "e4", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	res, err := g.Resume(context.Background(), "e4", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != workflow.StatusCompleted {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestResumeConcurrent(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.Approval("gate", "x"), echo("fin", "ok"))
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "j"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("e5"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	for i := 0; i < 2; i++ {
		i := i
		go func() {
			defer wg.Done()
			_, errs[i] = g.Resume(context.Background(), "e5", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
		}()
	}
	wg.Wait()
	ok, notWaiting := 0, 0
	for _, e := range errs {
		if e == nil {
			ok++
		}
		if errors.Is(e, workflow.ErrNotWaiting) {
			notWaiting++
		}
	}
	if ok != 1 || notWaiting != 1 {
		t.Fatalf("errs=%v", errs)
	}
}

func TestResumeTimeoutOnContinuation(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	slow := workflow.Named("slow", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		<-ctx.Done()
		return env, ctx.Err()
	})
	g := workflow.Sequential("h", workflow.Approval("gate", "x"), slow)
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "j"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("e6"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	_, err = g.Resume(context.Background(), "e6", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp), workflow.WithTimeout(20*time.Millisecond))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
}

func TestHITLTimeoutOfRunDoesNotBlock(t *testing.T) {
	// Approval returns immediately with WaitError; WithTimeout does not apply to the paused interval.
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.Approval("gate", "x"))
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "j"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("e7"), workflow.WithTimeout(time.Hour))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatalf("err=%v", err)
	}
	if res.Status != workflow.StatusWaiting {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestDecisionNotStoredInMetadata(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("h", workflow.Approval("gate", "x"), echo("fin", "ok"))
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "j", Metadata: map[string]string{"k": "v"}}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("e8"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	res, err := g.Resume(context.Background(), "e8", workflow.Decision{Approved: true, Comment: "n"}, workflow.WithCheckpointer(cp))
	if err != nil {
		t.Fatal(err)
	}
	if res.Envelope.Metadata["decision"] != "" {
		t.Fatalf("decision leaked into metadata: %+v", res.Envelope.Metadata)
	}
	if res.Envelope.Decision == nil || !res.Envelope.Decision.Approved {
		t.Fatalf("typed decision missing: %+v", res.Envelope.Decision)
	}
}

func TestCancelledRunPersistsNotWaiting(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	started := make(chan struct{})
	block := workflow.Named("block", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		close(started)
		<-ctx.Done()
		return env, ctx.Err()
	})
	g := workflow.Sequential("c", block)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := g.Run(ctx, workflow.Envelope{Input: "x"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("cx"))
		done <- err
	}()
	<-started
	cancel()
	<-done
	_, err := g.Resume(context.Background(), "cx", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if !errors.Is(err, workflow.ErrNotWaiting) {
		t.Fatalf("err=%v", err)
	}
}
