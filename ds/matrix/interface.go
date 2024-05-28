// Package matrix implements several different matrix data structures, each
// optimised for different purposes.
//
// In general, this package is for matrices of arbitrary size and dimension.
// Fixed-size matrices (e.g. 4x4) would benefit from a special-purpose package.
package matrix

import (
    "errors"
    "fmt"
)

// ErrNotImplemented is the type of error raised by a panic if a matrix method
// is not implemented. For example, the [Hash] implementation does not support
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
type Constructor[T any] func() Interface[T]

// Interface is the basic interface implemented by any matrix type. It
// maps coordinates (e.g. x, y, z, w) to elements of type T.
type Interface[T any] interface {
    // Resize1D (and variants) updates a matrix, if necessary, so that it has
    // at least capacity for the given number of elements, and has the
    // appropriate dimensionality. It may return a new matrix, but reuses
    // underlying memory from the existing matrix where possible. All values
    // are cleared.
    Resize1D(index int)
    Resize2D(width, height int)
    Resize3D(width, height, depth int)
    Resize4D(width, height, depth, extent int)
    ResizeN(lengths ... int)

    // Width and Height and Depth, and Extent return the number of elements
    // in the x, y, z and w dimensions respectively.
    Width()  int
    Height() int
    Depth()  int
    Extent() int

    // Dimensions2D and Dimensions3D, and Dimensions4D return the number of
    // elements in the (x, y), (x, y, z), and (x, y, z, w) dimensions,
    // respectively.
    Dimensions2D() (width, height int)
    Dimensions3D() (width, height, depth int)
    Dimensions4D() (width, height, depth, extent int)

    // DimensionN returns the number of elements in the nth dimension.
    DimensionN(n int) int

    // Dimensionality returns the number of dimensions. For example, returns 1,
    // 2, 3 or 4 if the matrix is 1D, 2D, 3D or 4D, respectively.
    Dimensionality() int

    // Get1D (and variants) return the value at the given coordinate. This
    // will panic if out of bounds.
    Get1D(x int) T
    Get2D(x, y int) T
    Get3D(x, y, z int) T
    Get4D(x, y, z, w int) T

    // GetN returns the value at the given offsets with arbitrary dimension.
    // Must contain an offset for each dimension.
    GetN(offsets ... int) T

    // Set1D (and variants) writes a value at the given coordinate. This will
    // panic if out of bounds.
    Set1D(value T, x int)
    Set2D(value T, x, y int)
    Set3D(value T, x, y, z int)
    Set4D(value T, x, y, z, w int)

    // SetN writes a value at the given coordinate at the given offsets with
    // arbitrary dimension. Must contain one offset for each dimension.
    SetN(value T, offset ...int)

    // Clear sets all elements to zero.
    Clear()

    // Capacity returns the total number of elements in the matrix (across all
    // dimensions).
    Capacity() int
}

// New1D creates, sizes, and returns a 1D implementation of the matrix interface
// with the given constructor e.g. [NewGrid].
func New1D[T any](implementation Constructor[T], width int) Interface[T] {
    m := implementation()
    m.Resize1D(width)
    return m
}

// New2D creates, sizes, and returns a 2D implementation of the matrix interface
// with the given constructor e.g. [NewGrid].
func New2D[T any](implementation Constructor[T], width, height int) Interface[T] {
    m := implementation()
    m.Resize2D(width, height)
    return m
}

// New3D creates, sizes, and returns a 3D implementation of the matrix
// interface with the given constructor e.g. [NewGrid].
func New3D[T any](implementation Constructor[T], width, height, depth int) Interface[T] {
    m := implementation()
    m.Resize3D(width, height, depth)
    return m
}

// New4D creates, sizes, and returns a 4D implementation of the matrix
// interface with the given constructor e.g. [NewGrid].
func New4D[T any](implementation Constructor[T], width, height, depth, extent int) Interface[T] {
    m := implementation()
    m.Resize4D(width, height, depth, extent)
    return m
}

// NewN creates, sizes, and returns an implementation of the matrix interface
// with the given constructor, e.g. [NewGrid], for a matrix of arbitrary
// dimension.
func NewN[T any](implementation Constructor[T], lengths ... int) Interface[T] {
    m := implementation()
    m.ResizeN(lengths...)
    return m
}

// dimensions holds common information about a sized matrix.
type dimensions struct {
    dimensionality int // 1, 2, 3, 4 or more.
    width, height, depth, extent int
    dimensions []int // set only if dimensionality 5 or above
}

func (m *dimensions) setDimensions(dimensionality int, lengths ... int) {
    n := len(lengths)
    if dimensionality != n || n == 0 {
        panic(ErrDimension{
            Requested: dimensionality,
            Actual:    n,
        })
    }

    m.dimensionality = dimensionality
    m.width = 1; m.height = 1; m.depth = 1; m.extent = 1
    m.dimensions = nil

    switch {
        case (n >= 5): {
            m.dimensions = make([]int, (n - 4))
            for i := 4; i < n; i++ {
                m.dimensions[i-4] = lengths[i]
            }
        }; fallthrough
        case (n >= 4): m.extent = lengths[3]; fallthrough
        case (n >= 3): m.depth  = lengths[2]; fallthrough
        case (n >= 2): m.height = lengths[1]; fallthrough
        case (n >= 1): m.width  = lengths[0]
    }
}

// check (for 1D, 2D, 3D, or 4D accesses) verifies a correct dimensionality
// and offset.
func (m dimensions) check(dimensionality, x, y, z, w int) {
    okDimension := (m.dimensionality == dimensionality)
    okDimension = okDimension && (dimensionality >= 1) && (dimensionality <= 4)
    okNegative := (x >= 0) && (y >= 0) && (z >= 0) && (w >= 0)
    okPositive := (x < m.width) && (y < m.height) && (z < m.depth) && (w < m.extent)
    if !okDimension {
        panic(ErrDimension{
            Requested: dimensionality,
            Actual:    m.dimensionality,
        })
    }
    if (!okNegative) && (!okPositive) {
        panic(ErrIndexOutOfRange{
            Index: [4]int{x, y, z, w},
            Range: [4]int{m.width, m.height, m.depth, m.extent},
        })
    }
}

func (m dimensions) DimensionN(n int) int {
    okDimension := (n >= 1) && (n <= m.dimensionality)
    if !okDimension {
        panic(ErrDimension{
            Requested: n,
            Actual:    m.dimensionality,
        })
    }
    switch n {
        case 1: return m.width
        case 2: return m.height
        case 3: return m.depth
        case 4: return m.extent
    }
    // m.dimensions is not nil if n >= 5
    return m.dimensions[n - 5]
}

func (m dimensions) Capacity() int {
    if m.dimensionality <= 4 {
        return m.width * m.height * m.depth * m.extent
    } else {
        size := m.width * m.height * m.depth * m.extent
        for i := 0; i < len(m.dimensions); i++ {
            size *= m.dimensions[i]
        }
        return size
    }
}

func (m dimensions) Dimensionality() int                { return m.dimensionality }
func (m dimensions) Width()  int                        { return m.width }
func (m dimensions) Height() int                        { return m.height }
func (m dimensions) Depth()  int                        { return m.depth }
func (m dimensions) Extent() int                        { return m.extent }
func (m dimensions) Dimensions1D() (int)                { return m.width }
func (m dimensions) Dimensions2D() (int, int)           { return m.width, m.height }
func (m dimensions) Dimensions3D() (int, int, int)      { return m.width, m.height, m.depth }
func (m dimensions) Dimensions4D() (int, int, int, int) { return m.width, m.height, m.depth, m.extent }
