package orm

import "testing"

type demoObserver struct {
	created int
	updated int
	deleted int
}

func (o *demoObserver) Created(model any) error {
	o.created++
	return nil
}
func (o *demoObserver) Updated(model any) error {
	o.updated++
	return nil
}
func (o *demoObserver) Deleted(model any) error {
	o.deleted++
	return nil
}

func TestObserveModelHooks(t *testing.T) {
	ResetObservers()
	t.Cleanup(ResetObservers)
	obs := &demoObserver{}
	ObserveModel("observeuser", obs)
	if err := dispatchModel("created", observeUser{}); err != nil {
		t.Fatal(err)
	}
	if err := dispatchModel("updated", observeUser{}); err != nil {
		t.Fatal(err)
	}
	if err := dispatchModel("deleted", observeUser{}); err != nil {
		t.Fatal(err)
	}
	if obs.created != 1 || obs.updated != 1 || obs.deleted != 1 {
		t.Fatalf("unexpected counts %#v", obs)
	}
}

func TestObserveMany(t *testing.T) {
	ResetObservers()
	t.Cleanup(ResetObservers)
	var seen string
	ObserveMany("namedmodel", map[string]ObserverFunc{
		"paid": func(model any) error {
			seen = "paid"
			return nil
		},
	})
	if err := dispatchModel("paid", namedModel{}); err != nil {
		t.Fatal(err)
	}
	if seen != "paid" {
		t.Fatalf("expected paid, got %q", seen)
	}
}

type observeUser struct{}
type namedModel struct{}
