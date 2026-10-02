package factory_test

import (
	"errors"
	"testing"
	"time"

	"github.com/zatrano/packages/factory"
)

type stateUser struct {
	Name            string
	Email           string
	EmailVerifiedAt *time.Time
}

func TestFactoryStates(t *testing.T) {
	factory.ClearStates()
	factory.For[stateUser](func() map[string]any {
		return map[string]any{
			"name":  "Plain",
			"email": "plain@example.test",
		}
	})
	factory.RegisterState[stateUser]("verified", func() map[string]any {
		now := time.Now().UTC()
		return map[string]any{"email_verified_at": now, "name": "Verified"}
	})
	factory.RegisterState[stateUser]("admin", func() map[string]any {
		return map[string]any{"name": "Admin"}
	})

	if !factory.HasState[stateUser]("verified") {
		t.Fatal("expected verified state")
	}

	attrs, err := factory.Of[stateUser]().State("verified", "admin").Merge(map[string]any{
		"email": "combo@example.test",
	}).Make()
	if err != nil {
		t.Fatal(err)
	}
	if attrs["name"] != "Admin" || attrs["email"] != "combo@example.test" || attrs["email_verified_at"] == nil {
		t.Fatalf("attrs=%v", attrs)
	}

	_, err = factory.Of[stateUser]().State("verified").Create()
	if !errors.Is(err, factory.ErrDeprecatedPersist) {
		t.Fatalf("Create want ErrDeprecatedPersist, got %v", err)
	}
}
