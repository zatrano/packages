package workflow

import (
	"context"
	"fmt"
	"sync"
)

// Checkpoint is in-process hop state for pause/resume. It is not crash-safe durability.
type Checkpoint struct {
	ExecutionID string
	WorkflowID  string
	StepID      string
	Status      Status
	Envelope    Envelope
	Path        []string
}

// Checkpointer stores and loads execution checkpoints for in-process Resume.
// Implementations are not required to survive process restart.
type Checkpointer interface {
	Save(ctx context.Context, cp Checkpoint) error
	Load(ctx context.Context, executionID string) (Checkpoint, error)
	ClaimWaiting(ctx context.Context, executionID string) (Checkpoint, error)
}

// MemoryCheckpointer is an in-process Checkpointer. Data is lost when the process exits.
type MemoryCheckpointer struct {
	mu   sync.Mutex
	data map[string]Checkpoint
}

func (m *MemoryCheckpointer) init() {
	if m.data == nil {
		m.data = map[string]Checkpoint{}
	}
}

func copyCheckpoint(cp Checkpoint) Checkpoint {
	cp.Envelope = cp.Envelope.clone()
	cp.Path = append([]string(nil), cp.Path...)
	return cp
}

// Save implements Checkpointer. It does not require a live ctx so a cancelled
// Run can still persist StatusCancelled / StatusFailed.
func (m *MemoryCheckpointer) Save(_ context.Context, cp Checkpoint) error {
	if m == nil {
		return fmt.Errorf("%w: checkpointer is nil", ErrInvalid)
	}
	if cp.ExecutionID == "" {
		return fmt.Errorf("%w: checkpoint missing ExecutionID", ErrInvalid)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.init()
	m.data[cp.ExecutionID] = copyCheckpoint(cp)
	return nil
}

// Load implements Checkpointer.
func (m *MemoryCheckpointer) Load(ctx context.Context, executionID string) (Checkpoint, error) {
	if m == nil {
		return Checkpoint{}, fmt.Errorf("%w: checkpointer is nil", ErrInvalid)
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return Checkpoint{}, err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cp, ok := m.data[executionID]
	if !ok {
		return Checkpoint{}, fmt.Errorf("%w: %q", ErrUnknownExecution, executionID)
	}
	return copyCheckpoint(cp), nil
}

// ClaimWaiting loads a waiting checkpoint and marks it running so a second Resume fails.
func (m *MemoryCheckpointer) ClaimWaiting(ctx context.Context, executionID string) (Checkpoint, error) {
	if m == nil {
		return Checkpoint{}, fmt.Errorf("%w: checkpointer is nil", ErrInvalid)
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return Checkpoint{}, err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cp, ok := m.data[executionID]
	if !ok {
		return Checkpoint{}, fmt.Errorf("%w: %q", ErrUnknownExecution, executionID)
	}
	if cp.Status != StatusWaiting {
		return Checkpoint{}, fmt.Errorf("%w: status=%s", ErrNotWaiting, cp.Status)
	}
	cp.Status = StatusRunning
	m.data[executionID] = cp
	return copyCheckpoint(cp), nil
}
