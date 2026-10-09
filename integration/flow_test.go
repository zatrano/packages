package integration

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/zatrano/framework/v3/core/kernel"
	khttp "github.com/zatrano/framework/v3/core/kernel/http"
	"github.com/zatrano/framework/v3/core/kernel/middleware"
	"github.com/zatrano/framework/v3/core/kernel/routing"
	stdhttp "net/http"

	"github.com/zatrano/packages/broadcasting"
	"github.com/zatrano/packages/flash"
	"github.com/zatrano/packages/queue"
	"github.com/zatrano/packages/ratelimit"
	"github.com/zatrano/packages/validation"
	"github.com/zatrano/packages/websocket"
)

func TestFlowsOnce(t *testing.T) {
	app := bootApp(t)
	if err := exercise(app, "once-token"); err != nil {
		t.Fatal(err)
	}
}

func bootApp(t *testing.T) *kernel.Application {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_DEBUG", "false")
	t.Setenv("APP_KEY", "zatrano-ci-test-key-not-prod!!!!")
	t.Setenv("LOG_LEVEL", "error")
	t.Setenv("HTTP_POOL_REQUESTS", "true")
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "public"), 0o755); err != nil {
		t.Fatal(err)
	}
	app := kernel.NewApplication(dir)
	t.Cleanup(func() {
		if c, ok := app.Logger().(interface{ Close() error }); ok && c != nil {
			_ = c.Close()
		}
	})
	registerFlows(app)
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	return app
}

type memSession struct {
	mu   sync.Mutex
	data map[string]any
}

