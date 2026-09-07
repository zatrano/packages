package apitoken_test

import (
	"testing"
	"time"

	"github.com/zatrano/packages/apitoken"
	"github.com/zatrano/packages/auth"
)

type stubProvider struct {
	user *auth.GenericUser
}

func (p *stubProvider) RetrieveByID(id any) (auth.Authenticatable, error) {
	return p.user, nil
}
func (p *stubProvider) RetrieveByCredentials(credentials map[string]string) (auth.Authenticatable, error) {
	return p.user, nil
}
func (p *stubProvider) ValidateCredentials(user auth.Authenticatable, credentials map[string]string) bool {
	return true
}

func TestPersonalAccessTokenCreateAndFind(t *testing.T) {
	user := &auth.GenericUser{Attributes: map[string]any{"id": int64(1), "email": "ada@zatrano.test"}}
	mgr := apitoken.New(apitoken.NewMemoryStore(), &stubProvider{user: user})
	token, err := mgr.Create(1, "cli", []string{"posts:read"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if token.PlainText == "" {
		t.Fatal("expected plain text once")
	}
	found, authUser, err := mgr.Find(token.PlainText)
	if err != nil || found == nil || authUser == nil {
		t.Fatalf("find failed: %#v %#v %v", found, authUser, err)
	}
	if !found.Can("posts:read") || found.Can("admin") {
		t.Fatal("ability mismatch")
	}
}

func TestFindWithoutProviderReturnsToken(t *testing.T) {
	mgr := apitoken.New(apitoken.NewMemoryStore(), nil)
	token, err := mgr.Create(1, "cli", []string{"*"})
	if err != nil {
		t.Fatal(err)
	}
	found, user, err := mgr.Find(token.PlainText)
	if err != nil || found == nil || user != nil {
		t.Fatalf("find without provider: %#v %#v %v", found, user, err)
	}
}

func TestFromNil(t *testing.T) {
	if apitoken.From(nil) != nil {
		t.Fatal("From(nil) must be nil")
	}
}
