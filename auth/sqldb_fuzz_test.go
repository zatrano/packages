package auth

import "testing"

func FuzzSanitizeIdentifier(f *testing.F) {
	f.Add("id")
	f.Add("users")
	f.Add("tokenable_id")
	f.Add("")
	f.Add("a;drop")
	f.Add("1bad")
	f.Add("_ok")
	f.Fuzz(func(t *testing.T, name string) {
		err := requireIdent(name)
		ok := safeIdent.MatchString(name)
		if ok && err != nil {
			t.Fatalf("valid ident %q rejected: %v", name, err)
		}
		if !ok && err == nil {
			t.Fatalf("invalid ident %q accepted", name)
		}
	})
}
