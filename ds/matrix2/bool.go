//go:build exclude
package matrix2

import (
    "github.com/tawesoft/golib/v2/ds/bitseq"
)

type boolm struct {
    dimensions
    values bitseq.Store
}

// NewBool is a matrix [Constructor] for a matrix implemented as a contiguous
// sequence of true or false bits. This type is more efficient than [NewGrid]
// for storing single-bit values.
func NewBool() Interface[bool] {
    return &boolm{}
}

func (m boolm) index(dimensionality, x, y, z, w int) int {
    m.dimensions.check(dimensionality, x, y, z, w)
    return (w * m.width * m.height * m.depth) +
        (z * m.width * m.height) +
        (y * m.width) +
        (x)
}

func (m boolm) indexN(offsets ...int) int {
    offset := 0
    stride := 1

    for i := 0; i < len(offsets); i++ {
        offset += offsets[i] * stride
        stride *= m.DimensionN(i)
    }

    return offset
}

func (m *boolm) resize(dimensionality int, lengths ...int) {
    m.setDimensions(dimensionality, lengths...)
    m.Clear()
}

func (m *boolm) Clear() {
    m.values.Clear()
}

func (m boolm) Get1D(x int) bool                  { return m.values.Get(m.index(1, x, 0, 0, 0)) }
func (m boolm) Get2D(x, y int) bool               { return m.values.Get(m.index(2, x, y, 0, 0)) }
func (m boolm) Get3D(x, y, z int) bool            { return m.values.Get(m.index(3, x, y, z, 0)) }
func (m boolm) Get4D(x, y, z, w int) bool         { return m.values.Get(m.index(4, x, y, z, w)) }
func (m boolm) GetN(offsets ...int) bool          { return m.values.Get(m.indexN(offsets...)) }

func (m *boolm) Set1D(value bool, x int)          { m.values.Set(m.index(1, x, 0, 0, 0), value) }
func (m *boolm) Set2D(value bool, x, y int)       { m.values.Set(m.index(2, x, y, 0, 0), value) }
func (m *boolm) Set3D(value bool, x, y, z int)    { m.values.Set(m.index(3, x, y, z, 0), value) }
func (m *boolm) Set4D(value bool, x, y, z, w int) { m.values.Set(m.index(4, x, y, z, w), value) }
func (m *boolm) SetN(value bool, offsets ...int)  { m.values.Set(m.indexN(offsets...), value) }

func (m *boolm) Resize1D(w int)                   { m.resize(1, w) }
func (m *boolm) Resize2D(w, h int)                { m.resize(2, w, h) }
func (m *boolm) Resize3D(w, h, d int)             { m.resize(3, w, h, d) }
func (m *boolm) Resize4D(w, h, d, x int)          { m.resize(4, w, h, d, x) }
func (m *boolm) ResizeN(lengths ...int)           { m.resize(len(lengths), lengths...) }
