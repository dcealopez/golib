// Package ks ("kitchen sink") implements assorted helpful things that don't
// fit anywhere else.
package ks

import (
    "encoding/binary"
    "errors"
    "hash/crc64"
    "slices"
    "strings"
    "unicode/utf8"

    "golang.org/x/exp/utf8string"
)

var ErrTODO = errors.New("TODO")

var crc64table = crc64.MakeTable(crc64.ECMA)

func Checksum64(crc uint64, value uint64) uint64 {
    var buf [8]byte
    binary.LittleEndian.PutUint64(buf[0:8], value)
    return crc64.Update(crc, crc64table, buf[0:8])
}

// LiftErrorFunc takes any 1-arity function "f(x) => error", and
// returns a new function that takes an additional error input. If that error
// is not nil, it is returned immediately before calling f. Otherwise, the
// result of calling f(x) normally is returned.
//
// This allows multiple simple error-returning functions to be called in
// sequence, with only one error check at the end.
func LiftErrorFunc[X any](f func(x X) error) func(err error, x X) error {
    return func(err error, x X) error {
        if err != nil { return err }
        return f(x)
    }
}

// FilterError returns err, unless errors.Is(err, i) returns true for any
// i in ignore, in which case it returns nil.
//
// For example,
//
//     // Create a symlink but ignore an error if the file exists.
//     err := FilterError(os.Symlink(oldname, newname), fs.ErrExist)
func FilterError(err error, ignore ... error) error {
    if err == nil { return nil }
    for _, i := range ignore {
        if errors.Is(err, i) { return nil }
    }
    return err
}

// MustMap is used to construct a map if it is nil, or return the input
// unchanged (i.e. the identity function) if it is not nil. This is useful
// for conditionally initialising a map that may or may not be its zero value.
func MustMap[K comparable, V any](m map[K]V) map[K]V {
    if m != nil { return m }
    return make(map[K]V)
}

// Reserve grows a slice, if necessary, to fit at least size extra elements.
//
// Deprecated: use [slices.Grow].
func Reserve[T any](xs []T, size int) []T {
    return slices.Grow(xs, size)
}

// SetLength grows a slice, if necessary, so that has a capacity of at least
// size elements, and a length of exactly size elements. Any trailing elements
// in the underlying array that fall beyond the original capacity are zeroed.
func SetLength[T any](xs []T, size int) []T {
    precap := cap(xs)
    grow := size - cap(xs)
    if grow <= 0 { return xs }
    xs = slices.Grow(xs, grow)
    xs = xs[0:size]
    clear(xs[precap:cap(xs)])
    return xs
}

// WrapBlock word-wraps a whitespace-delimited string to a given number of
// columns. The column length is given in runes (Unicode code points), not
// bytes.
//
// This is a simple implementation without any configuration options, designed
// for circumstances such as quickly wrapping a single error message for
// display.
//
// Save for bug fixes, the output of this function for any given input is
// frozen and will not be changed in future. This means you can reliably test
// against the return value of this function without your tests being brittle.
//
// Caveat: Single words longer than the column length will be truncated.
//
// Caveat: all whitespace, including existing new lines, is collapsed. An input
// consisting of multiple paragraphs will be wrapped into a single word-wrapped
// paragraph.
//
// Caveat: assumes all runes in the input string represent a glyph of length
// one. Whether this is true or not depends on how the display and font treats
// different runes. For example, some runes where [Unicode.IsGraphic] returns
// false might still be displayed as a special escaped character. Some letters
// might be displayed wider than usual, even in a monospaced font.
func WrapBlock(message string, columns int) string {
    var atoms = strings.Fields(strings.TrimSpace(message))
    var sb = strings.Builder{}
    var currentLength int

    if columns <= 0 { return "" }

    for i, atom := range atoms {
        isLast := (i + 1 == len(atoms))
        atomLength := utf8.RuneCountInString(atom)

        // special case for an atom longer than a whole line
        if (currentLength == 0) && (atomLength >= columns) {
            truncated := utf8string.NewString(atom).Slice(0, columns)
            sb.WriteString(truncated)
            if !isLast { sb.WriteByte('\n') }
            currentLength = 0
            continue
        }

        // will overflow?
        if currentLength + atomLength + 1 > columns {
            sb.WriteByte('\n')
            currentLength = 0
        }

        // mid-line?
        if currentLength > 0 {
            sb.WriteByte(' ')
            currentLength += 1
        }

        sb.WriteString(atom)
        currentLength += atomLength
    }

    return sb.String()
}
