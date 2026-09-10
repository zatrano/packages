package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zatrano/framework/v2/kernel"
)

func TestToRuleName(t *testing.T) {
	if got := toRuleName("UniqueEmailRule"); got != "unique_email" {
		t.Fatalf("got %q", got)
	}
}

func TestRequestTypeName(t *testing.T) {
	cases := map[[2]string]string{
		{"Post", "store"}:      "PostStoreRequest",
		{"PostStore", "store"}: "PostStoreRequest",
		{"Post", "update"}:     "PostUpdateRequest",
		{"Post", "index"}:      "PostIndexRequest",
		{"Post", ""}:           "PostRequest",
		{"Login", ""}:          "LoginRequest",
	}
	for in, want := range cases {
		if got := requestTypeName(in[0], in[1]); got != want {
			t.Errorf("requestTypeName(%q, %q)=%q want %q", in[0], in[1], got, want)
		}
	}
}

func TestRequestStubIndexRules(t *testing.T) {
	body := requestStub("PostIndexRequest", "index")
	if !strings.Contains(body, `"q"`) || strings.Contains(body, `"name": "required`) {
		t.Fatalf("index stub:\n%s", body)
	}
}

func TestMakeRequestStoreFlag(t *testing.T) {
	dir := t.TempDir()
	cmd := &MakeRequestCommand{app: kernel.NewApplication(dir)}
	if err := cmd.Handle([]string{"Post", "--store"}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "app", "http", "requests", "post_store_request.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "type PostStoreRequest struct") {
		t.Fatalf("got:\n%s", text)
	}
}

func TestCommandsRegisterRequestAndRule(t *testing.T) {
	cmds := Commands(nil)
	if len(cmds) != 2 {
		t.Fatalf("got %d commands", len(cmds))
	}
	names := map[string]bool{}
	for _, c := range cmds {
		names[c.Name] = true
	}
	if !names["make:request"] || !names["make:rule"] {
		t.Fatalf("got %#v", names)
	}
}
