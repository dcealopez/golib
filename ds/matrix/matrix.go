// Package matrix implements several different data structures for efficiently
// representing large matrices of arbitrary size and dimensionality.
//
// Note that fixed-size small matrices would be better represented using a
// special-purpose package e.g. a GLM-style implementation.
//
// The concrete implementations are exposed for use by performance-sensitive
// code.
package matrix

import (
    "github.com/tawesoft/golib/v2/ds/matrix/dimensions"
    "github.com/tawesoft/golib/v2/math/series"
)

// M is the interface implemented by a matrix of Values.
type M[T comparable] interface {
    // D provides an efficient way to index the matrix and query its shape.
    //
    // Performance-sensitive code may like to cast D to a concrete type e.g.
    // [dimensions.D2] for a 2-Dimensional matrix.
    dimensions.D

    // Get returns the value stored at the given index. Use the Index method
    // on D to obtain an index from coordinates.
    Get(idx int) T

    // Set sets the value stored at the given index. Use the Index method
    // on D to obtain an index from coordinates.
    Set(idx int, value T)

    // Next returns the subsequent index for a non-zero value in the matrix
    // after idx. Use a negative index to start searching from the first
    // element. Use the Offsets method on [dimensions.D] to obtain coordinates
    // from an index. The second return value is false iff all remaining Values
    // are the zero value for type T.
    //
    // This function therefore efficiently enumerates all non-zero values in a
    // sparse matrix.
    Next(idx int) (int, bool)

    // Clear sets every element in the matrix to the zero value for type T.
    Clear()
}

// Copy clears dest and copies every value from src into dest at the same
// offsets. If dest is smaller than src along any dimension, the results are
// cropped.
func Copy[T comparable](dest, src M[T]) {
    dest.Clear()
    offsets := make([]int, src.Dimensionality())
    for i := 0; i < src.Size(); i++ {
        src.Offsets(offsets, i)
        if dest.Contains(offsets...) {
            dest.Set(dest.Index(offsets...), src.Get(i))
        }
    }
}

type constMatrix[T comparable] struct { dimensions.D; m M[T] }

    // Const returns a read-only view of matrix m. Changes to m will affect the
    // returned matrix. The returned matrix will panic if its Set or Clear
    // methods are called.
    func Const[T comparable](m M[T]) M[T] {
        return constMatrix[T]{
            D: m,
            m: m,
        }
    }

    func (c constMatrix[T]) Get(idx int) T {
        return c.m.Get(idx)
    }

    func (c constMatrix[T]) Set(idx int, value T) {
        panic("const matrix is read only")
    }

    func (c constMatrix[T]) Next(idx int) (int, bool) {
        return c.m.Next(idx)
    }

    func (c constMatrix[T]) Clear() {
        panic("const matrix is read only")
    }


// Grid is an implementation of the matrix interface [M] that stores data using
// a contiguous slice of values. In most cases, this is initialised by calling
// [New] or [NewGrid]. Performance sensitive code may cast M to this type.
type Grid[T comparable] struct {
    dimensions.D
    values []T
}

    // NewGrid allocates and returns a new grid matrix implementing M.
    func NewGrid[T comparable](lengths ... int) M[T] {
        dims := dimensions.New(lengths...)
        return Grid[T]{
            D: dims,
            values: make([]T, dims.Size()),
        }
    }

    // NewSharedGrid returns a new [Grid] matrix implementing M. The grid
    // uses the provided slice of values as its storage. Values are laid out in
    // row-major order. This memory is shared: modifications to the values
    // slice will modify the matrix, and modifications to the matrix will
    // modify the values slice. The length of the values slice must be greater
    // than or equal to the product of all lengths.
    func NewSharedGrid[T comparable](lengths []int, values []T) M[T] {
        dims := dimensions.New(lengths...)
        if len(values) < dims.Size() { panic("shared grid buffer too small") }
        return Grid[T]{
            D: dims,
            values: values,
        }
    }

    func (g Grid[T]) Get(idx int) T {
        return g.values[idx]
    }

    func (g Grid[T]) Set(idx int, value T) {
        g.values[idx] = value
    }

    func (g Grid[T]) Next(idx int) (int, bool) {
        var zero T
        if idx < 0 { idx = -1 }
        idx++
        for i := idx; i < g.Size(); i++ {
            if g.values[i] != zero {
                return i, true
            }
        }
        return 0, false
    }

    func (g Grid[T]) Clear() {
        clear(g.values)
    }

