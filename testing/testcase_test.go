package testing

import (
	stdhttp "net/http"
	"testing"
)

func TestResponseAssertions(t *testing.T) {
	r := &TestResponse{
		StatusCode: 200,
		Headers:    stdhttp.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"ok":true,"nested":{"n":1}}`),
	}
	r.AssertOK().AssertSee("ok").AssertDontSee("fail").AssertHeader("Content-Type", "application/json")
	r.AssertJSONContains("ok", true)
	r.AssertJSONPath("nested.n", float64(1))
}

func TestFromNilAppPanicsOnCall(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	tc := &TestCase{}
	tc.Get("/")
}
