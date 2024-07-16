// Package matrix implements several different matrix data structures, each
// optimised for different purposes.
//
// In general, this package is for matrices of arbitrary size and dimension.
// Fixed-size matrices (e.g. 4x4) would benefit from a special-purpose package.
package matrix2

// TODO: clean up error handling with a separate wrapper interface.

import (
    "errors"
    "fmt"

    "github.com/tawesoft/golib/v2/ds/matrix/indexes"
)

// ErrNotImplemented is the type of error raised by a panic if a matrix method
// is not implemented. For example, the [hash] implementation does not support
// an arbitrary dimension.
var ErrNotImplemented = errors.New("not implemented")

// ErrIndexOutOfRange is the type of error raised by a panic if a matrix is
// accessed with a coordinate lying outside its range.
type ErrIndexOutOfRange struct {
    Index [4]int // x, y, w, z index attempted to access
    Range [4]int // width, height, depth, extent of each dimension
}

func (e ErrIndexOutOfRange) Error() string {
    const s = "matrix index out of range: expected (0, 0, 0, 0) <= (%d, %d, %d, %d) < (%d, %d, %d, %d)"
    return fmt.Sprintf(s,
        e.Index[0], e.Index[1], e.Index[2], e.Index[3],
        e.Range[0], e.Range[1], e.Range[2], e.Range[3],
    )
}

// ErrDimension is the type of error raised by a panic if a matrix is
// accessed using coordinates with the wrong dimensionality, for example (x, y)
// coordinates in a 3D matrix expecting (x, y, z) coordinates.
type ErrDimension struct {
    Requested, Actual int // 1, 2, 3 or 4.
}

func (e ErrDimension) Error() string {
    const s = "matrix dimension mismatch: %dD access, but matrix is %dD"
    return fmt.Sprintf(s, e.Requested, e.Actual)
}

// Constructor is any function that returns a new matrix - e.g. [NewGrid].
type Constructor[T comparable] func() Interface[T]

