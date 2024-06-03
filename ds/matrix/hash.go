package matrix

type hash[T any] struct {
    dimensions
    mapping1D map[int]T
    mapping2D map[[2]int]T
    mapping3D map[[3]int]T
    mapping4D map[[4]int]T
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
    m.check(1, x, 0, 0, 0)
    value, _ := m.mapping1D[x]
    return value
}

func (m hash[T]) Get2D(x, y int) T {
    m.check(2, x, y, 0, 0)
    value, _ := m.mapping2D[[2]int{x, y}]
    return value
}

func (m hash[T]) Get3D(x, y, z int) T {
    m.check(3, x, y, z, 0)
    value, _ := m.mapping3D[[3]int{x, y, z}]
    return value
}

func (m hash[T]) Get4D(x, y, z, w int) T {
    m.check(4, x, y, z, w)
    value, _ := m.mapping4D[[4]int{x, y, z, w}]
    return value
}

func (m *hash[T]) Set1D(value T, x int) {
    m.check(1, x, 0, 0, 0)
    m.mapping1D[x] = value
}

func (m *hash[T]) Set2D(value T, x, y int) {
    m.check(2, x, y, 0, 0)
    m.mapping2D[[2]int{x, y}] = value
}

func (m *hash[T]) Set3D(value T, x, y, z int) {
    m.check(3, x, y, z, 0)
    m.mapping3D[[3]int{x, y, z}] = value
}

func (m *hash[T]) Set4D(value T, x, y, z, w int) {
    m.check(4, x, y, z, w)
    m.mapping4D[[4]int{x, y, z, w}] = value
}

func (m *hash[T]) resize(dimensionality, width, height, depth, extent int) Interface[T] {
    m.dimensionality = dimensionality
    m.width  = width
    m.height = height
    m.depth  = depth
    m.extent = extent

    switch m.dimensionality {
        case 1: {
            if m.mapping1D == nil {
                m.mapping1D = make(map[int]T)
            }
            m.mapping2D = nil
            m.mapping3D = nil
            m.mapping4D = nil
        }
        case 2: {
            m.mapping1D = nil
            if m.mapping2D == nil {
                m.mapping2D = make(map[[2]int]T)
            }
            m.mapping3D = nil
            m.mapping4D = nil
        }
        case 3: {
            m.mapping1D = nil
            m.mapping2D = nil
            if m.mapping3D == nil {
                m.mapping3D = make(map[[3]int]T)
            }
            m.mapping4D = nil
        }
        case 4: {
            m.mapping1D = nil
            m.mapping2D = nil
            m.mapping3D = nil
            if m.mapping4D == nil {
                m.mapping4D = make(map[[4]int]T)
            }
        }
    }

    m.Clear()
    return m
}

func (m *hash[T]) Clear() {
    clear(m.mapping1D)
    clear(m.mapping2D)
    clear(m.mapping3D)
    clear(m.mapping4D)
}

func (m hash[T]) Reduce(identity T, sum func(a, b T) T) T {
    total := identity
    switch m.dimensionality {
        case 1: {
            for _, v := range m.mapping1D {
                total = sum(total, v)
            }
        }
        case 2: {
            for _, v := range m.mapping2D {
                total = sum(total, v)
            }
        }
        case 3: {
            for _, v := range m.mapping3D {
                total = sum(total, v)
            }
        }
        case 4: {
            for _, v := range m.mapping4D {
                total = sum(total, v)
            }
        }
    }
    // case 0 is fine, too
    return total
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
