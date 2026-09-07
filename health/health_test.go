package health

import (
	"context"
	"fmt"
	"testing"
)

func TestRunEmptyOK(t *testing.T) {
	m := New()
	overall, results := m.Run(context.Background())
	if overall != StatusOK {
		t.Fatalf("overall=%q", overall)
	}
	if len(results) != 0 {
		t.Fatalf("results=%d", len(results))
	}
}

func TestCustomFail(t *testing.T) {
	m := New()
	m.Custom("db", func(ctx context.Context) error {
		return fmt.Errorf("down")
	})
	overall, results := m.Run(context.Background())
	if overall != StatusFail {
		t.Fatalf("overall=%q", overall)
	}
	if len(results) != 1 || results[0].Status != StatusFail {
		t.Fatalf("results=%#v", results)
	}
}

func TestFromNil(t *testing.T) {
	if From(nil) != nil {
		t.Fatal("From(nil) must be nil")
	}
}
