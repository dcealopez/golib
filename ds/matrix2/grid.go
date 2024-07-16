package matrix2

import (
    "github.com/tawesoft/golib/v2/ds/matrix/indexes"
    "github.com/tawesoft/golib/v2/fun/slices"
    "github.com/tawesoft/golib/v2/must"
    "github.com/tawesoft/golib/v2/operator"
)

// Dev note: like Go's strategy for reallocating a slice with spare memory,
// we want to grow a slice backed by an array that grows in several dimensions
// (even if ultimately mapped to a 1D array).
//
// As Go historically did for slices, we could grow by x1.25 in each dimension:
// for n = 1 to 4 and a regular matrix, total memory then increases by x1.25,
// ~x1.56, ~x1.95, ~x2.44... We should only grow a dimension if that
// dimension's size changes, to not waste memory when supporting
// rectangular/irregular matrices.
//
// For lengths of 4 or more, growing by x1.25 will trivially always grow that
// length by at least one. But it seems reasonable to grow quicker for small
// lengths, as Go does for small slices where it simply doubles the old
// capacity. Of course, in the case where we increase lengths in n dimensions
// at once, we would not be multiplying the total memory by two, but by 2^n,
// which is far too much unless the lengths are all very small!
//
// So:
// For lengths of 1 to 3: x2
// For lengths of 4 to 11: x1.5 (roughly 2^(n/2) but cheaper to calculate!)
// For lengths of 12 or more: x1.25
//
// We can also consider that Go may give us more capacity than we asked for,
// due to malloc size classes. We don't take advantage of this straight away,
// but only at the next resize, to keep the implementation simple.
func nextLength(x int) int {
    switch {
        case (x >= 12): return x + (x/4)
        case (x >=  4): return x + (x/2)
        default:        return x + x
    }
}

type grid[T comparable] struct {
    // matrix as seen by the user
    dimensions

    // matrix including spare memory in a backing array
    mem dimensions
    values []T
}

// NewGrid is a matrix [Constructor] for a matrix implemented as a contiguous
// slice of values. This type is more efficient for representing dense matrices
// (i.e. one where few values are zero).
func NewGrid[T comparable]() Interface[T] {
    return &grid[T]{}
}

func (m *grid[T]) setSize(makeCopy bool, lengths ...int) {
    if m.values == nil {
        m.setDimensions(lengths...)
        m.mem.setDimensions(lengths...)
        m.values = make([]T, m.Volume())
        return
    }

    // grow
    wantedVolume := slices.Reduce(1, operator.Mul[int], lengths)
    memLengths := slices.Map(nextLength, lengths)
    memVolume := slices.Reduce(1, operator.Mul[int], memLengths)
    if (wantedVolume > cap(m.values)) {
        // reserve a bigger matrix than necessary...
        buf := make([]T, memVolume)
        var newMem dimensions
        newMem.setDimensions(memLengths...)
        if makeCopy {
            gridMove(m, buf, newMem)
        }
        m.values = buf
        m.setDimensions(lengths...)
        m.mem.setDimensions(memLengths...)
        return
    }

    // arrange in existing memory without reallocation.
    //
    // We need a temporary buffer to hold the cropped results, but it
    // shouldn't need to escape to the heap.
    //
    // m.mem, the backing-array representation, stays at its larger size if
    // already big enough, but m.dimensions might get smaller.
    fits := m.mem.dimensionality >= len(lengths)
    if fits {
        dims := m.mem.Dimensions()
        for i := 0; i < len(lengths); i++ {
            if dims[i] < lengths[i] { fits = false; break }
        }
    }
    if makeCopy {
        var newMem dimensions
        var buf []T
        if fits {
            buf = make([]T, m.mem.Volume())
            newMem.setDimensions(m.mem.Dimensions()...)
        } else {
            buf = make([]T, wantedVolume)
            newMem.setDimensions(lengths...)
        }
        gridMove(m, buf, newMem)
        clear(m.values)
        copy(m.values, buf)
        m.setDimensions(lengths...)
    } else {
        m.setDimensions(lengths...)
        m.Clear()
    }
    if !fits {
        m.mem.setDimensions(lengths...)
    }
}

