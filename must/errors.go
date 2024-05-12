package must

import (
    "errors"
    "fmt"
)



// ResultError is the type of error that may be returned by [Result] or
// [Resultf].
//
// It is generic in order to capture the information about the type of the
// hoped result without having to allocate memory at the error creation time.
type ResultError[T any] struct {
    fmtErr, err error
}



// Error implements the standard error interface.
func (e ResultError[T]) Error() string {
    var t T
    return fmt.Sprintf("must.ResultError[%T]: %v", t, e.Unwrap())
}

// Unwrap (for use with [errors.Is], etc.) the original error caught by
// [Result], or the original error and the formatted error message received by
// [Resultf], in which case the returned error implements `[]Unwrap()`.
func (e ResultError[T]) Unwrap() error {
    if e.fmtErr != nil {
        return errors.Join(e.fmtErr, e.err)
    }
    return e.err
}



// OkError is the type of error that may be returned by [Ok] or [Okf].
//
// It is generic in order to capture the information about the type of the
// hoped result without having to allocate memory at the error creation time.
type OkError[T any] struct {
    fmtErr error
}

// Error implements the standard error interface.
func (e OkError[T]) Error() string {
    var t T
    if e.fmtErr == nil {
        return fmt.Sprintf("must.OkError[%T]", t)
    } else {
        return fmt.Sprintf("must.OkError[%T]: %s", t, e.fmtErr)
    }
}

// Unwrap (for use with [errors.Is], etc.) returns nil for an error returned
// by [Ok], or the formatted error message received by [Okf].
func (e OkError[T]) Unwrap() error {
    return e.fmtErr // may be nil
}


// CompareError is the type of error that may be returned by a comparison
// function such as [True] or [Equal], or a formating variant of a comparison
// function (one ending in "f").
type CompareError[T comparable] struct {
    operation, a, b string
    fmtErr error
}

func newCompareError[T comparable](operation string, a T, b T, err error) CompareError[T] {
    fmtComparable := func(x T) string {
        if interface{}(&x) == nil {
            return "nil"
        }
        switch v := interface{}(&x).(type) {
            case string: return fmt.Sprintf("%q", v)
        }
        if stringer, ok := interface{}(&x).(interface{String() string}); ok {
            return stringer.String()
        }
        return fmt.Sprintf("%v", x)
    }
    return CompareError[T]{
        operation: operation,
        a:         fmtComparable(a),
        b:         fmtComparable(b),
        fmtErr:    err,
    }
}

// Error implements the standard error interface.
func (e CompareError[T]) Error() string {
    var t T
    if e.fmtErr == nil {
        return fmt.Sprintf("must.CompareError[%T]<%s>(%q, %q)",
            t, e.operation, e.a, e.b)
    } else {
        return fmt.Sprintf("must.CompareError[%T]<%s>(%q, %q): %s",
            t, e.operation, e.a, e.b, e.fmtErr)
    }
}

// Unwrap (for use with [errors.Is], etc.) returns nil for an error returned
// by a comparison function, or the formatted error message received by a
// formating variant of a comparison function (one ending in "f").
func (e CompareError[T]) Unwrap() error {
    return e.fmtErr // may be nil
}


// CheckError is the type of error that may be returned by a check function
// such as [Check], [CheckAll], and [Checkf].
type CheckError struct {
    fmtErr, err error
}

// Error implements the standard error interface.
func (e CheckError) Error() string {
    return fmt.Sprintf("must.CheckError: %v", e.Unwrap())
}

// Unwrap (for use with [errors.Is], etc.) the original error caught by a check
// function, or the original error and the formatted error message received by
// formatting variant of a check function (one ending in "f"), in which case
// the returned error implements `[]Unwrap()`.
func (e CheckError) Unwrap() error {
    if e.fmtErr != nil {
        return errors.Join(e.fmtErr, e.err)
    }
    return e.err
}



// ValueError is the type of error that may hold a value, for example
// a non-error value recovered from a panic in [Try].
//
// It is generic in order to capture the information about the type of the
// hoped result without having to allocate memory at the error creation time.
//
// Note that the value itself is type `any`, not one that matches the generic
// type constraint.
type ValueError[T any] struct {
    value any
}

func (e ValueError[T]) Value() any {
    return e.value
}

// Error implements the standard error interface.
func (e ValueError[T]) Error() string {
    var t T
    return fmt.Sprintf("must.ValueError[%T]: %v", t, e.value)
}


// NeverError is the type of error that may be returned by [Never], and
// represents an event that should never happen.
type NeverError struct {
    fmtErr error
}

// Error implements the standard error interface.
func (e NeverError) Error() string {
    if e.fmtErr == nil {
        return "must.Never: this should never happen"
    } else {
        return fmt.Sprintf("must.Never: %s", e.fmtErr)
    }
}

// Unwrap (for use with [errors.Is], etc.) returns nil for an error returned
// by [Never], or the formatted error message received by [Neverf].
func (e NeverError) Unwrap() error {
    return e.fmtErr // may be nil
}
