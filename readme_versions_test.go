package packages

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReadmeVersions(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, piece := range []string{
		"releases/tag/v3.2.3",
		"releases/tag/v1.0.1",
		"releases/tag/v1.15.3",
		"github.com/zatrano/framework/v3 v3.1.2",
		"github.com/zatrano/rawhttp v1.0.0",
		"github.com/zatrano/canvas v1.0.1",
	} {
		if !strings.Contains(text, piece) {
			t.Fatalf("README is missing %s", piece)
		}
	}
	if strings.Contains(text, "v0.2.5") {
		t.Fatal("README still names rawhttp v0.2.5")
	}
	if strings.Contains(text, "v3.1.2</code>") {
		t.Fatal("current-release line still names framework v3.1.2")
	}
}
