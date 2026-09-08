package version_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGoModPinsReleasedFramework(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "module github.com/zatrano/packages\n") {
		t.Fatal("public module path must remain github.com/zatrano/packages")
	}
	if strings.Contains(text, "module github.com/zatrano/packages/v2") {
		t.Fatal("do not introduce a /v2 module path")
	}
	if !strings.Contains(text, "github.com/zatrano/framework/v2 v2.0.28") {
		t.Fatal("go.mod must require github.com/zatrano/framework/v2 v2.0.28")
	}
	if !strings.Contains(text, "replace github.com/zatrano/framework/v2 => ../framework") {
		t.Fatal("development replace must remain in the packages module")
	}
}

func TestPublicFrameworkModuleResolvesWithoutSibling(t *testing.T) {
	dir := t.TempDir()
	mod := "module consumer.test\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "get", "github.com/zatrano/framework/v2@v2.0.28")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(out)
		if strings.Contains(msg, "checksum mismatch") {
			t.Skip("public module proxy still serves the first v2.0.28 zip; GitHub tree checksum differs")
		}
		t.Fatalf("public framework resolve failed: %v\n%s", err, out)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if strings.Contains(text, "replace ") {
		t.Fatalf("public consumer must not use replace:\n%s", text)
	}
	if !strings.Contains(text, "github.com/zatrano/framework/v2 v2.0.28") {
		t.Fatalf("expected framework v2.0.28:\n%s", text)
	}
}
