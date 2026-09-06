package markdown_test

import (
	"strings"
	"testing"

	"github.com/zatrano/packages/markdown"
)

func TestMarkdownToHTML(t *testing.T) {
	html := markdown.ToHTML("# Hello\n\nThis is **bold** and a [link](https://zatrano.test).")
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Fatalf("missing h1: %s", html)
	}
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Fatalf("missing bold: %s", html)
	}
	if !strings.Contains(html, `<a href="https://zatrano.test">link</a>`) {
		t.Fatalf("missing link: %s", html)
	}
}

func TestInlineCodeWithSnakeEnvNames(t *testing.T) {
	html := markdown.ToHTML("when `CACHE_STORE=redis` or `QUEUE_CONNECTION=redis`. There is **no** `redisx.From(app)`.")
	if strings.Contains(html, "<em>") {
		t.Fatalf("underscore italic leaked into env code: %s", html)
	}
	if !strings.Contains(html, "<code>CACHE_STORE=redis</code>") {
		t.Fatalf("missing CACHE_STORE code: %s", html)
	}
	if !strings.Contains(html, "<code>QUEUE_CONNECTION=redis</code>") {
		t.Fatalf("missing QUEUE_CONNECTION code: %s", html)
	}
	if !strings.Contains(html, "<strong>no</strong>") {
		t.Fatalf("missing bold: %s", html)
	}
}

func TestLinkWithInnerCode(t *testing.T) {
	html := markdown.ToHTML("Enable this together with [`cache`](/docs/cache) or [`queue`](/docs/queue).")
	if !strings.Contains(html, `<a href="/docs/cache"><code>cache</code></a>`) {
		t.Fatalf("broken cache link: %s", html)
	}
	if !strings.Contains(html, `<a href="/docs/queue"><code>queue</code></a>`) {
		t.Fatalf("broken queue link: %s", html)
	}
	if strings.Contains(html, "[<code>") || strings.Contains(html, "</code>](") {
		t.Fatalf("raw markdown leaked: %s", html)
	}
}

func TestInlineEscapesAroundCode(t *testing.T) {
	html := markdown.ToHTML("Use `x` if a < b.")
	if !strings.Contains(html, "a &lt; b") {
		t.Fatalf("missing escape: %s", html)
	}
}

func TestFlankingUnderscoreItalic(t *testing.T) {
	html := markdown.ToHTML("Say _hello_ but keep CACHE_STORE intact.")
	if !strings.Contains(html, "<em>hello</em>") {
		t.Fatalf("missing flanking italic: %s", html)
	}
	if strings.Contains(html, "<em>STORE</em>") || strings.Contains(html, "<em>CACHE</em>") {
		t.Fatalf("snake case was italicized: %s", html)
	}
}

func TestFlankingStarItalic(t *testing.T) {
	html := markdown.ToHTML("Say *hello* then `DB_*` and `MAIL_*` plus DB_* MAIL_*.")
	if !strings.Contains(html, "<em>hello</em>") {
		t.Fatalf("missing star italic: %s", html)
	}
	if strings.Count(html, "<em>") != 1 {
		t.Fatalf("expected one em, got: %s", html)
	}
	if !strings.Contains(html, "<code>DB_*</code>") || !strings.Contains(html, "<code>MAIL_*</code>") {
		t.Fatalf("missing glob code: %s", html)
	}
}

func TestMarkdownFencedCodeAndTable(t *testing.T) {
	html := markdown.ToHTML("```go\nfmt.Println(\"hi\")\n```\n\n| Key | Purpose |\n|-----|---------|\n| `app.name` | Name |")
	if !strings.Contains(html, `<pre><code class="language-go">`) {
		t.Fatalf("missing fenced code: %s", html)
	}
	if !strings.Contains(html, "fmt.Println") {
		t.Fatalf("missing code body: %s", html)
	}
	if !strings.Contains(html, "<table>") || !strings.Contains(html, "<th>Key</th>") {
		t.Fatalf("missing table: %s", html)
	}
}
