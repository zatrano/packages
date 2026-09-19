package orm

import (
	"reflect"
	"strings"
	"sync"
)

// ObserverFunc handles one ORM persistence lifecycle step for a model instance.
// These hooks are not application Facts.
type ObserverFunc func(model any) error

// ModelObserver handles created/updated/deleted persistence hooks.
type ModelObserver interface {
	Created(model any) error
	Updated(model any) error
	Deleted(model any) error
}

// LifecycleObserver extends ModelObserver with saving/saved/retrieved/replicating/restoring hooks.
type LifecycleObserver interface {
	ModelObserver
	Saving(model any) error
	Saved(model any) error
	Retrieved(model any) error
	Replicating(model any) error
	Restoring(model any) error
	Restored(model any) error
	ForceDeleted(model any) error
}

var (
	observerMu sync.RWMutex
	observers  = map[string][]ObserverFunc{}
)

// Observe registers a persistence hook for "{subject}.{action}" (for example user.created).
func Observe(subject, action string, fn ObserverFunc) {
	if fn == nil {
		return
	}
	key := observerKey(subject, action)
	if key == "" {
		return
	}
	observerMu.Lock()
	defer observerMu.Unlock()
	observers[key] = append(observers[key], fn)
}

// ObserveMany registers several actions for one subject.
func ObserveMany(subject string, handlers map[string]ObserverFunc) {
	for action, fn := range handlers {
		Observe(subject, action, fn)
	}
}

// ObserveModel registers created/updated/deleted (and extra lifecycle hooks when implemented).
func ObserveModel(subject string, observer ModelObserver) {
	if observer == nil {
		return
	}
	handlers := map[string]ObserverFunc{
		"created": observer.Created,
		"updated": observer.Updated,
		"deleted": observer.Deleted,
	}
	if life, ok := observer.(LifecycleObserver); ok {
		handlers["saving"] = life.Saving
		handlers["saved"] = life.Saved
		handlers["retrieved"] = life.Retrieved
		handlers["replicating"] = life.Replicating
		handlers["restoring"] = life.Restoring
		handlers["restored"] = life.Restored
		handlers["forceDeleted"] = life.ForceDeleted
	}
	ObserveMany(subject, handlers)
}

// ResetObservers clears persistence hooks (tests).
func ResetObservers() {
	observerMu.Lock()
	defer observerMu.Unlock()
	observers = map[string][]ObserverFunc{}
}

func dispatchModel(action string, model any) error {
	if model == nil {
		return nil
	}
	subject := observerSubject(model)
	if subject == "" {
		return nil
	}
	key := observerKey(subject, action)
	observerMu.RLock()
	fns := append([]ObserverFunc{}, observers[key]...)
	observerMu.RUnlock()
	for _, fn := range fns {
		if err := fn(model); err != nil {
			return err
		}
	}
	return nil
}

func observerKey(subject, action string) string {
	subject = strings.ToLower(strings.TrimSpace(subject))
	action = strings.TrimSpace(action)
	if subject == "" || action == "" {
		return ""
	}
	return subject + "." + action
}

func observerSubject(model any) string {
	rt := reflect.TypeOf(model)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.Name()
	if name == "" {
		return ""
	}
	return strings.ToLower(name)
}
