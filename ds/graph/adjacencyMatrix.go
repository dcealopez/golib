package graph

import (
    "github.com/tawesoft/golib/v2/ds/matrix"
    "github.com/tawesoft/golib/v2/math/integer"
)

// AdjacencyMatrix represents the number of directed edges from a source vertex
// (along the x-axis) to a target vertex (along the y-axis), including
// self-loops, if any.
//
// As a representation of a graph itself, AdjacencyMatrix implements the graph
// [Iterator] interface. As it can be updated incrementally, AdjacencyMatrix
// also implements the [Dynamic] interface.
type AdjacencyMatrix struct {
    mat matrix.M[int]
    multi bool
}

// TODO implement Dynamic
// TODO implement transpose, undirected

// NewAdjacencyMatrix returns an AdjacencyMatrix. Each vertex pair can only
// have one edge. For a multigraph, use [NewMultiAdjacencyMatrix].
func NewAdjacencyMatrix() AdjacencyMatrix {
    return AdjacencyMatrix{
        mat:   matrix.NewBit(4, 4),
        multi: false,
    }
}

// NewAdjacencyMatrix returns an AdjacencyMatrix. Each vertex pair can have
// a count of multiple edges. For a simple graph, use [NewAdjacencyMatrix].
func NewMultiAdjacencyMatrix() AdjacencyMatrix {
    return AdjacencyMatrix{
        mat:   matrix.NewGrid[int](4, 4),
        multi: true,
    }
}

// Matrix returns a pointer to the underlying [matrix.M] (of type int).
func (m AdjacencyMatrix) Matrix() matrix.M[int] {
    return m.mat
}

// Get returns the number of directed edges from a source vertex to a target
// vertex (including self-loops, if any).
func (m AdjacencyMatrix) Get(source, target VertexIndex) int {
    if int(max(source, target)) > m.Width() { return 0 }
    idx := m.mat.Index(int(source), int(target))
    return m.mat.Get(idx)
}

// Set stores the number of directed edges from a source vertex to a target
// vertex (including self-loops, if any).
func (m AdjacencyMatrix) Set(source, target VertexIndex, count int) {
    largest := max(source, target)
    m.Resize(int(largest))
    idx := m.mat.Index(int(source), int(target))
    m.mat.Set(idx, count)
}

// Width returns the width of the adjacency matrix. Note that if VertexIndexes
// are sparsely distributed, width may be greater the number of vertexes
// produced by iteration.
func (m AdjacencyMatrix) Width() int {
    return m.mat.Length(0)
}

// CountEdges returns the total number of edges in the adjacency matrix.
func (m AdjacencyMatrix) CountEdges() int {
    sum := 0

    // optimisation: walk the matrix sparsely
    for idx, ok := -1, true; ok; idx, ok = m.mat.Next(idx) {
        if idx < 0 { continue }
        sum += m.mat.Get(idx)
    }

    return sum
}

func (m AdjacencyMatrix) Clear() {
    m.mat.Clear()
}

// Resize updates the adjacency matrix, if necessary, so that it has at least
// capacity for width elements in each dimension.
func (m AdjacencyMatrix) Resize(width int) {
    if m.Width() <= width { return }
    width = int(integer.AlignPowTwo(uint(width)))

    // TODO move this into matrix itself

    var mcons matrix.M[int]
    if m.multi {
        mcons = matrix.NewGrid[int](width, width)
    } else {
        mcons = matrix.NewBit[int](width, width)
    }
    dest := mcons
    matrix.Copy(dest, m.mat)
    m.mat = dest
}

// Indegree returns the number of directed edges from any vertex to the
// specific target vertex (including self-loops from the target to itself).
//
// This is computed with O(n) complexity; construct a [DegreeMatrix] for
// constant time.
func (m AdjacencyMatrix) Indegree(target VertexIndex) int {
    var sum = 0
    for i := 0; i < m.Width(); i++ {
        idx := m.mat.Index(i, int(target))
        sum += m.mat.Get(idx)
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
    for i := 0; i < m.Width(); i++ {
        idx := m.mat.Index(int(source), i)
        sum += m.mat.Get(idx)
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
    i, width := 0, m.Width()
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
    i, width := 0, m.Width()
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