// Bool is an implementation of the matrix interface [M] that stores data using
// a densely packed sequence of bits that are either true or false. In most
// cases, this is initialised by calling [New] or [NewBool]. Performance
// sensitive code may cast M to this type.
type Bool struct {
    dimensions.D
    buckets []uint64
}

    // NewBool allocates and returns a new [Bool] matrix implementing M.
    func NewBool(lengths ... int) M[bool] {
        dims := dimensions.New(lengths...)
        numBuckets := dims.Size() / 64
        if dims.Size() % 64 != 0 {
            numBuckets++
        }
        return Bool{
            D: dims,
            buckets: make([]uint64, numBuckets),
        }
    }

    func bucketIndex(idx int) (int, uint64) {
        // bucket at idx / 64, and bit set at sub-bucket offset
        return idx >> 6, 1 << (idx % 64)
    }

    func (b Bool) Get(idx int) bool {
        bucket, mask := bucketIndex(idx)
        return (mask & b.buckets[bucket]) != 0
    }

    func (b Bool) Set(idx int, value bool) {
        bucket, mask := bucketIndex(idx)
        if value {
            b.buckets[bucket] |= mask
        } else {
            b.buckets[bucket] &= ^mask
        }
    }

    func (b Bool) Next(idx int) (int, bool) {
        if idx < 0 { idx = -1 }
        idx++
        start, offset := idx / 64, uint64(idx % 64)
        target := -1

        // Find the next non-zero bucket
        for i := start; i < len(b.buckets); i++ {
            if b.buckets[i] != 0 {
                target = i
                break
            }
        }
        if target < 0 { return 0, false }

        // Find the next non-zero bit
        for i := offset; i < 64; i++ {
            mask := uint64(1) << i
            if b.buckets[i] & mask == mask {
                return int(i), true
            }
        }

        return 0, false
    }

    func (b Bool) Clear() {
        clear(b.buckets)
    }

// Bit is an implementation of the matrix interface [M] that stores data using
// a densely packed sequence of bits that are either 1 or 0. In most cases,
// this is initialised by calling [New] or [NewBit]. Performance sensitive
// code may cast M to this type.
type Bit struct {
    dimensions.D
    buckets []uint64
}

    // NewBit allocates and returns a [Bit] matrix implementing M.
    func NewBit(lengths ... int) M[int] {
        dims := dimensions.New(lengths...)
        numBuckets := dims.Size() / 64
        if dims.Size() % 64 != 0 {
            numBuckets++
        }
        return Bit{
            D: dims,
            buckets: make([]uint64, numBuckets),
        }
    }

    func (b Bit) Get(idx int) int {
        bucket, mask := bucketIndex(idx)
        return int((mask & b.buckets[bucket]) >> (idx % 64))
    }

    func (b Bit) Set(idx int, value int) {
        bucket, mask := bucketIndex(idx)
        if value > 0 {
            b.buckets[bucket] |= mask
        } else {
            b.buckets[bucket] &= ^mask
        }
    }

    func (b Bit) Next(idx int) (int, bool) {
        if idx < 0 { idx = -1 }
        idx++
        start, offset := idx / 64, uint64(idx % 64)
        target := -1

        // Find the next non-zero bucket
        for i := start; i < len(b.buckets); i++ {
            if b.buckets[i] != 0 {
                target = i
                break
            }
        }
        if target < 0 { return 0, false }

        // Find the next non-zero bit
        for i := offset; i < 64; i++ {
            mask := uint64(1) << i
            if b.buckets[target] & mask == mask {
                return int(i), true
            }
        }

        return 0, false
    }

    func (b Bit) Clear() {
        clear(b.buckets)
    }

