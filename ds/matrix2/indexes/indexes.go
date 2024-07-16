// Package indexes implements useful features for indexing multidimensional
// arrays or matrices.
//
// Note that in this package an array or matrix with a size of zero in any
// dimension is not well-defined. Also, constructing an iterator with more
// permutations of indexes than that representable in a word-sized integer will
// panic. This means that the product of
// (An iterator itself, once constructed, will not panic).
package indexes

import (
    "errors"
    "slices"

    "github.com/tawesoft/golib/v2/must"
)

type Dimensions interface {
    ToIndex(offsets ...int) int
    FromIndex(dest []int, idx int)
    Size() int
}

type Dimensions2D struct { width, height int }

func (r Dimensions2D) ToIndex(offsets ... int) int {
    x, y := offsets[0], offsets[1]
    return (y * r.width) + x
}

func (r Dimensions2D) FromIndex(offsets ... int) int {
    x, y := offsets[0], offsets[1]
    return (y * r.width) + x
}



var ErrLength = errors.New("slices must have equal length")

// Iterator is the type of a iterator-style function that produces a
// permutation of each index from 0 (inclusive) up to the length (exclusive) in
// each dimension, storing it in the result slice. The slice returned by the
// iterator is only valid until the next call to the iterator. The iterator
// returns false if completed, in which case its first return value should be
// ignored.
//
// The order that indexes are generated depends on the implementation e.g.
// [Forwards], and the required length of the destination array must be the
// same as the length of the initial input to that implementation.
type Iterator func(dest []int) bool

// MapFunc is the type of a function that modifies a slice of indexes in-place
// to some other values.
type MapFunc func([]int)

// FilterFunc is the type of a function that examines a slice of indexes
// without modifying it and returns true or false depending on some condition.
type FilterFunc func([]int) bool

// Map applies f to transform an iterator function and returns a new iterator
// function. The transformation is applied in-place to its input, and the
// input iterator is pulled from every time the output iterator is pulled from.
func Map(f MapFunc, iterator Iterator) Iterator {
    return func(dest []int) bool {
        for iterator(dest) {
            f(dest)
            return true
        }
        return false
    }
}

// Filter applies f to an iterator function and returns a new iterator function
// that only produces a result if f would return true on its output. The input
// iterator is pulled from every time the output iterator is pulled from.
func Filter(f FilterFunc, iterator Iterator) Iterator {
    return func(dest []int) bool {
        for iterator(dest) {
            if !f(dest) { continue }
            return true
        }
        return false
    }
}

// Forwards implements an iterator of indexes produced in ascending order, with
// earlier dimensions treated as less significant. For example, Forwards(3, 2),
// produces the coordinate indices [0, 0], [1, 0], [2, 0], [0, 1], [1, 1], [2,
// 1], in that order.
func Forwards(dimensions ... int) Iterator {
    dimensionality := len(dimensions)
    offsets := make([]int, dimensionality)

    var increment func(i int) bool
    increment = func(i int) bool {
        if i >= dimensionality {
            return false
        }
        offsets[i]++
        if offsets[i] >= dimensions[i] {
            offsets[i] = 0
            return increment(i + 1)
        }
        return true
    }

    first := false
    ok := true
    return func(dest []int) bool {
        if !ok { return false }
        copy(dest, offsets) // return a copy so that it can be modified in-place
        if first { return true }
        ok = increment(0) // "increment x" (dimension at offsets[0])
        return true
    }
}

// Backwards generates the same values as [Iterator], but in reverse order.
func Backwards(dimensions ... int) Iterator {
    return Map(
        func(xs []int) { slices.Reverse(xs) },
        Forwards(dimensions...),
    )
}

// Offset creates a mapping function which adds a constant offset to each
// index.
func Offset(offsets ... int) MapFunc {
    return func(xs []int) {
        for i := 0; i < len(xs); i++ {
            xs[i] = xs[i] + offsets[i]
        }
    }
}

// Rotate creates a mapping function that can swap elements in a slice by
// indexing them and specifying a new order.
//
// For example, Rotate(0, 1) does nothing, while Rotate(1, 0) swaps the
// first and second indexes.
//
// Note that this must not change the dimensionality of an iterator.
func Rotate(offsets ... int) MapFunc {
    results := make([]int, len(offsets))

    return func(dest []int) {
        if len(offsets) != len(dest) {
            panic(ErrLength)
        }
        for i := 0; i < len(dest); i++ {
            results[i] = dest[offsets[i]]
        }
        copy(dest, results)
    }
}

const errMsgRangeSize = "indexes iterator: range.First and range.Last must be the same length"

// Range defines a range of indexes from a first index (inclusive) to a second
// index (exclusive).
type Range interface {
    Forwards() Iterator  // see [Forwards].
    Backwards() Iterator // see [Backwards]

    // Contains returns true if the given range contains the given vertex
    // index.
    //
    // As a special case, the dimensionality of indexes may be greater than the
    // dimensionality of the range. If so, the given vertex is said to be
    // contained by r only if every component at a higher dimension is zero.
    // For example, a range from (1, 1) to (4, 4) contains (2, 2, 0) but not
    // (2, 2, 1).
    Contains(indexes ... int) bool
}

// NewRange returns a new Range.
//
// The slices First and Last must have the same length.
//
// For example, NewRange([]int{1, 2}, []int{4, 4}).Iterator() would generate the
// indexes [1, 2], [2, 2], [3, 2], [1, 3], [2, 3], [3, 3], in that order.
func NewRange(first, last []int) Range {
    must.Equalf(len(first), len(last), errMsgRangeSize)

    // TODO depending on the size of the input slice, pick a
    //  statically-sized implementation.

    return rangeN{
        first: append([]int{}, first...),
        last:  append([]int{}, last...),
    }
}

// rangeN is a range of arbitrary dimension.
type rangeN struct {
    first, last []int
}

func (r rangeN) Forwards() Iterator {
    offsets := make([]int, len(r.last))
    lengths := make([]int, len(r.last))
    for i := 0; i < len(r.last); i++ {
        offsets[i] = r.first[i]
        lengths[i] = r.last[i] - r.first[i]
    }

    return Map(Offset(offsets...), Forwards(lengths...))
}

func (r rangeN) Backwards() Iterator {
    offsets := make([]int, len(r.last))
    lengths := make([]int, len(r.last))
    for i := 0; i < len(r.last); i++ {
        offsets[i] = r.first[i]
        lengths[i] = r.last[i] - r.first[i]
    }

    return Map(Offset(offsets...), Backwards(lengths...))
}

func (r rangeN) Contains(indexes ... int) bool {
    dims := len(r.first)
    for i := 0; i < len(indexes); i++ {
        if i >= dims {
            if indexes[i] != 0 { return false }
            continue
        }
        if indexes[i] < r.first[i] { return false }
        if indexes[i] >= r.last[i] { return false }
    }
    return true
}
