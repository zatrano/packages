package auth

import (
	"context"
	"time"

	"github.com/zatrano/framework/v2/kernel/http"
)

// Publisher receives authentication Facts. Bind facts.From(app) when the facts package is enabled.
type Publisher interface {
	Publish(ctx context.Context, fact any) error
}

// Occurrence is shared request/user context on auth Facts. It is business data, not infrastructure metadata.
type Occurrence struct {
	Request     *http.Request
	User        Authenticatable
	Credentials map[string]string
	Guard       string
	At          time.Time
}

type UserAttempting struct{ Occurrence }
type UserFailed struct{ Occurrence }
type UserLoggedIn struct{ Occurrence }
type UserLoggedOut struct{ Occurrence }
type UserRegistered struct{ Occurrence }
type EmailVerified struct{ Occurrence }
type PasswordReset struct{ Occurrence }
type CurrentDeviceLogout struct{ Occurrence }
type OtherDeviceLogout struct{ Occurrence }
type UserLockedOut struct{ Occurrence }
type TwoFactorChallenged struct{ Occurrence }
type TwoFactorAuthenticated struct{ Occurrence }

func factContext(req *http.Request) context.Context {
	if req != nil && req.Raw() != nil {
		return req.Raw().Context()
	}
	return context.Background()
}

func occur(req *http.Request, user Authenticatable, credentials map[string]string, guard string) Occurrence {
	return Occurrence{Request: req, User: user, Credentials: credentials, Guard: guard, At: time.Now().UTC()}
}
