package matrix2

// Dimension represents a finite
type Dimension interface {
    // Index computes an index from the offsets into each dimension.
    Index(offsets ...int) int

    // Offsets calculates the offset in each dimension identified by the given
    // index, in row-major order. The results are stored in dest.
    Offsets(dest []int, idx int)

    // Size returns the number of unique indexes, from 0 to size minus 1
    // inclusive, that can be used to obtain offsets.
    Size() int

    // Dimensionality returns the number of dimensions e.g. 2 for "2D".
    Dimensionality() int

    // Lengths returns the length of each dimension. The results are stored in
    // dest.
    Lengths(dest []int)
}

// d2D is a 2-dimensional implementation of the D interface.
type d2D struct { width, height int }

    func (r d2D) Size() int { return r.width * r.height }
    func (r d2D) Dimensionality() int { return 2 }

    func (r d2D) Lengths(dest []int) {
        dest[0] = r.width
        dest[1] = r.height
    }

    func (r d2D) Index(offsets ... int) int {
        x, y := offsets[0], offsets[1]
        return (y * r.width) + x
    }

    func (r d2D) Offsets(dest []int, idx int) {
        w, h := r.width, r.height
        x := idx % w
        y := (idx / w) % h
        dest[0] = x
        dest[1] = y
    }
