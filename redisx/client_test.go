package redisx

import "testing"

func TestParseDB(t *testing.T) {
	if ParseDB("2") != 2 {
		t.Fatal("expected 2")
	}
	if ParseDB("x") != 0 {
		t.Fatal("invalid should be 0")
	}
}

func TestClientFrom(t *testing.T) {
	if ClientFrom("no") != nil {
		t.Fatal("expected nil")
	}
	if ClientFrom(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestConnectRejectsClosedPort(t *testing.T) {
	_, err := Connect(Config{Host: "127.0.0.1", Port: "1"})
	if err == nil {
		t.Fatal("expected connect error")
	}
}
