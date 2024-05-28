// Package must implements assertions.
package must

import (
    "fmt"
)

func typeName(x any) string {
    // while reflect.TypeOf(x).String() is more efficient, the reflect package
    // is less portable.
    return fmt.Sprintf("%T", x)
}

// Result accepts a (value, err) tuple as input and panics if err != nil,
// otherwise returns value. The non-nil input error is wrapped in [ResultError]
// before panicking.
//
// For example, must.Result(os.Open("doesnotexist")) may panic with an error
// such as "error in must.Result[*os.File]: open doesnotexist: no such file or
// directory", or, on success, return *os.File.
func Result[T any](t T, err error) T {
    if err != nil {
        panic(ResultError[T]{nil, err})
    }
    return t
}

// Resultf accepts a (value, err) tuple, and a [fmt.Sprintf] -style format
// string with optional arguments, as input and panics if err != nil. Otherwise,
// returns value. The original non-nil input error, and the formatted error
// message, are joined (see [errors.Join] and wrapped in [ResultError] before
// panicking.
func Resultf[T any](t T, err error, format string, args ... any) T {
    if err != nil {
        panic(ResultError[T]{fmt.Errorf(format, args...), err})
    }
    return t
}

// Ok accepts a (value, ok) tuple as input and panics if ok is false, otherwise
// returns value.
func Ok[T any](t T, ok bool) T {
    if ok { return t }
    panic(OkError[T]{})
}

// Okf accepts a (value, ok) tuple, and a [fmt.Sprintf] -style format string
// with optional arguments, as input and panics if ok is false. Otherwise,
// returns value. The formatted error message is wrapped in [OkError] before
// panicking.
func Okf[T any](t T, ok bool, format string, args ... any) T {
    if ok { return t }
    panic(OkError[T]{fmt.Errorf(format, args...)})
}

// Equal panics if the provided comparable values are not equal. Otherwise,
// returns true.
func Equal[T comparable](a T, b T) bool {
    if a == b { return true }
    panic(newCompareError[T]("Equal", a, b, nil))
}

// Equalf panics if the provided comparable values are not equal. Otherwise,
// returns true.
//
// The [fmt.Sprintf] -style format string and with optional arguments are used
// to format the error message. The formatted error message is wrapped in
// [CompareError] before panicking.
func Equalf[T comparable](a T, b T, format string, args ... any) bool {
    if a == b { return true }
    err := fmt.Errorf(format, args...)
    panic(newCompareError[T]("Equal", a, b, err))
}

// True is the equivalent of Equal with a true value.
func True(q bool) bool {
    return Equal(q, true)
}

// Truef is the equivalent of Equalf with a true value.
func Truef(q bool, format string, args ... any) bool {
    return Equalf(q, true, format, args...)
}

// Not is the equivalent of Equal with a false value.
func Not(q bool) bool {
    return Equal(q, false)
}

// False is an alias of [Not].
var False = Not

// Notf is the equivalent of Equalf with a false value.
func Notf(q bool, format string, args ... any) bool {
    return Equalf(q, false, format, args...)
}

// Falsef is an alias of [Notf].
var Falsef = Notf

// Check panics if the error is not nil. Otherwise, it returns a nil error (so
// that it is convenient to chain). The raised non-nil error is wrapped
// in a [CheckError] before panicking.
func Check(err error) error {
    if err == nil { return nil }
    panic(CheckError{nil, err})
}

// Checkf panics if the error is not nil. Otherwise, it always returns a nil
// error (so that it is convenient to chain).
//
// The [fmt.Sprintf] -style format string and with optional arguments are used
// to format the error message raised in the panic. The original non-nil input
// error, and the formatted error message, are joined (see [errors.Join] and
// wrapped in [CheckError] before panicking.
func Checkf(err error, format string, args ... any) error {
    if err == nil { return nil }
    panic(CheckError{fmt.Errorf(format, args...), err})
}

// CheckAll panics at the first non-nil error. The raised non-nil error is
// wrapped in a [CheckError] before panicking.
func CheckAll(errs ... error) {
    for _, err := range errs {
        Check(err)
    }
}

// Try takes a function f() => x that may panic, and instead returns a
// function f() => (x, error).
//
// If the raised panic is of type error, it is returned directly. Otherwise,
// it is wrapped in a [ValueError].
func Try[X any](f func() X) func() (x X, err error) {
    return func() (x X, err error) {
        defer func() {
            if r := recover(); r != nil {
                if rErr, ok := r.(error); ok {
                    err = rErr
                } else {
                    err = ValueError[X]{r}
                }
            }
        }()

        return f(), nil
    }
}

// Func takes a function f() => (x, error), and returns a function f() => x
// that may panic in the event of error.
//
// Any raised error is wrapped in [ResultError].
func Func[X any](
    f func () (X, error),
) func () X {
    return func() X {
        return Result(f())
    }
}

// Never signifies code that should never be reached. It raises a panic when
// called.
func Never() {
    panic(NeverError{})
}

// Neverf signifies code that should never be reached. It raises a panic when
// called.
//
// The args parameter defines an optional fmt.Sprintf-style format string and
// arguments.
func Neverf(format string, args ... any) {
    panic(NeverError{fmt.Errorf(format, args...)})
}
