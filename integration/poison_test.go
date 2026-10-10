//go:build rawhttp_poison && zatrano_poison

package integration

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
)

func TestFlowsPoisonConcurrent(t *testing.T) {
	app := bootApp(t)
	var mu sync.Mutex
	var failures []string
	var panics []string
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		token := fmt.Sprintf("t%03d", i)
		go func() {
			defer wg.Done()
			defer func() {
				p := recover()
				if p == nil {
					return
				}
				stack := string(debug.Stack())
				if strings.Contains(fmt.Sprint(p), "request used after handler returned") {
					mu.Lock()
					panics = append(panics, fmt.Sprintf("%v\n%s", p, stack))
					mu.Unlock()
					return
				}
				mu.Lock()
				failures = append(failures, fmt.Sprintf("panic %v\n%s", p, stack))
				mu.Unlock()
			}()
			if err := exercise(app, token); err != nil {
				mu.Lock()
				failures = append(failures, token+": "+err.Error())
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	for _, p := range panics {
		t.Logf("POISON %s", p)
	}
	if len(failures) > 0 {
		t.Fatalf("%d failures\n%s", len(failures), strings.Join(failures, "\n"))
	}
	t.Logf("poison panics=%d", len(panics))
}
