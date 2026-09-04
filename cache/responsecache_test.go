package cache_test

import (
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zatrano/framework/http"
	"github.com/zatrano/packages/cache"
)

type memStore struct {
	data map[string]any
}

func (s *memStore) Get(key string) (any, bool) {
	v, ok := s.data[key]
	return v, ok
}
func (s *memStore) Put(key string, value any, ttl time.Duration) error {
	if s.data == nil {
		s.data = map[string]any{}
	}
	s.data[key] = value
	return nil
}
func (s *memStore) Forever(key string, value any) error { return s.Put(key, value, 0) }
func (s *memStore) Forget(key string) error {
	delete(s.data, key)
	return nil
}
func (s *memStore) Flush() error {
	s.data = map[string]any{}
	return nil
}
func (s *memStore) Has(key string) bool {
	_, ok := s.data[key]
	return ok
}
func (s *memStore) Pull(key string) (any, bool) {
	v, ok := s.Get(key)
	if ok {
		_ = s.Forget(key)
	}
	return v, ok
}
func (s *memStore) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	if v, ok := s.Get(key); ok {
		return v, nil
	}
	v, err := callback()
	if err != nil {
		return nil, err
	}
	_ = s.Put(key, v, ttl)
	return v, nil
}

var _ cache.Store = (*memStore)(nil)

func TestResponseCache(t *testing.T) {
	store := &memStore{}
	hits := 0
	mw := cache.ResponseCache(store, time.Minute)
	handler := mw(func(req *http.Request) *http.Response {
		hits++
		return http.JSON(map[string]any{"n": hits})
	})

	first := handler(http.NewRequest(httptest.NewRequest(stdhttp.MethodGet, "/demo", nil)))
	if first.Headers().Get("X-Response-Cache") != "MISS" {
		t.Fatalf("expected MISS, got %q", first.Headers().Get("X-Response-Cache"))
	}
	second := handler(http.NewRequest(httptest.NewRequest(stdhttp.MethodGet, "/demo", nil)))
	if second.Headers().Get("X-Response-Cache") != "HIT" {
		t.Fatalf("expected HIT, got %q", second.Headers().Get("X-Response-Cache"))
	}
	if hits != 1 {
		t.Fatalf("hits=%d", hits)
	}
}
