// Package dimensions implements an interface and concrete types for indexing
// a matrix of values along an arbitrary number of dimensions.
//
// The concrete implementations are exposed for use by performance-sensitive
// code.
package dimensions

import (
    "errors"
)

// TODO: various "swizzle"-style mappings

// D is the interface implemented by an element that represents the
// dimensionality (e.g. 2D, 3D, etc.) and size (in each axis, e.g. x, y, z) of
// a matrix of values.
//
// It implements a bidirectional mapping from a single integer index to a
// slice of offsets along each axis. The mapping assumes row-major order.
type D interface {
    // Index computes an index calculated from the offsets along each axis.
    //
    // If any individual offset is out of bounds, it is wrapped round, modulo
    // the length along that axis. The caller must ensure that dest has a
    // length >= Dimensionality.
    Index(offsets ...int) int

    // Offsets calculates the offset along each axis identified by the given
    // index. The results are stored in dest. If the index is out of bounds,
    // it is wrapped round modulo Size. The caller must ensure that dest has
    // a length >= Dimensionality.
    Offsets(dest []int, idx int)

    // Contains returns true if each provided offset is less than the length
    // along its respective axis.
    //
    // As a special case, if len(offsets) is greater than the Dimensionality of
    // the matrix, an element is still considered to be contained if the offset
    // is zero. For example, (2, 4, 0) is considered to be contained by a
    // 2-dimensional matrix iff (2, 4) is also contained.
    Contains(offsets ... int) bool

    // Size returns the number of unique indexes, from 0 to size minus 1
    // inclusive, that can be used to obtain an Index. This is also the unit
    // volume of the matrix.
    Size() int

    // Dimensionality returns the number of dimensions e.g. 2 for "2D".
    Dimensionality() int

    // Length returns the length along the specified zero-indexed axis. For
    // example, Length(2) returns the depth in the third axis (z axis).
    //
    // The result is undefined if idx is less than zero. If idx is >=
    // Dimensionality, returns zero.
    Length(idx int) int

    // Lengths returns the lengths along each axis. The results are stored in
    // dest. If dest is not large enough, the results are truncated.
    Lengths(dest []int)
}

var errZeroSize = errors.New("NewDimensions with zero-length size")
var errZeroDims = errors.New("NewDimensions with empty sizes slice")

// New returns a new element implementing the dimensions interface D.
//
// In performance-sensitive code, this may be cast into the concrete types
// [D1], [D2], [D3], [D4], or [DN], for 1-, 2-, 3-, 4-, or >=5-dimensional
// implementations, respectively.
func New(sizes ... int) D {
    for i := 0; i < len(sizes); i++ {
        if sizes[i] == 0 { panic(errZeroSize) }
    }
    if len(sizes) <= 0 { panic(errZeroDims) }
    switch len(sizes) {
        case 1:  return D1([1]int{sizes[0]})
        case 2:  return D2([2]int{sizes[0], sizes[1]})
        case 3:  return D3([3]int{sizes[0], sizes[1], sizes[2]})
        case 4:  return D4([4]int{sizes[0], sizes[1], sizes[2], sizes[3]})
        default: return DN(append([]int{}, sizes...)) // don't share memory
    }
}

func dimensionsContains(dims []int, offsets []int) bool {
    nDims := len(dims)
    for i := 0; i < len(offsets); i++ {
        // regular case
        if (i < nDims) && (offsets[i] >= dims[i]) { return false }

        // special case - allow trailing offsets that are zero
        if (i >= nDims) && (offsets[i] > 0) { return false }

        if offsets[i] < 0 { return false }
    }
    return true
}

