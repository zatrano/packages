package packages

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
	root := filepath.Dir(thisFile)
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
	if !strings.Contains(text, "github.com/zatrano/framework/v3 v3.0.0") {
		t.Fatal("go.mod must require github.com/zatrano/framework/v3 v3.0.0")
	}
	if !strings.Contains(text, "github.com/zatrano/canvas v0.2.0") {
		t.Fatal("go.mod must require github.com/zatrano/canvas v0.2.0")
	}
	if !strings.Contains(text, "github.com/zatrano/rawhttp v0.2.2") {
		t.Fatal("go.mod must require github.com/zatrano/rawhttp v0.2.2")
	}
	if strings.Contains(text, "replace ") {
		t.Fatal("go.mod must not use replace; depend on published module versions")
	}
	if i := strings.Index(text, "require ("); i >= 0 {
		if j := strings.Index(text[i:], "\n)"); j >= 0 {
			block := text[i : i+j]
			if strings.Contains(block, "github.com/zatrano/packages/database/driver/") {
				t.Fatal("root require must not include nested database drivers")
			}
			if strings.Contains(block, "github.com/zatrano/packages/orm") {
				t.Fatal("root require must not include removed packages/orm")
			}
		}
	}
}

func TestPublicCanvasAndRawHTTPResolveWithoutSibling(t *testing.T) {
	dir := t.TempDir()
	mod := "module consumer.test\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, spec := range []string{
		"github.com/zatrano/canvas@v0.2.0",
		"github.com/zatrano/rawhttp@v0.2.2",
	} {
		cmd := exec.Command("go", "get", spec)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOWORK=off")
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := string(out)
			if strings.Contains(msg, "checksum mismatch") || strings.Contains(msg, "invalid version") || strings.Contains(msg, "404") {
				t.Skipf("public module proxy has not ingested %s yet: %s", spec, msg)
			}
			t.Fatalf("public resolve %s failed: %v\n%s", spec, err, out)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if strings.Contains(text, "replace ") {
		t.Fatalf("public consumer must not use replace:\n%s", text)
	}
	if !strings.Contains(text, "github.com/zatrano/canvas v0.2.0") {
		t.Fatalf("expected canvas v0.2.0:\n%s", text)
	}
	if !strings.Contains(text, "github.com/zatrano/rawhttp v0.2.2") {
		t.Fatalf("expected rawhttp v0.2.2:\n%s", text)
	}
}
