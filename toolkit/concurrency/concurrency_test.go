package concurrency_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zatrano/packages/toolkit/concurrency"
)

func TestMap(t *testing.T) {
	results := concurrency.Map(map[string]func() (int, error){
		"a": func() (int, error) { return 1, nil },
		"b": func() (int, error) { return 2, nil },
	})
	if results["a"].Value != 1 || results["b"].Value != 2 {
		t.Fatalf("unexpected results: %#v", results)
	}
}

func TestPoolCancelMarksUnstarted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	var ran atomic.Int32
	tasks := make([]func(context.Context) error, 8)
	for i := range tasks {
		i := i
		tasks[i] = func(ctx context.Context) error {
			ran.Add(1)
			if i == 0 {
				close(started)
			}
			<-ctx.Done()
			return ctx.Err()
		}
	}
	done := make(chan []error, 1)
	go func() {
		done <- concurrency.Pool(ctx, 1, tasks)
	}()
	<-started
	cancel()
	errs := <-done
	if len(errs) != 8 {
		t.Fatalf("len=%d", len(errs))
	}
	for i, err := range errs {
		if err == nil {
			t.Fatalf("index %d reported success without running to completion", i)
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("index %d err=%v", i, err)
		}
	}
}

func TestPoolSuccessNilMeansRan(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	errs := concurrency.Pool(ctx, 2, []func(context.Context) error{
		func(context.Context) error { return nil },
		func(context.Context) error { return nil },
	})
	for i, err := range errs {
		if err != nil {
			t.Fatalf("index %d err=%v", i, err)
		}
	}
}
