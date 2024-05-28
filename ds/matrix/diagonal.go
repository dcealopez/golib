package matrix

import (
    "errors"
    "fmt"

    "github.com/tawesoft/golib/v2/ks"
)

// ErrDiagonalSize is the type of error raised by a panic if a [Diagonal]
// matrix is resized without equal length sides in each supported dimension.
var ErrDiagonalSize = errors.New("matrix must be diagonal")

// ErrOffDiagonal is the type of error raised by a panic if a value is set in
// a [Diagonal] matrix at an index that does not lie on the diagonal.
type ErrOffDiagonal struct {
    Index [4]int // x, y, w, z index attempted to access
    Width int
    Dimensionality int
}

func (e ErrOffDiagonal) Error() string {
    const s = "set values must lie on the diagonal in a diagonal matrix: got (%d, %d, %d, %d, ...) on a %dD matrix of width %d."
    return fmt.Sprintf(s,
        e.Index[0], e.Index[1], e.Index[2], e.Index[3],
        e.Dimensionality,
        e.Width,
    )
}


// Diagonal implements the matrix [Interface] as a contiguous slice of values
// making up only the diagonal entries. All entries off the diagonal are zero,
// and the matrix must have equal length sides in every dimension.
type Diagonal[T any] struct {
    dimensions
    values []T
}

// NewDiagonal is a matrix [Constructor].
func NewDiagonal[T any]() Interface[T] {
    return &Diagonal[T]{}
}

func (m Diagonal[T]) index(dimensionality, x, y, z, w int) int {
    m.dimensions.check(dimensionality, x, y, z, w)
    ok := false
    switch dimensionality {
        case 4: ok = (x == y) && (y == z) && (z == w)
        case 3: ok = (x == y) && (y == z)
        case 2: ok = (x == y)
        case 1: ok = true
    }
    if !ok { return -1 }
    return x
}

func (m Diagonal[T]) indexN(offsets ... int) int {
    first := offsets[0]
    for i := 1; i < len(offsets); i++ {
        if first != offsets[i] { return -1 }
    }
    return first
}

func (m *Diagonal[T]) resize(dimensionality int, lengths ... int) {
    okDiagonal := (dimensionality >= 1) && (dimensionality == len(lengths))
    first := lengths[0]
    for i := 1; i < len(lengths); i++ {
        if first != lengths[i] { okDiagonal = false; break }
    }
    if !okDiagonal { panic(ErrDiagonalSize) }

    width := lengths[0]
    m.setDimensions(dimensionality, lengths...)
    m.values = ks.SetLength(m.values, width)
    m.Clear()
}

func (m *Diagonal[T]) Clear() {
    if m.values == nil { return }
    clear(m.values[0:cap(m.values)])
}

func (m Diagonal[T]) Get1D(x int) T {
    var zero T
    idx := m.index(1, x, 0, 0, 0)
    if idx < 0 { return zero }
    return m.values[idx]
}

func (m Diagonal[T]) Get2D(x, y int) T {
    var zero T
    idx := m.index(2, x, y, 0, 0)
    if idx < 0 { return zero }
    return m.values[idx]
}

func (m Diagonal[T]) Get3D(x, y, z int) T {
    var zero T
    idx := m.index(3, x, y, z, 0)
    if idx < 0 { return zero }
    return m.values[idx]
}

func (m Diagonal[T]) Get4D(x, y, z, w int) T {
    var zero T
    idx := m.index(4, x, y, z, w)
    if idx < 0 { return zero }
    return m.values[idx]
}

func (m Diagonal[T]) GetN(offset ... int) T {
    var zero T
    idx := m.indexN(offset...)
    if idx < 0 { return zero }
    return m.values[idx]
}

func (m *Diagonal[T]) Set1D(value T, x int) {
    idx := m.index(1, x, 0, 0, 0)
    if idx < 0 { panic(ErrOffDiagonal{
        Index:          [4]int{x, 0, 0, 0},
        Width:          m.width,
        Dimensionality: m.dimensionality,
    }) }
    m.values[idx] = value
}

func (m *Diagonal[T]) Set2D(value T, x, y int) {
    idx := m.index(2, x, y, 0, 0)
    if idx < 0 { panic(ErrOffDiagonal{
        Index:          [4]int{x, y, 0, 0},
        Width:          m.width,
        Dimensionality: m.dimensionality,
    }) }
    m.values[idx] = value
}

func (m *Diagonal[T]) Set3D(value T, x, y, z int) {
    idx := m.index(3, x, y, z, 0)
    if idx < 0 { panic(ErrOffDiagonal{
        Index:          [4]int{x, y, z, 0},
        Width:          m.width,
        Dimensionality: m.dimensionality,
    }) }
    m.values[idx] = value
}

func (m *Diagonal[T]) Set4D(value T, x, y, z, w int) {
    idx := m.index(4, x, y, z, w)
    if idx < 0 { panic(ErrOffDiagonal{
        Index:          [4]int{x, y, z, w},
        Width:          m.width,
        Dimensionality: m.dimensionality,
    }) }
    m.values[idx] = value
}

func (m *Diagonal[T]) SetN(value T, offsets ... int) {
    idx := m.indexN(offsets...)
    if idx < 0 { panic(ErrOffDiagonal{
        Index:          [4]int{},
        Width:          m.width,
        Dimensionality: m.dimensionality,
    }) }
    m.values[idx] = value
}

func (m *Diagonal[T]) Resize1D(w int)                { m.resize(1, w) }
func (m *Diagonal[T]) Resize2D(w, h int)             { m.resize(2, w, h) }
func (m *Diagonal[T]) Resize3D(w, h, d int)          { m.resize(3, w, h, d) }
func (m *Diagonal[T]) Resize4D(w, h, d, x int)       { m.resize(4, w, h, d, x) }
func (m *Diagonal[T]) ResizeN(lengths ... int)       { m.resize(len(lengths), lengths...) }
