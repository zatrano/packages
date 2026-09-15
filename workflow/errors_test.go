package workflow_test

import (
	"errors"
	"testing"

	"github.com/zatrano/packages/workflow"
)

func TestErrorTaxonomyIsAs(t *testing.T) {
	cases := []struct {
		err    error
		target error
	}{
		{workflow.ErrInvalid, workflow.ErrInvalid},
		{workflow.ErrNotWaiting, workflow.ErrNotWaiting},
		{workflow.ErrUnknownExecution, workflow.ErrUnknownExecution},
		{workflow.ErrMaxHops, workflow.ErrMaxHops},
		{workflow.ErrUnknownNode, workflow.ErrUnknownNode},
		{workflow.ErrUnknownRoute, workflow.ErrUnknownRoute},
		{workflow.ErrRejected, workflow.ErrRejected},
		{workflow.ErrNeverStarted, workflow.ErrNeverStarted},
		{workflow.ErrNestedWait, workflow.ErrNestedWait},
		{workflow.ErrMissingCheckpoint, workflow.ErrMissingCheckpoint},
	}
	for _, tc := range cases {
		if !errors.Is(tc.err, tc.target) {
			t.Fatalf("%v", tc.err)
		}
	}
	var nw *workflow.NestedWaitError
	err := &workflow.NestedWaitError{Name: "x"}
	if !errors.As(err, &nw) || !errors.Is(err, workflow.ErrNestedWait) {
		t.Fatal(err)
	}
	if _, ok := workflow.IsWait(err); ok {
		t.Fatal("nested must not be wait")
	}
	var w *workflow.WaitError
	wait := &workflow.WaitError{Wait: workflow.Wait{Kind: "approval"}}
	if !errors.As(wait, &w) {
		t.Fatal("wait")
	}
	if _, ok := workflow.IsWait(wait); !ok {
		t.Fatal("expected wait")
	}
}