// D1 is a 1-dimensional implementation of the [D] interface. In most cases,
// this is initialised by calling [New] with 1 argument. Performance sensitive
// code may cast D to this type.
type D1 [1]int

    func (r D1) Size() int                     { return r[0] }
    func (r D1) Dimensionality() int           { return 1 }
    func (r D1) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }
    func (r D1) Length(idx int) int            { if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

    func (r D1) Lengths(dest []int) {
        if len(dest) == 0 { return }
        dest[0] = r[0]
    }

    func (r D1) Index(offsets ... int) int {
        x := offsets[0] % r[0]
        return x
    }

    func (r D1) Offsets(dest []int, idx int) {
        w := r[0]
        dest[0] = idx % w
    }

// D2 is a 2-dimensional implementation of the [D] interface. In most cases,
// this is initialised by calling [New] with 2 arguments. Performance sensitive
// code may cast D to this type.
type D2 [2]int

    func (r D2) Size() int                     { return r[0] * r[1] }
    func (r D2) Dimensionality() int           { return 2 }
    func (r D2) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }
    func (r D2) Length(idx int) int            { if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

    func (r D2) Lengths(dest []int) {
        if len(dest) < r.Dimensionality() { copy(dest, r[:]) }
        dest[0] = r[0]
        dest[1] = r[1]
    }

    func (r D2) Index(offsets ... int) int {
        x := offsets[0] % r[0]
        y := offsets[1] % r[1]
        w := r[0]
        return (y * w) + x
    }

    func (r D2) Offsets(dest []int, idx int) {
        w := r[0]
        h := r[1]
        x := idx % w
        y := (idx / w) % h
        dest[0] = x
        dest[1] = y
    }

// D3 is a 3-dimensional implementation of the [D] interface. In most cases,
// this is initialised by calling [New] with 3 arguments. Performance sensitive
// code may cast D to this type.
type D3 [3]int

    func (r D3) Size() int                     { return r[0] * r[1] * r[2] }
    func (r D3) Dimensionality() int           { return 3 }
    func (r D3) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }
    func (r D3) Length(idx int) int            { if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

    func (r D3) Lengths(dest []int) {
        if len(dest) < r.Dimensionality() { copy(dest, r[:]) }
        dest[0] = r[0]
        dest[1] = r[1]
        dest[2] = r[2]
    }

    func (r D3) Index(offsets ... int) int {
        x := offsets[0] % r[0]
        y := offsets[1] % r[1]
        z := offsets[2] % r[2]
        w := r[0]
        h := r[1]
        return (z * w * h) + (y * w) + x
    }

    func (r D3) Offsets(dest []int, idx int) {
        w := r[0]
        h := r[1]
        d := r[2]
        x := idx % w
        y := (idx / w) % h
        z := (idx / (w * h)) % d
        dest[0] = x
        dest[1] = y
        dest[2] = z
    }

// D4 is a 4-dimensional implementation of the [D] interface. In most cases,
// this is initialised by calling [New] with 4 arguments. Performance sensitive
// code may cast D to this type.
type D4 [4]int // width, height, depth, extent

    func (r D4) Size() int                     { return r[0] * r[1] * r[2] * r[3] }
    func (r D4) Dimensionality() int           { return 4 }
    func (r D4) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }
    func (r D4) Length(idx int) int            { if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

    func (r D4) Lengths(dest []int) {
        if len(dest) < r.Dimensionality() { copy(dest, r[:]) }
        dest[0] = r[0]
        dest[1] = r[1]
        dest[2] = r[2]
        dest[3] = r[3]
    }

    func (r D4) Index(offsets ... int) int {
        x := offsets[0] % r[0]
        y := offsets[1] % r[1]
        z := offsets[2] % r[2]
        e := offsets[3] % r[3]
        w := r[0]
        h := r[1]
        d := r[2]
        return (e * w * h * d) + (z * w * h) + (y * w) + x
    }

    func (r D4) Offsets(dest []int, idx int) {
        w := r[0]
        h := r[1]
        d := r[2]
        e := r[3]
        x := idx % w
        y := (idx / w) % h
        z := (idx / (w * h)) % d
        u := (idx / (w * h * d)) % e
        dest[0] = x
        dest[1] = y
        dest[2] = z
        dest[3] = u
    }

// DN is an N-dimensional implementation of the [D] interface. In most cases,
// this is initialised by calling [New] with 5 or more arguments. Performance
// sensitive code may cast D to this type.
type DN []int

    func (r DN) Dimensionality() int           { return len(r) }
    func (r DN) Contains(offsets ... int) bool { return dimensionsContains(r, offsets) }
    func (r DN) Length(idx int) int            { if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

    func (r DN) Size() int {
        d := r.Dimensionality()
        total := 1
        for i := 0; i < d; i++ {
            total *= r[i]
        }
        return total
    }

    func (r DN) Lengths(dest []int) {
        copy(dest, r)
    }

    func (r DN) Index(offsets ... int) int {
        d := r.Dimensionality()
        stride := 1
        total := 0
        for i := 0; i < d; i++ {
            total += (offsets[i] % r[i]) * stride
            stride *= r[i]
        }
        return total
    }

    func (r DN) Offsets(dest []int, idx int) {
        d := r.Dimensionality()

        dest[0] = idx % r[0]
        stride := r[0]

        for i := 1; i < d; i++ {
            dest[i] = (idx / stride) % r[i]
            stride *= r[i]
        }
    }
