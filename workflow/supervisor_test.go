package workflow_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zatrano/packages/workflow"
)

func TestSupervisorWorkerError(t *testing.T) {
	router := workflow.Named("router", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		env.Route = "bad"
		return env, nil
	})
	bad := workflow.Named("bad", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		return env, errors.New("worker boom")
	})
	_, err := workflow.Sequential("g", workflow.Supervisor("sup", router, map[string]workflow.Executor{"bad": bad}, 4)).Run(
		context.Background(), workflow.Envelope{Input: "t"},
	)
	if err == nil || err.Error() != "worker boom" {
		t.Fatalf("err=%v", err)
	}
}

func TestSupervisorInvalidRoute(t *testing.T) {
	router := workflow.Named("router", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		env.Route = "ghost"
		return env, nil
	})
	_, err := workflow.Sequential("g", workflow.Supervisor("sup", router, map[string]workflow.Executor{"w": echo("w", "x")}, 2)).Run(
		context.Background(), workflow.Envelope{Input: "t"},
	)
	if !errors.Is(err, workflow.ErrUnknownRoute) {
		t.Fatalf("err=%v", err)
	}
}

func TestSupervisorRouterError(t *testing.T) {
	router := workflow.Named("router", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		return env, errors.New("route fail")
	})
	_, err := workflow.Sequential("g", workflow.Supervisor("sup", router, map[string]workflow.Executor{}, 2)).Run(
		context.Background(), workflow.Envelope{Input: "t"},
	)
	if err == nil || err.Error() != "route fail" {
		t.Fatalf("err=%v", err)
	}
}

func TestSupervisorCancel(t *testing.T) {
	started := make(chan struct{})
	router := workflow.Named("router", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		env.Route = "w"
		return env, nil
	})
	w := workflow.Named("w", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		close(started)
		<-ctx.Done()
		return env, ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := workflow.Sequential("g", workflow.Supervisor("sup", router, map[string]workflow.Executor{"w": w}, 4)).Run(
			ctx, workflow.Envelope{Input: "t"},
		)
		done <- err
	}()
	<-started
	cancel()
	err := <-done
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestSupervisorTimeout(t *testing.T) {
	router := workflow.Named("router", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		env.Route = "w"
		return env, nil
	})
	w := workflow.Named("w", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		<-ctx.Done()
		return env, ctx.Err()
	})
	_, err := workflow.Sequential("g", workflow.Supervisor("sup", router, map[string]workflow.Executor{"w": w}, 4)).Run(
		context.Background(), workflow.Envelope{Input: "t"}, workflow.WithTimeout(20*time.Millisecond),
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
}

func TestSupervisorMaxRounds(t *testing.T) {
	router := workflow.Named("router", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		env.Route = "w"
		return env, nil
	})
	w := echo("w", "x")
	_, err := workflow.Sequential("g", workflow.Supervisor("sup", router, map[string]workflow.Executor{"w": w}, 2)).Run(
		context.Background(), workflow.Envelope{Input: "t"},
	)
	if !errors.Is(err, workflow.ErrMaxHops) {
		t.Fatalf("err=%v", err)
	}
}

func TestSupervisorNestedWaitRejected(t *testing.T) {
	router := workflow.Named("router", func(ctx context.Context, env workflow.Envelope) (workflow.Envelope, error) {
		env.Route = "gate"
		return env, nil
	})
	_, err := workflow.Sequential("g", workflow.Supervisor("sup", router, map[string]workflow.Executor{
		"gate": workflow.Approval("gate", "no"),
	}, 4)).Run(context.Background(), workflow.Envelope{Input: "t"})
	if !errors.Is(err, workflow.ErrNestedWait) {
		t.Fatalf("err=%v", err)
	}
}

func TestSupervisorMissingRouter(t *testing.T) {
	_, err := workflow.Sequential("g", workflow.Supervisor("sup", nil, nil, 1)).Run(context.Background(), workflow.Envelope{Input: "t"})
	if !errors.Is(err, workflow.ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestSupervisorDoesNotUseOutputAsRoute(t *testing.T) {
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
	res, err := workflow.Sequential("g", workflow.Supervisor("sup", router, map[string]workflow.Executor{"writer": echo("writer", "done")}, 4)).Run(
		context.Background(), workflow.Envelope{Input: "t"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Envelope.Output, "done") {
		t.Fatalf("%q", res.Envelope.Output)
	}
}
