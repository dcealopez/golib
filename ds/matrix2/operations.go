package matrix2

import (
    "github.com/tawesoft/golib/v2/ds/matrix/indexes"
)

// Reduce applies the given function pairwise to each value, using the result
// of the previous application as the first input to the next application, or,
// for the first application, using the identity value provided as the first
// input. Returns the last result returned by the application of the sum
// function or, if there are no elements, returns the provided identity value
// without calling the sum function.
//
// The reduce function is applied in arbitrary order.
func Reduce[T comparable](m Interface[T], identity T, sum func(a, b T) T) T {
    // We can only, in the general case, give this an optimised version for
    // sparse matrices if the zero value is the identity value.
    // TODO optimised case

    total := identity
    idx := make([]int, m.Dimensionality())
    iter := indexes.Forwards(m.Dimensions()...)

    for iter(idx) {
        value := m.GetN(idx...)
        total = sum(total, value)
    }

    return total
}
