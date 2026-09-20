package url_test

import (
	"testing"

	"github.com/zatrano/framework/v2/kernel/routing"
	"github.com/zatrano/packages/url"
)

func TestGenerator(t *testing.T) {
	router := routing.New()
	router.Get("/users/{id}", nil).As("users.show")
	router.RegisterName(router.Routes()[0])

	gen := url.New(router, "http://localhost:8080")
	got, err := gen.Route("users.show", map[string]string{"id": "5"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://localhost:8080/users/5" {
		t.Fatalf("got %q", got)
	}
	if gen.Asset("app.css") != "http://localhost:8080/app.css" {
		t.Fatalf("asset=%q", gen.Asset("app.css"))
	}
	if !gen.HasRoute("users.show") {
		t.Fatal("named route with required params must exist")
	}
	if gen.HasRoute("missing") {
		t.Fatal("unknown name")
	}

	bare := url.New(nil, "http://localhost:8080")
	if _, err := bare.Route("users.show"); err == nil {
		t.Fatal("nil router must error")
	}
	if bare.HasRoute("users.show") {
		t.Fatal("nil router has no routes")
	}
}
