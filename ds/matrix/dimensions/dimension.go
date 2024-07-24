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

func ascii(x int) int {
    original := x
    if x <= 64 { return x }
    if x >= 'a' { x -= 32 } // to uppercase
    switch x {
        case 'X': return 0
        case 'Y': return 1
        case 'Z': return 2
        case 'W': return 3

        default: return original
    }
}

// D is the interface implemented by an element that represents the
// dimensionality (e.g. 2D, 3D, etc.) and size (in each axis, e.g. x, y, z) of
// a matrix of values.
//
// It implements a bidirectional mapping from a single integer index to a
// slice of offsets along each axis. The mapping assumes row-major order.
//
// In general, D is treated as immutable and may be copied, but, for the sake
// of robustness, copies should be made using the Dimensions method.
type D interface {
    // Dimensions returns itself. If the implementation is heavy, it should
    // return a new D constructed with [New].
    Dimensions() D

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
    // example, Length(2) returns the depth in the third axis
    // (z axis).
    //
    // As a special case, the value of the ASCII characters in each of the
    // string "xyzw" may be used to refer to the dimensions 0, 1, 2 and 3,
    // respectively. For example, Length('z') == Length(2). Case is ignored.
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
var errLimitDims = errors.New("NewDimensions with more than 64 dimensions")

// New returns a new element implementing the dimensions interface D.
//
// In performance-sensitive code, this may be cast into the concrete types
// [D1], [D2], [D3], [D4], for 1-, 2-, 3-, 4-dimensional implementations,
// respectively.
//
// 5-dimensional and higher implementations are supported, but the concrete
// type is not exposed. Zero dimensions, dimensions of size zero, and
// dimensions higher than 64 are not supported and will panic.
func New(sizes ... int) D {
    if len(sizes) > 64 { panic(errLimitDims) }
    for i := 0; i < len(sizes); i++ {
        if sizes[i] == 0 { panic(errZeroSize) }
    }
    if len(sizes) <= 0 { panic(errZeroDims) }
    switch len(sizes) {
        case 1:  return D1([1]int{sizes[0]})
        case 2:  return D2([2]int{sizes[0], sizes[1]})
        case 3:  return D3([3]int{sizes[0], sizes[1], sizes[2]})
        case 4:  return D4([4]int{sizes[0], sizes[1], sizes[2], sizes[3]})
        default: return dN(append([]int{}, sizes...)) // don't share memory
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
    func (r D1) Dimensions() D                 { return r }
    func (r D1) Dimensionality() int           { return 1 }
    func (r D1) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }
    func (r D1) Length(idx int) int            { idx = ascii(idx); if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

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
    func (r D2) Dimensions() D                 { return r }
    func (r D2) Dimensionality() int           { return 2 }
    func (r D2) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }
    func (r D2) Length(idx int) int            { idx = ascii(idx); if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

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
    func (r D3) Dimensions() D                 { return r }
    func (r D3) Dimensionality() int           { return 3 }
    func (r D3) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }
    func (r D3) Length(idx int) int            { idx = ascii(idx); if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

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
    func (r D4) Dimensions() D                 { return r }
    func (r D4) Dimensionality() int           { return 4 }
    func (r D4) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }
    func (r D4) Length(idx int) int            { idx = ascii(idx); if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

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

// dN is an N-dimensional implementation of the [D] interface. In most cases,
// this is initialised by calling [New] with 5 or more arguments.
type dN []int

    func (r dN) Dimensions() D                 { return r } // immutable, so fine
    func (r dN) Dimensionality() int           { return len(r) }
    func (r dN) Contains(offsets ... int) bool { return dimensionsContains(r, offsets) }
    func (r dN) Length(idx int) int            { idx = ascii(idx); if idx < r.Dimensionality() { return r[idx] } else { return 0 } }

    func (r dN) Size() int {
        d := r.Dimensionality()
        total := 1
        for i := 0; i < d; i++ {
            total *= r[i]
        }
        return total
    }

    func (r dN) Lengths(dest []int) {
        copy(dest, r)
    }

    func (r dN) Index(offsets ... int) int {
        d := r.Dimensionality()
        stride := 1
        total := 0
        for i := 0; i < d; i++ {
            total += (offsets[i] % r[i]) * stride
            stride *= r[i]
        }
        return total
    }

    func (r dN) Offsets(dest []int, idx int) {
        d := r.Dimensionality()

        dest[0] = idx % r[0]
        stride := r[0]

        for i := 1; i < d; i++ {
            dest[i] = (idx / stride) % r[i]
            stride *= r[i]
        }
    }
