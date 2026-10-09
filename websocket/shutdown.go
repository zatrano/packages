package websocket

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zatrano/framework/v3/core/kernel/http"
)

const regShards = 16

type liveConn struct {
	c    *Conn
	gone chan struct{}
}

type connShard struct {
	mu sync.Mutex
	m  map[*Conn]*liveConn
}

var (
	shards    [regShards]connShard
	nextShard atomic.Uint32
	closing   atomic.Bool
)

func init() {
	for i := range shards {
		shards[i].m = make(map[*Conn]*liveConn)
	}
	http.RegisterShutdownHook("websocket", shutdownAll)
}

func track(c *Conn) {
	c.shard = int(nextShard.Add(1) % regShards)
	lv := &liveConn{c: c, gone: make(chan struct{})}
	sh := &shards[c.shard]
	sh.mu.Lock()
	sh.m[c] = lv
	sh.mu.Unlock()
}

func untrack(c *Conn) {
	sh := &shards[c.shard]
	sh.mu.Lock()
	lv := sh.m[c]
	delete(sh.m, c)
	sh.mu.Unlock()
	if lv == nil {
		return
	}
	select {
	case <-lv.gone:
	default:
		close(lv.gone)
	}
}

func liveCount() int {
	n := 0
	for i := range shards {
		shards[i].mu.Lock()
		n += len(shards[i].m)
		shards[i].mu.Unlock()
	}
	return n
}

func shutdownAll(ctx context.Context) error {
	if !closing.CompareAndSwap(false, true) {
		return nil
	}
	var all []*liveConn
	for i := range shards {
		shards[i].mu.Lock()
		for _, lv := range shards[i].m {
			all = append(all, lv)
		}
		shards[i].mu.Unlock()
	}
	var wg sync.WaitGroup
	for _, lv := range all {
		wg.Add(1)
		go func(lv *liveConn) {
			defer wg.Done()
			closeOne(ctx, lv)
		}(lv)
	}
	wg.Wait()
	return nil
}

func closeOne(ctx context.Context, lv *liveConn) {
	c := lv.c
	if c.raw != nil {
		_ = c.raw.SetWriteDeadline(time.Now().Add(time.Second))
	}
	_ = c.Close(1001)
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-lv.gone:
		return
	case <-timer.C:
	case <-ctx.Done():
	}
	if c.raw != nil {
		_ = c.raw.Close()
	}
	select {
	case <-lv.gone:
	case <-ctx.Done():
	}
}
