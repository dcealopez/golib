package graph

// AdjacencyMatrix represents the number of directed edges from a source vertex
// (along the x-axis) to a target vertex (along the y-axis), including
// self-loops, if any.
//
// As a representation of a graph itself, AdjacencyMatrix implements the graph
// [Iterator] interface. As it can be updated incrementally, AdjacencyMatrix
// also implements the [Dynamic] interface.
type AdjacencyMatrix struct {
    mat matrix2.Interface[int]
}

// TODO implement Dynamic
// TODO implement transpose, undirected

// NewAdjacencyMatrix returns an AdjacencyMatrix backed by the specified
// matrix implementation. If nil, defaults to [matrix.NewBit].
//
// A matrix implementation with a wider data type is needed to implement an
// AdjacencyMatrix for a multigraph e.g. [matrix.NewGrid].
func NewAdjacencyMatrix(m matrix2.Constructor[int]) AdjacencyMatrix {
    if m == nil { m = matrix2.NewBit }
    return AdjacencyMatrix{
        mat: matrix2.New2D(m, 0, 0),
    }
}

// Matrix returns a pointer to the underlying [matrix.Interface] (of type int).
func (m AdjacencyMatrix) Matrix() matrix2.Interface[int] {
    return m.mat
}

// Get returns the number of directed edges from a source vertex to a target
// vertex (including self-loops, if any).
func (m AdjacencyMatrix) Get(source, target VertexIndex) int {
    if int(source) > m.mat.Width() { return 0 }
    return m.mat.Get2D(int(source), int(target))
}

// Set stores the number of directed edges from a source vertex to a target
// vertex (including self-loops, if any).
func (m AdjacencyMatrix) Set(source, target VertexIndex, count int) {
    m.mat.Set2D(count, int(source), int(target))
}

// Width returns the width of the adjacency matrix. Note that if VertexIndexes
// are sparsely distributed, width may be greater the number of vertexes
// produced by iteration.
func (m AdjacencyMatrix) Width() int {
    return m.mat.Width()
}

// CountEdges returns the total number of edges in the adjacency matrix.
func (m AdjacencyMatrix) CountEdges() int {
    return matrix2.Reduce(m.mat, 0, func(a, b int) int { return a + b })
}

func (m AdjacencyMatrix) Clear() {
    m.mat.Clear()
}

// Resize updates the adjacency matrix, if necessary, so that it has at least
// capacity for width elements in each dimension. It reuses underlying memory
// from the existing matrix where possible. Note that this will clear the
// matrix.
func (m AdjacencyMatrix) Resize(width int) {
    m.mat.Resize2D(width, width)
}

// Indegree returns the number of directed edges from any vertex to the
// specific target vertex (including self-loops from the target to itself).
//
// This is computed with O(n) complexity; construct a [DegreeMatrix] for
// constant time.
func (m AdjacencyMatrix) Indegree(target VertexIndex) int {
    var sum = 0
    for i := 0; i < m.mat.Width(); i++ {
        sum += m.mat.Get2D(i, int(target))
    }
    return sum
}

// Outdegree returns the number of directed edges from a specific source vertex
// to any other vertex (including self-loops from the source to itself).
//
// This is computed with O(n) complexity; construct a [DegreeMatrix] for
// constant time.
func (m AdjacencyMatrix) Outdegree(source VertexIndex) int {
    var sum = 0
    for i := 0; i < m.mat.Width(); i++ {
        sum += m.mat.Get2D(int(source), i)
    }
    return sum
}

// Degree returns the number of edges either to or from a specific vertex,
// including self-loops which are counted (by definition) as two edges.
//
// This is computed with O(n) complexity; construct a [DegreeMatrix] for
// constant time.
func (m AdjacencyMatrix) Degree(source VertexIndex) int {
    return m.Indegree(source) + m.Outdegree(source)
}

// Vertexes implements the graph [Iterator] Vertexes method. Every vertex will
// have at least one edge (in either direction).
func (m AdjacencyMatrix) Vertexes() func() (VertexIndex, bool) {
    i, width := 0, m.mat.Width()
    return func() (_ VertexIndex, _ bool) {
        for {
            if i >= width { return }
            idx := VertexIndex(i)
            i++
            if m.Degree(idx) == 0 { continue }
            return idx, true
        }
    }
}

// Edges implements the graph [Iterator] Edges method.
func (m AdjacencyMatrix) Edges(source VertexIndex) func() (VertexIndex, int, bool) {
    i, width := 0, m.mat.Width()
    return func() (_ VertexIndex, _ int, _ bool) {
        for {
            if i >= width { return }
            target := VertexIndex(i)
            edges := m.Get(source, target)
            i++
            if edges == 0 { continue }
            return target, edges, true
        }
    }
}

// Weight implements the graph [Iterator] Weight method. The weight of each
// edge in an adjacency matrix is defined as the number of edges from source to
// target in the input matrix (including self-loops, if any). If the graph
// is not a multigraph, this is always exactly one for any existing edge in the
// matrix.
func (m AdjacencyMatrix) Weight(source, target VertexIndex) Weight {
    return Weight(m.Get(source, target))
}

// Calculate computes the adjacency matrix for a finite graph g. The created
// adjacency matrix is itself a graph implementing [Iterator] and [Dynamic],
// containing only the vertexes of g with at least one inward or outward edge.
//
// Each vertex index in the adjacency matrix corresponds to a matching index
// in graph g. Once an adjacency matrix has been constructed, it is not
// affected by future changes to graph g.
func (m AdjacencyMatrix) Calculate(g Iterator) {
    width := int(vertexIndexLimit(g.Vertexes))
    m.Resize(width)
    m.Clear()

    vertexIter := g.Vertexes()
    for {
        vertexIdx, vertexOk := vertexIter()
        if !vertexOk { break }

        edgeIter := g.Edges(vertexIdx)
        for {
            targetIdx, edges, edgeOk := edgeIter()
            if !edgeOk { break }
            if edges < 1 { continue }

            current := m.Get(vertexIdx, targetIdx)
            m.Set(vertexIdx, targetIdx, current + edges)
        }
    }
}
