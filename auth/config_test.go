package auth

import "testing"

func TestDefaultConfigMustVerifyEmail(t *testing.T) {
	t.Setenv("AUTH_MUST_VERIFY_EMAIL", "false")
	cfg := DefaultConfig()
	if v, _ := cfg["must_verify_email"].(bool); v {
		t.Fatal("default AUTH_MUST_VERIFY_EMAIL must be false")
	}
	t.Setenv("AUTH_MUST_VERIFY_EMAIL", "true")
	cfg = DefaultConfig()
	if v, _ := cfg["must_verify_email"].(bool); !v {
		t.Fatal("AUTH_MUST_VERIFY_EMAIL=true must set auth.must_verify_email")
	}
}
