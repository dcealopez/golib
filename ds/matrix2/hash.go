//go:build exclude
package matrix2

type hash[T any] struct {
    dimensions
    mapping map[[4]int]T
}

// NewHash is a matrix [Constructor] implemented as a hash map with coordinate
// pairs forming keys. This type is more efficient for representing sparse
// matrices (i.e. one where most values are zero).
//
// As a rule-of-thumb, prefer this type if less than 1/64th of the matrix has
// a non-zero value.
//
// This type only implements 1- to 4-dimensional matrices.
func NewHash[T any]() Interface[T] {
    return &hash[T]{}
}

func (m hash[T]) Get1D(x int) T {
    value, _ := m.mapping[[4]int{x, 0, 0, 0}]
    return value
}

func (m hash[T]) Get2D(x, y int) T {
    value, _ := m.mapping[[4]int{x, y, 0, 0}]
    return value
}

func (m hash[T]) Get3D(x, y, z int) T {
    value, _ := m.mapping[[4]int{x, y, z, 0}]
    return value
}

func (m hash[T]) Get4D(x, y, z, w int) T {
    value, _ := m.mapping[[4]int{x, y, z, w}]
    return value
}

func (m *hash[T]) Set1D(value T, x int) {
    m.mapping[[4]int{x, 0, 0, 0}] = value
}

func (m *hash[T]) Set2D(value T, x, y int) {
    m.mapping[[4]int{x, y, 0, 0}] = value
}

func (m *hash[T]) Set3D(value T, x, y, z int) {
    m.mapping[[4]int{x, y, z, 0}] = value
}

func (m *hash[T]) Set4D(value T, x, y, z, w int) {
    m.check(4, x, y, z, w)
    m.mapping[[4]int{x, y, z, w}] = value
}

func (m *hash[T]) resize(dimensionality, width, height, depth, extent int) Interface[T] {
    m.dimensionality = dimensionality
    m.width  = width
    m.height = height
    m.depth  = depth
    m.extent = extent
    m.mapping = make(map[[4]int]T)
    m.Clear()
    return m
}

func (m *hash[T]) Clear() {
    clear(m.mapping)
}

func (m *hash[T]) Resize1D(width int) {
    m.resize(1, width, 1, 1, 1)
}

func (m *hash[T]) Resize2D(width, height int) {
    m.resize(2, width, height, 1, 1)
}

func (m *hash[T]) Resize3D(width, height, depth int) {
    m.resize(3, width, height, depth, 1)
}

func (m *hash[T]) Resize4D(width, height, depth, extent int) {
    m.resize(4, width, height, depth, extent)
}

func (m *hash[T]) ResizeN(lengths ... int) {
    switch len(lengths) {
        case 1: m.Resize1D(lengths[0])
        case 2: m.Resize2D(lengths[0], lengths[1])
        case 3: m.Resize3D(lengths[0], lengths[1], lengths[2])
        case 4: m.Resize4D(lengths[0], lengths[1], lengths[2], lengths[3])
        default: panic(ErrNotImplemented)
    }
}

func (m *hash[T]) GetN(offsets ... int) T {
    switch len(offsets) {
        case 1: return m.Get1D(offsets[0])
        case 2: return m.Get2D(offsets[0], offsets[1])
        case 3: return m.Get3D(offsets[0], offsets[1], offsets[2])
        case 4: return m.Get4D(offsets[0], offsets[1], offsets[2], offsets[3])
        default: panic(ErrNotImplemented)
    }
}

func (m *hash[T]) SetN(value T, offsets ... int) {
    switch len(offsets) {
        case 1: m.Set1D(value, offsets[0])
        case 2: m.Set2D(value, offsets[0], offsets[1])
        case 3: m.Set3D(value, offsets[0], offsets[1], offsets[2])
        case 4: m.Set4D(value, offsets[0], offsets[1], offsets[2], offsets[3])
        default: panic(ErrNotImplemented)
    }
}
