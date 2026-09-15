package workflow

import (
	"context"
	"fmt"
	"strings"
)

// Sequential builds a Graph that runs executors in order.
func Sequential(id string, steps ...Executor) *Graph {
	g := &Graph{ID: strings.TrimSpace(id), Nodes: map[string]Node{}}
	if len(steps) == 0 {
		return g
	}
	names := uniqueNames(steps)
	g.Start = names[0]
	for i, exec := range steps {
		n := names[i]
		var route func(context.Context, Envelope) (string, error)
		if i < len(steps)-1 {
			next := names[i+1]
			route = func(context.Context, Envelope) (string, error) { return next, nil }
		}
		g.Nodes[n] = Node{Exec: exec, Route: route}
	}
	return g
}

func uniqueNames(steps []Executor) []string {
	taken := map[string]bool{}
	out := make([]string, len(steps))
	for i, exec := range steps {
		base := ""
		if exec != nil {
			base = strings.TrimSpace(exec.Name())
		}
		if base == "" {
			base = fmt.Sprintf("step-%d", i+1)
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
