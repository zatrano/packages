package db

import (
	"errors"
	"fmt"
)

// NotOpenError is returned when operations run against a closed or nil handle.
type NotOpenError struct{}

func (NotOpenError) Error() string { return "db: not open" }

// IsNotOpen reports whether err is or wraps NotOpenError.
func IsNotOpen(err error) bool {
	var n NotOpenError
	return errors.As(err, &n)
}

// OpError is a failed database infrastructure operation (connect, ping, begin, …).
// Query/SQL errors from application code are not wrapped here — they stay native.
type OpError struct {
	Adapter string // postgres, mysql, …
	Op      string // open, ping, begin, …
	Err     error
}

func (e *OpError) Error() string {
	if e == nil {
		return "db: op error"
	}
	if e.Adapter == "" {
		return fmt.Sprintf("db: %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("%s: %s: %v", e.Adapter, e.Op, e.Err)
}

func (e *OpError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// WrapOp returns an *OpError (nil err → nil).
func WrapOp(adapter, op string, err error) error {
	if err == nil {
		return nil
	}
	return &OpError{Adapter: adapter, Op: op, Err: err}
}