func (s *memSession) Get(key string, fallback ...any) any {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.data[key]; ok {
		return v
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return nil
}
func (s *memSession) Put(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}
func (s *memSession) Flash(key string, value any) { s.Put(key, value) }
func (s *memSession) Pull(key string, fallback ...any) any {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	if ok {
		delete(s.data, key)
		return v
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return nil
}
func (s *memSession) Forget(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}
func (s *memSession) Regenerate() error { return nil }
func (s *memSession) ID() string        { return "mem" }

var (
	sessions sync.Map
	jobs     sync.Map
	limiter  = ratelimit.New()
	channels = broadcasting.NewChannelRegistry()
	memQ     = queue.NewSyncQueue()
)

func init() {
	memQ.Register("work", func(payload map[string]any) error {
		token, _ := payload["token"].(string)
		jobs.Store(token, token)
		return nil
	})
	channels.Channel("room.*", func(req *khttp.Request, channel string) bool {
		return channel == "room."+req.Header("X-Token")
	})
}

func sessionFor(id string) *memSession {
	if id == "" {
		id = "anon"
	}
	v, _ := sessions.LoadOrStore(id, &memSession{data: map[string]any{}})
	return v.(*memSession)
}

func bindSession(next routing.HandlerFunc) routing.HandlerFunc {
	return func(req *khttp.Request) *khttp.Response {
		req.SetSession(sessionFor(req.Cookie("sid")))
		return next(req)
	}
}

func registerFlows(app *kernel.Application) {
	r := app.Router()
	r.Post("/login", func(req *khttp.Request) *khttp.Response {
		user := req.Header("X-User")
		resp := khttp.Text("user=" + user)
		resp.WithCookie(&stdhttp.Cookie{Name: "user", Value: user, Path: "/"})
		return resp
	})
	r.Get("/me", func(req *khttp.Request) *khttp.Response {
		return khttp.Text("user=" + req.Cookie("user"))
	})
	r.Get("/csrf", bindSession(func(req *khttp.Request) *khttp.Response {
		return khttp.Text(middleware.CSRFToken(req))
	}))
	r.Post("/csrf", bindSession(middleware.CSRF(func(req *khttp.Request) *khttp.Response {
		return khttp.Text("field=" + req.Input("field"))
	})))
	limited := ratelimit.ThrottleRequestsWith(limiter, 1000, time.Minute, func(req *khttp.Request) string {
		return "hdr:" + req.Header("X-Rate-Key")
	})
	r.Get("/limited", limited(func(req *khttp.Request) *khttp.Response {
		return khttp.Text("key=" + req.Header("X-Rate-Key"))
	}))
	r.Get("/tenant", func(req *khttp.Request) *khttp.Response {
		return khttp.Text("host=" + req.Host())
	})
	r.Post("/upload", func(req *khttp.Request) *khttp.Response {
		file, err := req.File("doc")
		if err != nil {
			return khttp.Text("err=" + err.Error())
		}
		f, err := file.File()
		if err != nil {
			return khttp.Text("err=" + err.Error())
		}
		defer f.Close()
		raw, _ := io.ReadAll(f)
		return khttp.Text("file=" + file.Name() + " body=" + string(raw))
	})
	r.Post("/job", func(req *khttp.Request) *khttp.Response {
		token := req.Header("X-Token")
		if err := memQ.Push(queue.NamedJob{Name: "work", Payload: map[string]any{"token": token}}); err != nil {
			return khttp.Text("err=" + err.Error())
		}
		return khttp.Text("queued=" + token)
	})
	r.Get("/event", func(req *khttp.Request) *khttp.Response {
		token := req.Header("X-Token")
		ok, err := channels.Authorize(req, "room."+token)
		if err != nil || !ok {
			return khttp.Text("deny")
		}
		return khttp.Text("event=" + token)
	})
	r.Get("/ws", websocket.Upgrade(func(conn *websocket.Conn) error {
		return conn.Close(1000)
	}))
	r.Post("/flash", bindSession(func(req *khttp.Request) *khttp.Response {
		token := req.Header("X-Token")
		flash.Success(req, token)
		flash.Old(req, map[string]string{"name": token})
		old := flash.OldInput(req)
		return khttp.Text("flash=" + flash.Get(req, flash.KeySuccess) + " old=" + old["name"])
	}))
	r.Post("/validate", func(req *khttp.Request) *khttp.Response {
		token := req.Header("X-Token")
		email := req.Input("email")
		v := validation.Make(map[string]string{"email": email}, map[string]string{"email": "required|email"})
		_ = v.Fails()
		bag := validation.NewMessageBag(v.Errors())
		return khttp.Text("token=" + token + " email=" + email + " err=" + bag.First("email"))
	})
}

func exercise(app *kernel.Application, token string) error {
	status, header, body, err := Exchange(app, "POST /login HTTP/1.1\r\nHost: app.example\r\nX-User: "+token+"\r\nContent-Length: 0\r\n\r\n")
	if err != nil || status != 200 || body != "user="+token {
		return fmt.Errorf("login status=%d body=%q err=%v", status, body, err)
	}
	if header.Get("Set-Cookie") == "" || !containsToken(header.Get("Set-Cookie"), token) {
		return fmt.Errorf("login cookie %q", header.Get("Set-Cookie"))
	}
	status, _, body, err = Exchange(app, "GET /me HTTP/1.1\r\nHost: app.example\r\nCookie: user="+token+"\r\n\r\n")
	if err != nil || status != 200 || body != "user="+token {
		return fmt.Errorf("me status=%d body=%q err=%v", status, body, err)
	}

	status, _, body, err = Exchange(app, "GET /csrf HTTP/1.1\r\nHost: app.example\r\nCookie: sid="+token+"\r\n\r\n")
	if err != nil || status != 200 || body == "" {
		return fmt.Errorf("csrf token status=%d body=%q err=%v", status, body, err)
	}
	form := "field=" + token
	raw := "POST /csrf HTTP/1.1\r\nHost: app.example\r\nCookie: sid=" + token + "\r\nContent-Type: application/x-www-form-urlencoded\r\nX-CSRF-TOKEN: " + body + "\r\nContent-Length: " + fmt.Sprint(len(form)) + "\r\n\r\n" + form
	status, _, got, err := Exchange(app, raw)
	if err != nil || status != 200 || got != "field="+token {
		return fmt.Errorf("csrf post status=%d body=%q err=%v", status, got, err)
	}

	status, _, body, err = Exchange(app, "GET /limited HTTP/1.1\r\nHost: app.example\r\nX-Rate-Key: "+token+"\r\n\r\n")
	if err != nil || status != 200 || body != "key="+token {
		return fmt.Errorf("limit status=%d body=%q err=%v", status, body, err)
	}
	status, _, body, err = Exchange(app, "GET /tenant HTTP/1.1\r\nHost: "+token+".example\r\n\r\n")
	if err != nil || status != 200 || body != "host="+token+".example" {
		return fmt.Errorf("tenant status=%d body=%q err=%v", status, body, err)
	}

	up := multipartBody(token)
	raw = "POST /upload HTTP/1.1\r\nHost: app.example\r\nContent-Type: " + up.ctype + "\r\nContent-Length: " + fmt.Sprint(len(up.body)) + "\r\n\r\n" + string(up.body)
	status, _, body, err = Exchange(app, raw)
	if err != nil || status != 200 || body != "file="+token+".txt body="+token {
		return fmt.Errorf("upload status=%d body=%q err=%v", status, body, err)
	}

	status, _, body, err = Exchange(app, "POST /job HTTP/1.1\r\nHost: app.example\r\nX-Token: "+token+"\r\nContent-Length: 0\r\n\r\n")
	if err != nil || status != 200 || body != "queued="+token {
		return fmt.Errorf("job status=%d body=%q err=%v", status, body, err)
	}
	if _, ok := jobs.Load(token); !ok {
		return fmt.Errorf("job payload missing %s", token)
	}
	status, _, body, err = Exchange(app, "GET /event HTTP/1.1\r\nHost: app.example\r\nX-Token: "+token+"\r\n\r\n")
	if err != nil || status != 200 || body != "event="+token {
		return fmt.Errorf("event status=%d body=%q err=%v", status, body, err)
	}

	status, _, body, err = Exchange(app, "GET /ws HTTP/1.1\r\nHost: app.example\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n")
	if err != nil || status != 101 {
		return fmt.Errorf("ws status=%d body=%q err=%v", status, body, err)
	}

	status, _, body, err = Exchange(app, "POST /flash HTTP/1.1\r\nHost: app.example\r\nCookie: sid=flash-"+token+"\r\nX-Token: "+token+"\r\nContent-Length: 0\r\n\r\n")
	if err != nil || status != 200 || body != "flash="+token+" old="+token {
		return fmt.Errorf("flash status=%d body=%q err=%v", status, body, err)
	}
	email := token + "@ex.test"
	form = "email=" + email
	raw = "POST /validate HTTP/1.1\r\nHost: app.example\r\nX-Token: " + token + "\r\nContent-Type: application/x-www-form-urlencoded\r\nContent-Length: " + fmt.Sprint(len(form)) + "\r\n\r\n" + form
	status, _, body, err = Exchange(app, raw)
	if err != nil || status != 200 || !strings.Contains(body, "token="+token) || !strings.Contains(body, "email="+email) {
		return fmt.Errorf("validate status=%d body=%q err=%v", status, body, err)
	}
	return nil
}

func containsToken(s, token string) bool { return strings.Contains(s, token) }

type uploadPart struct {
	ctype string
	body  []byte
}

func multipartBody(token string) uploadPart {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("doc", token+".txt")
	if err != nil {
		panic(err)
	}
	if _, err := fw.Write([]byte(token)); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return uploadPart{ctype: w.FormDataContentType(), body: buf.Bytes()}
}
