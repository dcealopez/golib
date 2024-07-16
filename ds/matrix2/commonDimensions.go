package matrix2

import (
    "slices"

    "github.com/tawesoft/golib/v2/ds/matrix/indexes"
)

// dimensions holds common information about a sized matrix.
type dimensions struct {
    dimensionality int // 1, 2, 3, 4 or more.
    dims [4]int

    // dimensions slices dims if dimensionality < 5; otherwise it is
    // allocated.
    dimensions []int
}

func (m dimensions) Indexes() indexes.Iterator {
    return indexes.Forwards(m.dimensions...)
}

func (m *dimensions) setDimensions(lengths ... int) {
    n := len(lengths)
    m.dimensionality = n
    m.dims = [4]int{1, 1, 1, 1} // defaults

    switch {
        case (n >= 5): {
            m.dimensions = make([]int, len(lengths))
            for i := 0; i < len(lengths); i++ {
                m.dimensions[i] = lengths[i]
            }
            for i := 0; i < len(lengths); i++ {
                if i >= 4 { break }
                m.dims[i] = lengths[i]
            }
            return
        };
        case (n >= 4): m.dims[3] = lengths[3]; fallthrough
        case (n >= 3): m.dims[2] = lengths[2]; fallthrough
        case (n >= 2): m.dims[1] = lengths[1]; fallthrough
        case (n >= 1): m.dims[0] = lengths[0]
    }
    m.dimensions = m.dims[0:n]
}

func (m dimensions) Volume() int {
    if m.dimensionality <= 4 {
        width, height, depth, extent := m.dims[0], m.dims[1], m.dims[2], m.dims[3]
        return width * height * depth * extent
    } else {
        size := 1
        for i := 0; i < len(m.dimensions); i++ {
            size *= m.dimensions[i]
        }
        return size
    }
}

func (m dimensions) equals(other dimensions) bool {
    return (
        m.dimensionality == other.dimensionality &&
        m.dims == other.dims &&
        slices.Equal(m.dimensions, other.dimensions))
}

func (m *dimensions) set(source dimensions)  {
        m.dimensionality = source.dimensionality
        copy(m.dims[:], source.dims[:])
        m.dimensions = append([]int{}, source.dimensions...)
}

func (m dimensions) index1D(x int) int {
    return x
}

func (m dimensions) index2D(x, y int) int {
    width := m.dims[0]
    return (y * width) + x
}

func (m dimensions) index3D(x, y, z int) int {
    width, height := m.dims[0], m.dims[1]
    return (z * width * height) +
        (y * width) +
        (x)
}

func (m dimensions) index4D(x, y, z, w int) int {
    width, height, depth := m.dims[0], m.dims[1], m.dims[2]
    return (w * width * height * depth) +
        (z * width * height) +
        (y * width) +
        (x)
}

func (m dimensions) indexN(offsets ...int) int {
    offset := 0
    stride := 1

    // As a special case, an offset in an undefined dimension is allowed only
    // if the offset is zero.
    for i := 0; i < len(offsets); i++ {
        if (i >= m.dimensionality) && (offsets[i] != 0) {
            panic("index out of bounds")
        }
        offset += offsets[i] * stride
        stride *= m.DimensionN(i)
    }

    return offset
}

func (m dimensions) DimensionN(n int) int {
    if n < 0 { return 0 }
    if n >= m.dimensionality { return 0 }
    return m.dimensions[n]
}

func (m dimensions) contains(offsets ... int) bool {
    // As a special case, an offset in an undefined dimension is allowed only
    // if the offset is zero.
    for i := 0; i < len(offsets); i++ {
        if i >= m.Dimensionality() && (offsets[i] != 0) { return false }
        if offsets[i] >= m.DimensionN(i) { return false }
    }
    return true
}

func (m dimensions) Dimensionality() int                { return m.dimensionality }
func (m dimensions) Width()  int                        { return m.dims[0] }
func (m dimensions) Height() int                        { return m.dims[1] }
func (m dimensions) Depth()  int                        { return m.dims[2] }
func (m dimensions) Extent() int                        { return m.dims[3] }
func (m dimensions) Dimensions1D() (int)                { return m.dims[0] }
func (m dimensions) Dimensions2D() (int, int)           { return m.dims[0], m.dims[1] }
func (m dimensions) Dimensions3D() (int, int, int)      { return m.dims[0], m.dims[1], m.dims[2] }
func (m dimensions) Dimensions4D() (int, int, int, int) { return m.dims[0], m.dims[1], m.dims[2], m.dims[3] }
func (m dimensions) Dimensions() []int                  { return m.dimensions }
