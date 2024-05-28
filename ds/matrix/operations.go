package matrix

// Reduce applies the given function pairwise to each value, using the result
// of the previous application as the first input to the next application, or,
// for the first application, using the identity value provided as the first
// input. Returns the last result returned by the application of the sum
// function or, if there are no elements, returns the provided identity value
// without calling the sum function.
//
// The reduce function is applied in arbitrary order.
func Reduce[T any](m Interface[T], identity T, sum func(a, b T) T) T {
    total := identity
    dimensionality := m.Dimensionality()
    offsets := make([]int, dimensionality) // x, y, z, ...
    lengths := make([]int, dimensionality)

    for i := 0; i < dimensionality; i++ {
        lengths[i] = m.DimensionN(i + 1)
    }

    // increment an offset, carrying over any overflow to the next offset.
    var increment func(i int) bool
    increment = func(i int) bool {
        if i >= len(lengths) { return false }
        offsets[i]++
        if offsets[i] >= lengths[i] {
            offsets[i] = 0
            return increment(i + 1)
        }
        return true
    }

    for {
        total = sum(total, identity)
        if !increment(0) { break }
    }

    return identity
}