// Hashmap is an implementation of the matrix interface [M] that stores data
// using a hashmap with element indexes as keys. Elements with the zero value
// are omitted. This implementation is best suited to representing very sparse
// matrices. In most cases, this is initialised by calling [New] or
// [NewHashmap].
type Hashmap[T comparable] struct {
    dimensions.D
    values map[int]T
}

    // NewHashmap allocates and returns a [Hashmap] implementating M.
    func NewHashmap[T comparable](lengths ... int) M[T] {
        return Hashmap[T]{
            D: dimensions.New(lengths...),
            values: make(map[int]T),
        }
    }

    // NewSharedHashmap returns a new [Hashmap] matrix implementing [M].
    // The matrix uses the provided map of values as its storage. This memory
    // is shared: modifications to the values map will modify the matrix, and
    // modifications to the matrix will modify the values map.
    func NewSharedHashmap[T comparable](lengths []int, values map[int]T) M[T] {
        return Hashmap[T]{
            D: dimensions.New(lengths...),
            values: values,
        }
    }

    func (m Hashmap[T]) Get(idx int) T {
        if values, ok := m.values[idx]; ok {
            return values
        } else {
            var zero T
            return zero
        }
    }

    func (m Hashmap[T]) Set(idx int, value T) {
        var zero T
        if zero == value {
            delete(m.values, idx)
        } else {
            m.values[idx] = value
        }
    }

    func (m Hashmap[T]) Next(idx int) (int, bool) {
        var zero T
        if idx < 0 { idx = -1 }
        idx++
        for i := idx; i < m.Size(); i++ {
            if value, ok := m.values[i]; ok && (value != zero) {
                return i, true
            }
        }
        return 0, false
    }

    func (m Hashmap[T]) Clear() {
        clear(m.values)
    }

// Diagonal is an implementation of the matrix interface [M] that represents a
// diagonal matrix (one where entries outside the main diagonal are all zero)
// as a contiguous slice of only the diagonal values. In most cases, this is
// initialised by calling [New] or [NewDiagonal].
//
// Setting an element of the diagonal matrix to a non-zero value if it does not
// lie on the diagonal is an error, and will panic.
type Diagonal[T comparable] struct {
    dimensions.D
    values []T
}

    // NewDiagonal allocates and returns a [Diagonal] implementing M.
    //
    // Note the unique constructor: a Diagonal matrix is the same size along
    // each axis, and therefore the caller need only specify the dimensionality
    // and the length of one side.
    func NewDiagonal[T comparable](dimensionality, length int) M[T] {
        lengths := make([]int, dimensionality)
        for i := 0; i < dimensionality; i++ {
            lengths[i] = length
        }
        return Diagonal[T]{
            D: dimensions.New(lengths...),
            values: make([]T, length),
        }
    }

    // NewSharedDiagonal returns a new [Diagonal] matrix implementing M. The
    // matrix uses the provided slice of values, which are the values along the
    // diagonal only, as its storage. This memory is shared: modifications to
    // the values slice will modify the matrix, and modifications to the matrix
    // will modify the values slice. The length of the values slice sets the
    // length of each side of the matrix.
    func NewSharedDiagonal[T comparable](dimensionality int, values []T) M[T] {
        length := len(values)
        lengths := make([]int, dimensionality)
        for i := 0; i < dimensionality; i++ {
            lengths[i] = length
        }
        return Diagonal[T]{
            D: dimensions.New(lengths...),
            values: values,
        }
    }

    // diagonalOffset converts an index into the matrix into an index into
    // the flat array of values on the diagonal, or -1 if not on the diagonal.
    func diagonalOffset(dimensionality int, length int, idx int) int {
        // the diagonal covers indexes 0, n, 2n, 3n, 4n ...
        // where d := the dimensionality, x is the length of one side,
        // n = x^0 + x^1 + x^2 + ... x^(d-1)
        n := series.NewGeometricInteger[int](1, length).Sum(dimensionality - 1)
        if idx % n != 0 {
            return -1
        }
        return idx / n
    }

    func nextDiagonalIndex(dimensionality int, length int, idx int) int {
        // the diagonal covers indexes 0, n, 2n, 3n, 4n ...
        // where d := the dimensionality, x is the length,
        // n = x^0 + x^1 + x^2 + ... x^(d-1)
        n := series.NewGeometricInteger[int](1, length).Sum(dimensionality - 1)
        if idx % n != 0 {
            idx -= (idx % n)
        }
        return idx + n
    }

    func (m Diagonal[T]) Get(idx int) T {
        var zero T
        i := diagonalOffset(m.Dimensionality(), len(m.values), idx)
        if i < 0 { return zero }
        if i > len(m.values) { return zero }
        return m.values[i]
    }

    func (m Diagonal[T]) Set(idx int, value T) {
        var zero T
        i := diagonalOffset(m.Dimensionality(), len(m.values), idx)
        if i < 0 {
            if value == zero { return }
            panic("can not set a non-zero value off the matrix diagonal")
        }
        m.values[i] = value
    }

    func (m Diagonal[T]) Next(idx int) (int, bool) {
        var zero T
        if idx < 0 { idx = -1 }
        for {
            idx = nextDiagonalIndex(m.Dimensionality(), len(m.values), idx)
            if idx > m.Size() { break }
            i := diagonalOffset(m.Dimensionality(), len(m.values), idx)
            if m.values[i] != zero {
                return idx, true
            }
        }
        return 0, false
    }

    func (m Diagonal[T]) Clear() {
        clear(m.values)
    }

