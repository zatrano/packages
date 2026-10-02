package testing

import (
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"net/url"
	"strings"

	"github.com/zatrano/framework/v3/core/contracts"
	kernelhttp "github.com/zatrano/framework/v3/core/kernel/http"
	"github.com/zatrano/rawhttp"
)

// TestCase wraps an application for HTTP feature tests.
type TestCase struct {
	App     contracts.App
	headers map[string]string
	cookies []*stdhttp.Cookie
}

// New creates a test case from an application.
func New(app contracts.App) (*TestCase, error) {
	if !appIsBooted(app) {
		if err := app.Bootstrap(); err != nil {
			return nil, err
		}
	}
	return &TestCase{
		App:     app,
		headers: map[string]string{},
		cookies: make([]*stdhttp.Cookie, 0),
	}, nil
}

func appIsBooted(app contracts.App) bool {
	// Bootstrap is safe to call only when needed; detect via logger presence.
	return app.Logger() != nil && app.Router() != nil
}

// WithHeader sets a request header for subsequent calls.
func (t *TestCase) WithHeader(key, value string) *TestCase {
	t.headers[key] = value
	return t
}

// WithToken sets a bearer token.
func (t *TestCase) WithToken(token string) *TestCase {
	return t.WithHeader("Authorization", "Bearer "+token)
}

// WithCookie adds a cookie for subsequent requests.
func (t *TestCase) WithCookie(name, value string) *TestCase {
	t.cookies = append(t.cookies, &stdhttp.Cookie{Name: name, Value: value})
	return t
}

// AcceptJSON sets Accept to application/json.
func (t *TestCase) AcceptJSON() *TestCase {
	return t.WithHeader("Accept", "application/json")
}

// Get performs a GET request.
func (t *TestCase) Get(uri string) *TestResponse {
	return t.call(stdhttp.MethodGet, uri, nil, "")
}

// Post performs a POST request with form or JSON body.
func (t *TestCase) Post(uri string, body any) *TestResponse {
	return t.call(stdhttp.MethodPost, uri, body, "")
}

// PostJSON performs a JSON POST request.
func (t *TestCase) PostJSON(uri string, body any) *TestResponse {
	return t.call(stdhttp.MethodPost, uri, body, "application/json")
}

// PutJSON performs a JSON PUT request.
func (t *TestCase) PutJSON(uri string, body any) *TestResponse {
	return t.call(stdhttp.MethodPut, uri, body, "application/json")
}

// Delete performs a DELETE request.
func (t *TestCase) Delete(uri string) *TestResponse {
	return t.call(stdhttp.MethodDelete, uri, nil, "")
}

func (t *TestCase) call(method, uri string, body any, contentType string) *TestResponse {
	var bodyBytes []byte
	switch v := body.(type) {
	case nil:
		bodyBytes = nil
	case string:
		bodyBytes = []byte(v)
	case []byte:
		bodyBytes = v
	case map[string]string:
		if contentType == "application/json" {
			bodyBytes, _ = json.Marshal(v)
		} else {
			form := url.Values{}
			for key, value := range v {
				form.Set(key, value)
			}
			bodyBytes = []byte(form.Encode())
			if contentType == "" {
				contentType = "application/x-www-form-urlencoded"
			}
		}
	default:
		bodyBytes, _ = json.Marshal(v)
		if contentType == "" {
			contentType = "application/json"
		}
	}

	if t.App == nil {
		panic("testing: App is nil")
	}

	headers := map[string]string{}
	for key, value := range t.headers {
		headers[key] = value
	}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	if len(t.cookies) > 0 {
		parts := make([]string, 0, len(t.cookies))
		for _, c := range t.cookies {
			parts = append(parts, c.Name+"="+c.Value)
		}
		headers["Cookie"] = strings.Join(parts, "; ")
	}

	rawReq := buildRawHTTPRequest(method, uri, headers, string(bodyBytes))
	er, err := kernelhttp.ExchangeForTest(func(ctx *rawhttp.Ctx) {
		t.App.Handle(ctx)
	}, rawReq)
	if err != nil {
		panic(fmt.Sprintf("testing: ExchangeForTest: %v", err))
	}

	// Mirror Set-Cookie into jar for follow-up requests.
	for _, v := range er.Header.Values("Set-Cookie") {
		if c := parseSetCookie(v); c != nil {
			t.cookies = append(t.cookies, c)
		}
	}

	return &TestResponse{
		StatusCode: er.Status,
		Headers:    er.Header.Clone(),
		Body:       er.Body,
	}
}

