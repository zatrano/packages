package hashing

import "testing"

func TestMakeCheckAndRehash(t *testing.T) {
	h := New()
	hash, err := h.Make("secret")
	if err != nil {
		t.Fatal(err)
	}
	if !h.Check("secret", hash) {
		t.Fatal("expected match")
	}
	if h.Check("wrong", hash) {
		t.Fatal("expected mismatch")
	}
	if h.NeedsRehash(hash) {
		t.Fatal("fresh hash should not need rehash")
	}
}

func TestFromNil(t *testing.T) {
	if From(nil) != nil {
		t.Fatal("From(nil) must be nil")
	}
}