// View is an implementation of the matrix interface [M] that represents a
// custom view of a parent matrix. The View presents its own dimensions and
// element offsets and indexes, which are mapped appropriately to the parent
// matrix.
//
// The underlying memory is shared, so that a modification to either the view
// or the parent modifies the other.
//
// Clearing a view clears only the elements in the view that map to an element
// in the parent.
type View[T comparable] struct {
    dimensions.D
    parent  M[T]
    mapping dimensions.Map
}

    // NewView returns a [View], implementing the matrix interface M, using the
    // provided mapping.
    //
    // See [dimensions.Crop] and the Bind method on [dimensions.Sampler] for
    // constructing maps from a parent.
    func NewView[T comparable](parent M[T], mapping dimensions.Map) M[T] {
        return View[T]{
            D: mapping.Dimensions(),
            parent:  parent,
            mapping: mapping,
        }
    }

    func (v View[T]) Get(idx int) T {
        return v.parent.Get(v.mapping.MapIndex(idx))
    }

    func (v View[T]) Set(idx int, value T) {
        v.parent.Set(v.mapping.MapIndex(idx), value)
    }

    func (v View[T]) Next(idx int) (int, bool) {
        var zero T
        if idx < 0 { idx = -1 }
        idx++
        for i := idx; i < v.Size(); i++ {
            if v.Get(i) != zero {
                return i, true
            }
        }
        return 0, false
    }

    func (v View[T]) Clear() {
        var zero T
        for idx, ok := -1, true; ok; idx, ok = v.Next(idx) {
            if idx < 0 { continue }
            v.Set(idx, zero)
        }
    }

    // Crop is a shortcut for NewView(parent, dimensions.Crop(...))
    //
    // See [dimensions.Crop].
    func Crop[T comparable](parent M[T], startIdx int, lengths ... int) M[T] {
        return NewView[T](parent, dimensions.Crop(parent, startIdx, lengths...))
    }

    // Sample is a shortcut for NewView(parent, dimensions.Sampler(...).Bind(parent)).
    //
    // See [dimensions.Sampler].
    func Sample[T comparable](parent M[T], sampler string, constants ... int) M[T] {
        return NewView[T](parent, dimensions.Sampler(sampler, constants...).Bind(parent))
    }