func buildRawHTTPRequest(method, uri string, headers map[string]string, body string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s HTTP/1.1\r\n", method, uri)
	if _, ok := headers["Host"]; !ok {
		b.WriteString("Host: localhost\r\n")
	}
	b.WriteString("Connection: close\r\n")
	for k, v := range headers {
		fmt.Fprintf(&b, "%s: %s\r\n", k, v)
	}
	if body != "" || method == "POST" || method == "PUT" || method == "PATCH" {
		fmt.Fprintf(&b, "Content-Length: %d\r\n", len(body))
	}
	b.WriteString("\r\n")
	b.WriteString(body)
	return b.String()
}

func parseSetCookie(v string) *stdhttp.Cookie {
	parts := strings.Split(v, ";")
	if len(parts) == 0 {
		return nil
	}
	nv := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(nv) != 2 {
		return nil
	}
	return &stdhttp.Cookie{Name: nv[0], Value: nv[1]}
}

// TestResponse asserts against an HTTP response.
type TestResponse struct {
	StatusCode int
	Headers    stdhttp.Header
	Body       []byte
}

// AssertStatus asserts the status code.
func (r *TestResponse) AssertStatus(status int) *TestResponse {
	if r.StatusCode != status {
		panic(newAssertionError("expected status %d, got %d; body=%s", status, r.StatusCode, string(r.Body)))
	}
	return r
}

// AssertOK asserts 200.
func (r *TestResponse) AssertOK() *TestResponse {
	return r.AssertStatus(200)
}

// AssertCreated asserts 201.
func (r *TestResponse) AssertCreated() *TestResponse {
	return r.AssertStatus(201)
}

// AssertUnauthorized asserts 401.
func (r *TestResponse) AssertUnauthorized() *TestResponse {
	return r.AssertStatus(401)
}

// AssertNotFound asserts 404.
func (r *TestResponse) AssertNotFound() *TestResponse {
	return r.AssertStatus(404)
}

// AssertJSONContains asserts a key/value exists in JSON object body.
func (r *TestResponse) AssertJSONContains(key string, expected any) *TestResponse {
	var payload map[string]any
	if err := json.Unmarshal(r.Body, &payload); err != nil {
		panic(newAssertionError("invalid json: %v; body=%s", err, string(r.Body)))
	}
	actual, ok := payload[key]
	if !ok {
		panic(newAssertionError("missing json key %q; body=%s", key, string(r.Body)))
	}
	if stringify(actual) != stringify(expected) {
		panic(newAssertionError("json key %q expected %v, got %v", key, expected, actual))
	}
	return r
}

// AssertSee asserts the body contains needle.
func (r *TestResponse) AssertSee(needle string) *TestResponse {
	if !strings.Contains(string(r.Body), needle) {
		panic(newAssertionError("expected body to contain %q; body=%s", needle, string(r.Body)))
	}
	return r
}

// AssertDontSee asserts the body does not contain needle.
func (r *TestResponse) AssertDontSee(needle string) *TestResponse {
	if strings.Contains(string(r.Body), needle) {
		panic(newAssertionError("expected body not to contain %q; body=%s", needle, string(r.Body)))
	}
	return r
}

// AssertHeader asserts a response header equals value.
func (r *TestResponse) AssertHeader(key, value string) *TestResponse {
	actual := r.Headers.Get(key)
	if actual != value {
		panic(newAssertionError("header %q expected %q, got %q", key, value, actual))
	}
	return r
}

// AssertRedirect asserts a redirect Location.
func (r *TestResponse) AssertRedirect(to string) *TestResponse {
	if r.StatusCode < 300 || r.StatusCode >= 400 {
		panic(newAssertionError("expected redirect status, got %d", r.StatusCode))
	}
	loc := r.Headers.Get("Location")
	if loc != to {
		panic(newAssertionError("expected redirect to %q, got %q", to, loc))
	}
	return r
}

// AssertJSONPath asserts a dotted path in a JSON object.
func (r *TestResponse) AssertJSONPath(path string, expected any) *TestResponse {
	var payload any
	if err := json.Unmarshal(r.Body, &payload); err != nil {
		panic(newAssertionError("invalid json: %v; body=%s", err, string(r.Body)))
	}
	actual, ok := digJSON(payload, strings.Split(path, ".")...)
	if !ok {
		panic(newAssertionError("missing json path %q; body=%s", path, string(r.Body)))
	}
	if stringify(actual) != stringify(expected) {
		panic(newAssertionError("json path %q expected %v, got %v", path, expected, actual))
	}
	return r
}

// JSON decodes the body.
func (r *TestResponse) JSON(dest any) error {
	return json.Unmarshal(r.Body, dest)
}

// String returns the body as string.
func (r *TestResponse) String() string {
	return string(r.Body)
}

func digJSON(value any, parts ...string) (any, bool) {
	current := value
	for _, part := range parts {
		if part == "" {
			continue
		}
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, ok := obj[part]
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}

type assertionError string

func (e assertionError) Error() string { return string(e) }

func newAssertionError(format string, args ...any) assertionError {
	return assertionError(fmt.Sprintf(format, args...))
}

func stringify(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(raw)
}
