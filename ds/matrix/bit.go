package matrix

import (
    "github.com/tawesoft/golib/v2/ds/bitseq"
)

// Bit implements a bit matrix, implementing the matrix [Interface], as a
// contiguous sequence of bits with integer value 1 or 0.
//
// This type is more efficient than [Grid] for storing single-bit values.
//
// When setting a value in the matrix, any non-zero integer is treated as true.
type Bit struct {
    dimensions
    values bitseq.Store
}

func fromBool(b bool) int {
    if b { return 1 } else { return 0 }
}

func toBool(i int) bool {
    return (i != 0)
}

// NewBit is a matrix [Constructor].
func NewBit() Interface[int] {
    return &Bit{}
}

func (m Bit) index(dimensionality, x, y, z, w int) int {
    m.dimensions.check(dimensionality, x, y, z, w)
    return (w * m.width * m.height * m.depth) +
           (z * m.width * m.height) +
           (y * m.width) +
           (x)
}

func (m Bit) indexN(offsets ... int) int {
    offset := 0
    stride := 1

    for i := 0; i < len(offsets); i++ {
        offset += offsets[i] * stride
        stride *= m.DimensionN(i)
    }

    return offset
}

func (m *Bit) resize(dimensionality int, lengths ... int) {
    m.setDimensions(dimensionality, lengths...)
    m.Clear()
}

func (m *Bit) Clear() {
    m.values.Clear()
}

func (m Bit) Get1D(x int) int                  { return fromBool(m.values.Get(m.index(1, x, 0, 0, 0))) }
func (m Bit) Get2D(x, y int) int               { return fromBool(m.values.Get(m.index(2, x, y, 0, 0))) }
func (m Bit) Get3D(x, y, z int) int            { return fromBool(m.values.Get(m.index(3, x, y, z, 0))) }
func (m Bit) Get4D(x, y, z, w int) int         { return fromBool(m.values.Get(m.index(4, x, y, z, w))) }
func (m Bit) GetN(offsets ... int) int         { return fromBool(m.values.Get(m.indexN(offsets...))) }

func (m *Bit) Set1D(value int, x int)          { m.values.Set(m.index(1, x, 0, 0, 0), toBool(value)) }
func (m *Bit) Set2D(value int, x, y int)       { m.values.Set(m.index(2, x, y, 0, 0), toBool(value)) }
func (m *Bit) Set3D(value int, x, y, z int)    { m.values.Set(m.index(3, x, y, z, 0), toBool(value)) }
func (m *Bit) Set4D(value int, x, y, z, w int) { m.values.Set(m.index(4, x, y, z, w), toBool(value)) }
func (m *Bit) SetN(value int, offsets ... int) { m.values.Set(m.indexN(offsets...), toBool(value)) }

func (m *Bit) Resize1D(w int)                  { m.resize(1, w) }
func (m *Bit) Resize2D(w, h int)               { m.resize(2, w, h) }
func (m *Bit) Resize3D(w, h, d int)            { m.resize(3, w, h, d) }
func (m *Bit) Resize4D(w, h, d, x int)         { m.resize(4, w, h, d, x) }
func (m *Bit) ResizeN(lengths ... int)         { m.resize(len(lengths), lengths...) }
