package matrix

import (
	"github.com/tawesoft/golib/v2/ks"
)

type grid[T any] struct {
    dimensions
    values []T
}

// NewGrid is a matrix [Constructor] for a matrix implemented as a contiguous
// slice of values. This type is more efficient for representing dense matrices
// (i.e. one where few values are zero).
func NewGrid[T any]() Interface[T] {
    return &grid[T]{}
}

func (m grid[T]) index(dimensionality, x, y, z, w int) int {
    m.dimensions.check(dimensionality, x, y, z, w)
    return (w * m.width * m.height * m.depth) +
        (z * m.width * m.height) +
        (y * m.width) +
        (x)
}

func (m grid[T]) indexN(offsets ...int) int {
    offset := 0
    stride := 1

    for i := 0; i < len(offsets); i++ {
        offset += offsets[i] * stride
        stride *= m.DimensionN(i)
    }

    return offset
}

func (m *grid[T]) resize(dimensionality int, lengths ...int) {
    m.setDimensions(dimensionality, lengths...)
    m.values = ks.SetLength(m.values, m.Capacity())
    m.Clear()
}

func (m *grid[T]) Clear() {
    if m.values == nil {
        return
    }
    clear(m.values[0:cap(m.values)])
}

func (m grid[T]) Get1D(x int) T                  { return m.values[m.index(1, x, 0, 0, 0)] }
func (m grid[T]) Get2D(x, y int) T               { return m.values[m.index(2, x, y, 0, 0)] }
func (m grid[T]) Get3D(x, y, z int) T            { return m.values[m.index(3, x, y, z, 0)] }
func (m grid[T]) Get4D(x, y, z, w int) T         { return m.values[m.index(4, x, y, z, w)] }
func (m grid[T]) GetN(offsets ...int) T          { return m.values[m.indexN(offsets...)] }

func (m *grid[T]) Set1D(value T, x int)          { m.values[m.index(1, x, 0, 0, 0)] = value }
func (m *grid[T]) Set2D(value T, x, y int)       { m.values[m.index(2, x, y, 0, 0)] = value }
func (m *grid[T]) Set3D(value T, x, y, z int)    { m.values[m.index(3, x, y, z, 0)] = value }
func (m *grid[T]) Set4D(value T, x, y, z, w int) { m.values[m.index(4, x, y, z, w)] = value }
func (m grid[T]) SetN(value T, offsets ...int)   { m.values[m.indexN(offsets...)] = value }

func (m *grid[T]) Resize1D(w int)                { m.resize(1, w) }
func (m *grid[T]) Resize2D(w, h int)             { m.resize(2, w, h) }
func (m *grid[T]) Resize3D(w, h, d int)          { m.resize(3, w, h, d) }
func (m *grid[T]) Resize4D(w, h, d, x int)       { m.resize(4, w, h, d, x) }
func (m *grid[T]) ResizeN(lengths ...int)        { m.resize(len(lengths), lengths...) }
