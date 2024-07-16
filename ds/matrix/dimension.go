package matrix

import (
    "errors"
)

// Dimensions represents the dimensionality (e.g. 2D, 3D, etc.) and size (in
// each dimension, e.g. x, y, z) of a matrix of values.
//
// It implements a bidirectional mapping from a single integer index to a
// slice of offsets in each dimension. The mapping assumes row-major order.
type Dimensions interface {
    // Index computes an index calculated from the offsets into each dimension.
    //
    // If any individual offset is out of bounds, it is wrapped round, modulo
    // the length of that dimension. The caller must ensure that dest has
    // a length >= Dimensionality.
    Index(offsets ...int) int

    // Offsets calculates the offset in each dimension identified by the given
    // index. The results are stored in dest. If the index is out of bounds,
    // it is wrapped round modulo Size. The caller must ensure that dest has
    // a length >= Dimensionality.
    Offsets(dest []int, idx int)

    // Contains returns true if each provided offset is less than the length of
    // its respective dimension.
    //
    // As a special case, if len(offsets) is greater than the Dimensionality of
    // the matrix, an element is still considered to be contained if the offset
    // is zero. For example, (2, 4, 0) is considered to be contained by a
    // 2-dimensional matrix iff (2, 4) is also contained.
    Contains(offsets ... int) bool

    // Size returns the number of unique indexes, from 0 to size minus 1
    // inclusive, that can be used to obtain an Index.
    Size() int

    // Dimensionality returns the number of dimensions e.g. 2 for "2D".
    Dimensionality() int

    // Lengths returns the length of each dimension. The results are stored in
    // dest. If dest is not large enough, the results are truncated.
    Lengths(dest []int)
}

var errZeroSize = errors.New("NewDimensions with zero-length size")
var errZeroDims = errors.New("NewDimensions with empty sizes slice")

func NewDimensions(sizes ... int) Dimensions {
    for i := 0; i < len(sizes); i++ {
        if sizes[i] == 0 { panic(errZeroSize) }
    }
    switch len(sizes) {
        case 0: panic(errZeroDims)
        case 1: return d1D([1]int{sizes[0]})
        case 2: return d2D([2]int{sizes[0], sizes[1]})
        case 3: return d3D([3]int{sizes[0], sizes[1], sizes[2]})
        case 4: return d4D([4]int{sizes[0], sizes[1], sizes[2], sizes[3]})
        default: return dND(append([]int{}, sizes...)) // don't share memory
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

// d1D is a 1-dimensional implementation of the Dimensions interface.
type d1D [1]int

    func (r d1D) Size() int { return r[0] }
    func (r d1D) Dimensionality() int { return 1 }
    func (r d1D) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }

    func (r d1D) Lengths(dest []int) {
        if len(dest) == 0 { return }
        dest[0] = r[0]
    }

    func (r d1D) Index(offsets ... int) int {
        x := offsets[0] % r[0]
        return x
    }

    func (r d1D) Offsets(dest []int, idx int) {
        w := r[0]
        dest[0] = idx % w
    }

// d2D is a 2-dimensional implementation of the Dimensions interface.
type d2D [2]int

    func (r d2D) Size() int { return r[0] * r[1] }
    func (r d2D) Dimensionality() int { return 2 }
    func (r d2D) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }

    func (r d2D) Lengths(dest []int) {
        if len(dest) < r.Dimensionality() { copy(dest, r[:]) }
        dest[0] = r[0]
        dest[1] = r[1]
    }

    func (r d2D) Index(offsets ... int) int {
        x := offsets[0] % r[0]
        y := offsets[1] % r[1]
        w := r[0]
        return (y * w) + x
    }

    func (r d2D) Offsets(dest []int, idx int) {
        w := r[0]
        h := r[1]
        x := idx % w
        y := (idx / w) % h
        dest[0] = x
        dest[1] = y
    }

// d3D is a 3-dimensional implementation of the Dimensions interface.
type d3D [3]int

    func (r d3D) Size() int { return r[0] * r[1] * r[2] }
    func (r d3D) Dimensionality() int { return 3 }
    func (r d3D) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }

    func (r d3D) Lengths(dest []int) {
        if len(dest) < r.Dimensionality() { copy(dest, r[:]) }
        dest[0] = r[0]
        dest[1] = r[1]
        dest[2] = r[2]
    }

    func (r d3D) Index(offsets ... int) int {
        x := offsets[0] % r[0]
        y := offsets[1] % r[1]
        z := offsets[2] % r[2]
        w := r[0]
        h := r[1]
        return (z * w * h) + (y * w) + x
    }

    func (r d3D) Offsets(dest []int, idx int) {
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

// d4D is a 4-dimensional implementation of the Dimensions interface.
type d4D [4]int // width, height, depth, extent

    func (r d4D) Size() int { return r[0] * r[1] * r[2] * r[3] }
    func (r d4D) Dimensionality() int { return 4 }
    func (r d4D) Contains(offsets ... int) bool { return dimensionsContains(r[:], offsets) }

    func (r d4D) Lengths(dest []int) {
        if len(dest) < r.Dimensionality() { copy(dest, r[:]) }
        dest[0] = r[0]
        dest[1] = r[1]
        dest[2] = r[2]
        dest[3] = r[3]
    }

    func (r d4D) Index(offsets ... int) int {
        x := offsets[0] % r[0]
        y := offsets[1] % r[1]
        z := offsets[2] % r[2]
        e := offsets[3] % r[3]
        w := r[0]
        h := r[1]
        d := r[2]
        return (e * w * h * d) + (z * w * h) + (y * w) + x
    }

    func (r d4D) Offsets(dest []int, idx int) {
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

// dND is an N-dimensional implementation of the Dimensions interface.
type dND []int

    func (r dND) Dimensionality() int { return len(r) }
    func (r dND) Contains(offsets ... int) bool { return dimensionsContains(r, offsets) }

    func (r dND) Size() int {
        d := r.Dimensionality()
        total := 1
        for i := 0; i < d; i++ {
            total *= r[i]
        }
        return total
    }

    func (r dND) Lengths(dest []int) {
        copy(dest, r)
    }

    func (r dND) Index(offsets ... int) int {
        d := r.Dimensionality()
        stride := 1
        total := 0
        for i := 0; i < d; i++ {
            total += (offsets[i] % r[i]) * stride
            stride *= r[i]
        }
        return total
    }

    func (r dND) Offsets(dest []int, idx int) {
        d := r.Dimensionality()

        dest[0] = idx % r[0]
        stride := r[0]

        for i := 1; i < d; i++ {
            dest[i] = (idx / stride) % r[i]
            stride *= r[i]
        }
    }
