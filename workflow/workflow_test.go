package workflow_test

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/zatrano/packages/workflow"
)

func echo(name, suffix string) workflow.Executor {
	return workflow.Named(name, func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		if err := ctx.Err(); err != nil {
			return env, err
		}
		env.Output = strings.TrimSpace(env.Input + " " + suffix)
		env.Previous = env.Output
		return env, nil
	})
}

func TestSequential(t *testing.T) {
	g := workflow.Sequential("s", echo("a", "A"), echo("b", "B"))
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != workflow.StatusCompleted {
		t.Fatalf("status=%s", res.Status)
	}
	if !strings.Contains(res.Envelope.Output, "B") {
		t.Fatalf("output=%q", res.Envelope.Output)
	}
	if len(res.Traces) != 2 {
		t.Fatalf("traces=%d", len(res.Traces))
	}
}

func TestBranch(t *testing.T) {
	yes := workflow.Sequential("yes", echo("y", "YES"))
	no := workflow.Sequential("no", echo("n", "NO"))
	g := workflow.Branch("pick", func(env workflow.Envelope) (string, error) {
		if strings.Contains(env.Input, "go") {
			return "yes", nil
		}
		return "no", nil
	}, map[string]*workflow.Graph{"yes": yes, "no": no}, nil)
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "go"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Envelope.Output, "YES") {
		t.Fatalf("%q", res.Envelope.Output)
	}
}

func TestParallelJoinAndPartialFailure(t *testing.T) {
	ok := echo("ok", "OK")
	bad := workflow.Named("bad", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		return env, errors.New("boom")
	})
	sib := echo("sib", "SIB")
	p := workflow.Parallel("fan", workflow.ConcatJoin, ok, bad, sib)
	g := workflow.Sequential("p", p)
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "in"})
	if err == nil {
		t.Fatal("expected partial error")
	}
	var pe *workflow.PartialError
	if !errors.As(err, &pe) {
		t.Fatalf("got %T %v", err, err)
	}
	if pe.Failed["bad"] == nil {
		t.Fatalf("missing bad: %+v", pe.Failed)
	}
	if pe.Parts["sib"].Output == "" && pe.Parts["ok"].Output == "" {
		t.Fatalf("siblings discarded: %+v", pe.Parts)
	}
	if res.Status != workflow.StatusFailed {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestParallelSuccess(t *testing.T) {
	g := workflow.Sequential("p", workflow.Parallel("fan", nil, echo("left", "L"), echo("right", "R")))
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "in"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Envelope.Output, "L") || !strings.Contains(res.Envelope.Output, "R") {
		t.Fatalf("%q", res.Envelope.Output)
	}
}

func TestSupervisorStrategy(t *testing.T) {
	rounds := 0
	router := workflow.Named("router", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		rounds++
		if rounds == 1 {
			env.Route = "writer"
			return env, nil
		}
		env.Route = ""
		return env, nil
	})
	writer := echo("writer", "done")
	sup := workflow.Supervisor("sup", router, map[string]workflow.Executor{"writer": writer}, 4)
	g := workflow.Sequential("g", sup)
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Envelope.Output, "done") {
		t.Fatalf("%q", res.Envelope.Output)
	}
}

func TestHumanApprovalPauseResume(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("hitl", echo("prep", "ready"), workflow.Approval("gate", "need manager"), echo("fin", "shipped"))
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "job"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("e1"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatalf("want wait, got %v", err)
	}
	if res.Status != workflow.StatusWaiting {
		t.Fatalf("status=%s", res.Status)
	}
	res2, err := g.Resume(context.Background(), "e1", workflow.Decision{Approved: true}, workflow.WithCheckpointer(cp))
	if err != nil {
		t.Fatal(err)
	}
	if res2.Status != workflow.StatusCompleted {
		t.Fatalf("status=%s", res2.Status)
	}
	if !strings.Contains(res2.Envelope.Output, "shipped") {
		t.Fatalf("%q", res2.Envelope.Output)
	}
}

func TestApprovalReject(t *testing.T) {
	cp := &workflow.MemoryCheckpointer{}
	g := workflow.Sequential("hitl", workflow.Approval("gate", "no"))
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"}, workflow.WithCheckpointer(cp), workflow.WithExecutionID("e2"))
	if _, ok := workflow.IsWait(err); !ok {
		t.Fatal(err)
	}
	res, err := g.Resume(context.Background(), "e2", workflow.Decision{Approved: false}, workflow.WithCheckpointer(cp))
	if err == nil {
		t.Fatal("expected reject")
	}
	if res.Status != workflow.StatusFailed {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestCancelDuringStep(t *testing.T) {
	started := make(chan struct{})
	block := workflow.Named("block", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		close(started)
		<-ctx.Done()
		return env, ctx.Err()
	})
	g := workflow.Sequential("c", block)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var res *workflow.Result
	var err error
	go func() {
		res, err = g.Run(ctx, workflow.Envelope{Input: "x"})
		close(done)
	}()
	<-started
	cancel()
	<-done
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	if res.Status != workflow.StatusCancelled {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestStepTimeout(t *testing.T) {
	block := workflow.Named("slow", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		<-ctx.Done()
		return env, ctx.Err()
	})
	g := &workflow.Graph{
		ID:    "t",
		Start: "slow",
		Nodes: map[string]workflow.Node{
			"slow": {Exec: block, Timeout: 20 * time.Millisecond},
		},
	}
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
	if res.Status != workflow.StatusFailed {
		t.Fatalf("status=%s (timeout is step failure, not cancel)", res.Status)
	}
}

func TestWorkflowTimeout(t *testing.T) {
	block := workflow.Named("slow", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		<-ctx.Done()
		return env, ctx.Err()
	})
	g := workflow.Sequential("t", block)
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"}, workflow.WithTimeout(20*time.Millisecond))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
}

func TestMaxHops(t *testing.T) {
	g := &workflow.Graph{
		Start:   "loop",
		MaxHops: 3,
		Nodes: map[string]workflow.Node{
			"loop": {
				Exec:  echo("loop", "x"),
				Route: func(context.Context, workflow.Envelope) (string, error) { return "loop", nil },
			},
		},
	}
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "i"})
	if !errors.Is(err, workflow.ErrMaxHops) {
		t.Fatalf("err=%v", err)
	}
}

func TestHandoffEnvelopeNotTranscript(t *testing.T) {
	step := workflow.Named("handoff", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		if env.Task != "refund" || env.Goal == "" {
			t.Fatalf("envelope %+v", env)
		}
		env.Output = "ok"
		env.Artifacts = map[string]string{"ticket": "1"}
		return env, nil
	})
	res, err := workflow.Sequential("h", step).Run(context.Background(), workflow.Envelope{
		Task: "refund", Goal: "credit the user", Input: "please refund",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Envelope.Artifacts["ticket"] != "1" {
		t.Fatalf("%+v", res.Envelope.Artifacts)
	}
}

func TestWorkflowDoesNotImportIntelligence(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "github.com/zatrano/packages/workflow").CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "github.com/zatrano/packages/ai",
			"github.com/zatrano/packages/rag",
			"github.com/zatrano/packages/agent",
			"github.com/zatrano/packages/queue":
			t.Fatalf("forbidden dependency %s", line)
		}
	}
}
