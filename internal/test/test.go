package test

import (
    "errors"
    "testing"
    "time"
)

// Panics returns true iff calling f panics with an error and either target
// is nil or errors.Is(error, target) is true.
func Panics(t *testing.T, f func(), target error) (result bool) {
    t.Helper()
    defer func() {
        if r := recover(); r == nil {
            result = false
        } else if err, ok := r.(error); ok {
            result = (target == nil) || errors.Is(err, target)
        } else {
            result = (target == nil)
        }
    }()
    f()
    return false
}

// Completes executes f (in a goroutine), and blocks until either f returns,
// or the provided duration has elapsed. In the latter case, calls t.Errorf to
// fail the test. Provide optional format string and arguments to add
// context to the test error message.
func Completes(t *testing.T, duration time.Duration, f func(), args ... interface{}) {
    done := make(chan struct{}, 1)
    timeout := time.After(duration)
    go func() {
        f()
        done <- struct{}{}
    }()

    select {
        case <-done: // OK
        case <-timeout:
            if len(args) > 0 {
                t.Errorf("test timed out after "+duration.String()+": " + args[0].(string), args[1:]...)
            } else {
                t.Errorf("test timed out after %s", duration.String())
            }
    }
}
