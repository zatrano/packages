package workflow

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/zatrano/packages/toolkit/concurrency"
)

// Joiner merges successful Parallel child envelopes into one.
type Joiner func(ctx context.Context, parent Envelope, parts map[string]Envelope) (Envelope, error)

type parallelExec struct {
	name  string
	parts []Executor
	join  Joiner
}

// Parallel fan-out/fan-in as a single Executor hop.
// Children run concurrently via toolkit/concurrency.Pool.
// A failed child does not erase sibling results; they are returned on *PartialError.
// Nested human waits are rejected (ErrNestedWait), not treated as resumable.
func Parallel(name string, join Joiner, parts ...Executor) Executor {
	if join == nil {
		join = ConcatJoin
	}
	return &parallelExec{name: strings.TrimSpace(name), parts: parts, join: join}
}

func (p *parallelExec) Name() string { return p.name }

func (p *parallelExec) Execute(ctx context.Context, env Envelope) (Envelope, error) {
	if p == nil || len(p.parts) == 0 {
		return env, fmt.Errorf("%w: parallel %q has no parts", ErrInvalid, p.name)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return env, err
	}
	n := len(p.parts)
	results := make([]Envelope, n)
	begun := make([]atomic.Bool, n)
	tasks := make([]func(context.Context) error, n)
	keys := uniquePartKeys(p.parts)
	for i, ex := range p.parts {
		i, ex := i, ex
		key := keys[i]
		tasks[i] = func(ctx context.Context) error {
			if ex == nil {
				return fmt.Errorf("%w: parallel part %q is nil", ErrInvalid, key)
			}
			begun[i].Store(true)
			out, err := ex.Execute(ctx, env.clone())
			results[i] = out
			return err
		}
	}
	errs := concurrency.Pool(ctx, n, tasks)
	parts := make(map[string]Envelope, n)
	failed := map[string]error{}
	for i, key := range keys {
		parts[key] = results[i]
		err := classifyParallelErr(ctx, begun[i].Load(), errs[i])
		if err != nil {
			failed[key] = rejectNestedWait(key, err)
		}
	}
	if len(failed) > 0 {
		return env, &PartialError{Failed: failed, Parts: parts}
	}
	if p.join == nil {
		return ConcatJoin(ctx, env, parts)
	}
	return p.join(ctx, env, parts)
}

func classifyParallelErr(ctx context.Context, started bool, err error) error {
	if !started {
		base := err
		if base == nil {
			base = ctx.Err()
		}
		if base == nil {
			base = ErrNeverStarted
		}
		if errors.Is(base, ErrNeverStarted) {
			return base
		}
		return fmt.Errorf("%w: %w", ErrNeverStarted, base)
	}
	return err
}

func uniquePartKeys(parts []Executor) []string {
	taken := map[string]bool{}
	out := make([]string, len(parts))
	for i, ex := range parts {
		base := fmt.Sprintf("part-%d", i+1)
		if ex != nil {
			if nm := strings.TrimSpace(ex.Name()); nm != "" {
				base = nm
			}
		}
		n := base
		for k := 2; taken[n]; k++ {
			n = fmt.Sprintf("%s-%d", base, k)
		}
		taken[n] = true
		out[i] = n
	}
	return out
}

// ConcatJoin writes named child outputs into Envelope.Output (stable key order).
func ConcatJoin(_ context.Context, parent Envelope, parts map[string]Envelope) (Envelope, error) {
	out := parent.clone()
	keys := make([]string, 0, len(parts))
	for k := range parts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(parts[k].Output)
	}
	out.Output = b.String()
	out.Previous = out.Output
	if out.Artifacts == nil {
		out.Artifacts = map[string]string{}
	}
	for _, k := range keys {
		out.Artifacts[k] = parts[k].Output
	}
	return out, nil
}
