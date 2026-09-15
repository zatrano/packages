package workflow_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zatrano/packages/workflow"
)

func pick(yes, no string) func(workflow.Envelope) (string, error) {
	return func(env workflow.Envelope) (string, error) {
		if strings.Contains(env.Input, "go") {
			return yes, nil
		}
		return no, nil
	}
}

func TestBranchFalsePath(t *testing.T) {
	yes := workflow.Sequential("yes", echo("y", "YES"))
	no := workflow.Sequential("no", echo("n", "NO"))
	g := workflow.Branch("pick", pick("yes", "no"), map[string]*workflow.Graph{"yes": yes, "no": no}, nil)
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "stay"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Envelope.Output, "NO") {
		t.Fatalf("%q", res.Envelope.Output)
	}
}

func TestBranchConditionError(t *testing.T) {
	g := workflow.Branch("pick", func(workflow.Envelope) (string, error) {
		return "", errors.New("bad cond")
	}, map[string]*workflow.Graph{"yes": workflow.Sequential("y", echo("y", "Y"))}, nil)
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if err == nil || err.Error() != "bad cond" {
		t.Fatalf("err=%v", err)
	}
}

func TestBranchInvalidRoute(t *testing.T) {
	g := workflow.Branch("pick", func(workflow.Envelope) (string, error) {
		return "missing", nil
	}, map[string]*workflow.Graph{"yes": workflow.Sequential("y", echo("y", "Y"))}, nil)
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if !errors.Is(err, workflow.ErrUnknownRoute) {
		t.Fatalf("err=%v", err)
	}
}

func TestBranchOtherwise(t *testing.T) {
	yes := workflow.Sequential("yes", echo("y", "YES"))
	other := workflow.Sequential("other", echo("o", "OTHER"))
	g := workflow.Branch("pick", func(workflow.Envelope) (string, error) {
		return "nope", nil
	}, map[string]*workflow.Graph{"yes": yes}, other)
	res, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Envelope.Output, "OTHER") {
		t.Fatalf("%q", res.Envelope.Output)
	}
}

func TestBranchSelectedFails(t *testing.T) {
	bad := workflow.Sequential("bad", workflow.Named("x", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		return env, errors.New("boom")
	}))
	g := workflow.Branch("pick", func(workflow.Envelope) (string, error) {
		return "bad", nil
	}, map[string]*workflow.Graph{"bad": bad}, nil)
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err=%v", err)
	}
}

func TestBranchNestedWaitRejected(t *testing.T) {
	inner := workflow.Sequential("in", workflow.Approval("gate", "no"))
	g := workflow.Branch("pick", func(workflow.Envelope) (string, error) {
		return "in", nil
	}, map[string]*workflow.Graph{"in": inner}, nil)
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if !errors.Is(err, workflow.ErrNestedWait) {
		t.Fatalf("err=%v", err)
	}
}

func TestBranchCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := workflow.Branch("pick", pick("yes", "no"), map[string]*workflow.Graph{
		"yes": workflow.Sequential("y", echo("y", "Y")),
	}, nil)
	_, err := g.Run(ctx, workflow.Envelope{Input: "go"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestBranchTimeout(t *testing.T) {
	slow := workflow.Sequential("s", workflow.Named("slow", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		<-ctx.Done()
		return env, ctx.Err()
	}))
	g := workflow.Branch("pick", func(workflow.Envelope) (string, error) { return "s", nil }, map[string]*workflow.Graph{"s": slow}, nil)
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"}, workflow.WithTimeout(20*time.Millisecond))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
}

func TestBranchNilCondition(t *testing.T) {
	g := workflow.Branch("pick", nil, map[string]*workflow.Graph{
		"a": workflow.Sequential("a", echo("a", "A")),
	}, nil)
	_, err := g.Run(context.Background(), workflow.Envelope{Input: "x"})
	if !errors.Is(err, workflow.ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}