func (m *grid[T]) Compact() {
    if m.values == nil { return }

    volume := m.Volume()
    if volume == 0 {
        m.values = nil
        return
    }

    // nothing to do
    if m.dimensions.equals(m.mem) { return }

    // aligns m.memDimensions with m.dimensions
    buf := make([]T, volume)
    gridMove(m, buf, m.dimensions)
    m.values = buf
    m.mem.set(m.dimensions)
}

func (m *grid[T]) Clear() {
    if m.values == nil { return }
    clear(m.values[0:cap(m.values)])
}

// gridMove copies the values in m to the buffer dest. destDims indicates
// the layout of the buffer.
func gridMove[T comparable](source Interface[T], dest []T, destDims dimensions) {
    iter := indexes.Forwards(source.Dimensions()...)
    indexes := make([]int, source.Dimensionality())
    for iter(indexes) {
        if !destDims.contains(indexes...) { continue }
        idx := destDims.indexN(indexes...)
        must.Truef(idx < len(dest),
            "index out of range: dest access %v (offset %d, length %d) for dest size %v and source size %v",
            indexes, idx, len(dest), destDims.Dimensions(), source.Dimensions())
        dest[idx] = source.GetN(indexes...)
    }
}

func (m grid[T]) SparseIndexes() indexes.Iterator {
    return m.Indexes()
    /*
    var zero T
    iter := indexes.Forwards(m.Dimensions()...)
    return func(dest []int) bool {
        for iter(dest) {
            if zero == m.values[m.mem.indexN(dest...)] { continue }
            return true
        }
        return false
    }
    */
}

func (m grid[T]) Get1D(x int) T                  { return m.values[m.mem.index1D(x)] }
func (m grid[T]) Get2D(x, y int) T               { return m.values[m.mem.index2D(x, y)] }
func (m grid[T]) Get3D(x, y, z int) T            { return m.values[m.mem.index3D(x, y, z)] }
func (m grid[T]) Get4D(x, y, z, w int) T         { return m.values[m.mem.index4D(x, y, z, w)] }
func (m grid[T]) GetN(offsets ...int) T          { return m.values[m.mem.indexN(offsets...)] }

func (m *grid[T]) Set1D(value T, x int)          { m.values[m.mem.index1D(x)] = value }
func (m *grid[T]) Set2D(value T, x, y int)       { m.values[m.mem.index2D(x, y)] = value }
func (m *grid[T]) Set3D(value T, x, y, z int)    { m.values[m.mem.index3D(x, y, z)] = value }
func (m *grid[T]) Set4D(value T, x, y, z, w int) { m.values[m.mem.index4D(x, y, z, w)] = value }
func (m *grid[T]) SetN(value T, offsets ...int)  { m.values[m.mem.indexN(offsets...)] = value }

func (m *grid[T]) SetSize1D(w int)               { m.setSize(false, w) }
func (m *grid[T]) SetSize2D(w, h int)            { m.setSize(false, w, h) }
func (m *grid[T]) SetSize3D(w, h, d int)         { m.setSize(false, w, h, d) }
func (m *grid[T]) SetSize4D(w, h, d, x int)      { m.setSize(false, w, h, d, x) }
func (m *grid[T]) SetSizeN(lengths ...int)       { m.setSize(false, lengths...) }

func (m *grid[T]) Resize1D(w int)                { m.setSize(true, w) }
func (m *grid[T]) Resize2D(w, h int)             { m.setSize(true, w, h) }
func (m *grid[T]) Resize3D(w, h, d int)          { m.setSize(true, w, h, d) }
func (m *grid[T]) Resize4D(w, h, d, x int)       { m.setSize(true, w, h, d, x) }
func (m *grid[T]) ResizeN(lengths ...int)        { m.setSize(true, lengths...) }