// Interface is the basic interface implemented by any matrix type. It
// maps coordinates (e.g. x, y, z, w) to elements of type T.
type Interface[T comparable] interface {
    // Resize1D (and variants) resizes a matrix, if necessary, so that it has
    // at least capacity for the given number of elements, and has the
    // appropriate dimensionality.
    //
    // If growing, existing values are persisted  at their existing indexes.
    // If shrinking, some elements will be cropped out and zeroed. This may,
    // depending on the implementation, have to copy and move elements around
    // in memory.
    //
    // To set the matrix's size and dimension and clear it at the same time,
    // it is more efficient to use SetSize1D (and variants).
    //
    // To minimise memory allocations when growing a matrix, resizing the
    // matrix reuses underlying memory from the existing matrix where possible,
    // may allocate more than immediately needed, and may not release memory
    // when resized to a smaller size. See the Compact method to reclaim
    // memory.
    Resize1D(index int)
    Resize2D(width, height int)
    Resize3D(width, height, depth int)
    Resize4D(width, height, depth, extent int)
    ResizeN(lengths ... int)

    // SetSize1D (and variants) resizes a matrix, if necessary, so that it has
    // at least capacity for the given number of elements, and has the
    // appropriate dimensionality. Unlike Resize1D (and variants), this also
    // clears the matrix, setting each value to zero.
    //
    // To minimise memory allocations when growing a matrix, resizing the
    // matrix reuses underlying memory from the existing matrix where possible,
    // may allocate more than immediately needed, and may not release memory
    // when resized to a smaller size. See the Compact method to reclaim
    // memory.
    SetSize1D(index int)
    SetSize2D(width, height int)
    SetSize3D(width, height, depth int)
    SetSize4D(width, height, depth, extent int)
    SetSizeN(lengths ... int)

    // Width and Height and Depth, and Extent return the size of the matrix
    // in the x, y, z and w dimensions respectively.
    Width()  int
    Height() int
    Depth()  int
    Extent() int

    // Volume returns the result of multiplying the size of the matrix in
    // each direction e.g. for a 2D matrix, this is width times height.
    Volume() int

    // Dimensions2D and Dimensions3D, and Dimensions4D return the number of
    // elements in the (x, y), (x, y, z), and (x, y, z, w) dimensions,
    // respectively.
    Dimensions2D() (width, height int)
    Dimensions3D() (width, height, depth int)
    Dimensions4D() (width, height, depth, extent int)

    // DimensionN returns the number of elements in the given zero-indexed
    // dimension, e.g. for n == 0, 1D; for n == 1, 2D ...
    DimensionN(n int) int

    // Dimensions returns a slice containing the number of elements in each
    // dimension. The returned slice must not be modified, and is only valid
    // until the matrix is resized.
    Dimensions() []int

    // Dimensionality returns the number of dimensions. For example, returns 1,
    // 2, 3 or 4 if the matrix is 1D, 2D, 3D or 4D, respectively.
    Dimensionality() int

    // Get1D (and variants) return the value at the given coordinate.
    //
    // If an offset is out of bounds, the behaviour, including possible panics,
    // is defined by the implementation. As a special case, an offset in an
    // undefined dimension is allowed only if the offset is zero. For example,
    // Get3D(3, 3, 0) is the same as Get2D(3, 3).
    Get1D(x int) T
    Get2D(x, y int) T
    Get3D(x, y, z int) T
    Get4D(x, y, z, w int) T
    GetN(offsets ... int) T

    // Set1D (and variants) writes a value at the given coordinate.
    //
    //
    // If an offset is out of bounds, the behaviour, including possible panics,
    // is defined by the implementation. As a special case, an offset in an
    // undefined dimension is allowed only if the offset is zero. For example,
    // Set3D(value, 3, 3, 0) is the same as Set3D(value, 3, 3).
    Set1D(value T, x int)
    Set2D(value T, x, y int)
    Set3D(value T, x, y, z int)
    Set4D(value T, x, y, z, w int)
    SetN(value T, offset ...int)

    // Indexes returns an iterator that generates every index in the matrix.
    Indexes() indexes.Iterator

    // SparseIndexes returns an iterator that generates indexes that have been
    // set to a non-zero value. Depending on the implementation, this may be
    // more efficient than a general iteration of indexes.
    SparseIndexes() indexes.Iterator

    // Clear sets all elements to zero.
    Clear()

    // Compact reallocates the matrix, if necessary and possible, to achieve as
    // compact as possible in-memory representation, possibly copying and
    // moving existing elements to newly allocated memory.
    Compact()
}

// New1D creates, sizes, and returns a 1D implementation of the matrix
// interface with the given constructor e.g. [NewGrid].
func New1D[T comparable](implementation Constructor[T], width int) Interface[T] {
    m := implementation()
    m.SetSize1D(width)
    return m
}

// New2D creates, sizes, and returns a 2D implementation of the matrix interface
// with the given constructor e.g. [NewGrid].
func New2D[T comparable](implementation Constructor[T], width, height int) Interface[T] {
    m := implementation()
    m.SetSize2D(width, height)
    return m
}

// New3D creates, sizes, and returns a 3D implementation of the matrix
// interface with the given constructor e.g. [NewGrid].
func New3D[T comparable](implementation Constructor[T], width, height, depth int) Interface[T] {
    m := implementation()
    m.SetSize3D(width, height, depth)
    return m
}

// New4D creates, sizes, and returns a 4D implementation of the matrix
// interface with the given constructor e.g. [NewGrid].
func New4D[T comparable](implementation Constructor[T], width, height, depth, extent int) Interface[T] {
    m := implementation()
    m.SetSize4D(width, height, depth, extent)
    return m
}

// NewN creates, sizes, and returns an implementation of the matrix interface
// with the given constructor, e.g. [NewGrid], for a matrix of arbitrary
// dimension.
func NewN[T comparable](implementation Constructor[T], lengths ... int) Interface[T] {
    m := implementation()
    m.SetSizeN(lengths...)
    return m
}
