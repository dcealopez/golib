package graph

import (
    "github.com/tawesoft/golib/v2/ds/matrix"
)

// DegreeMatrix represents the number of edges on a vertex (to or from any
// other vertex in aggregate). This may be the in-, out-, or undirected-,
// degree.
//
// Once built, a degree matrix can be queried in constant time.
//
// Values are indexed by [VertexIndex]. A DegreeMatrix is a diagonal-matrix;
// values off the diagonal are zero.
type DegreeMatrix struct {
    mat matrix.Interface[int]
}

// NewDegreeMatrix returns a new degree matrix of undefined size.
func NewDegreeMatrix() DegreeMatrix {
    return DegreeMatrix{
        mat: matrix.New2D[int](matrix.NewDiagonal[int], 0, 0),
    }
}

// Matrix returns a pointer to the underlying [matrix.Interface] (of type int).
func (m DegreeMatrix) Matrix() matrix.Interface[int] {
    return m.mat
}

// Get returns the (in-, out-, or undirected) degree of the given vertex.
func (m DegreeMatrix) Get(source VertexIndex) int {
    if int(source) > m.mat.Width() { return 0 }
    return m.mat.Get2D(int(source), int(source))
}

// Set stores the (in-, out-, or undirected) degree of the given vertex.
func (m DegreeMatrix) Set(source VertexIndex, count int) {
    m.mat.Set2D(count, int(source), int(source))
}

// Width returns the width of the degree matrix. Note that if VertexIndexes
// are sparsely distributed, width may be greater the number of vertexes
// produced by iteration.
func (m DegreeMatrix) Width() int {
    return m.mat.Width()
}

// CountEdges returns the total number of edges in the degree matrix.
func (m DegreeMatrix) CountEdges() int {
    return matrix.Reduce(m.mat, 0, func(a, b int) int { return a + b })
}

// Resize updates the degree matrix, if necessary, so that it has at least
// capacity for width elements in each dimension. It reuses underlying memory
// from the existing matrix where possible. Note that this will clear the
// matrix.
func (m DegreeMatrix) Resize(width int) {
    m.mat.Resize2D(width, width)
}

func (m DegreeMatrix) Clear() {
    m.mat.Clear()
}

// Calculate computes the degree matrix from any vertex iterator and any degree
// function.
//
// For example, an in-, out-, or undirected degree matrix may be constructed
// from an [AdjacencyMatrix] using its Vertexes method and its Indegree,
// Outdegree, and Degree methods respectively.
//
// Each vertex index in the degree matrix corresponds to the matching index
// in the input graph g. Once a degree matrix has been constructed, it is not
// affected by future changes to g.
func (m DegreeMatrix) Calculate(g func() VertexIterator, degree func(index VertexIndex) int) {
    width := int(vertexIndexLimit(g))
    m.Resize(width)
    m.Clear()

    vertexIter := g()
    for {
        vertexIdx, vertexOk := vertexIter()
        if !vertexOk { break }

        deg := degree(vertexIdx)
        if deg == 0 { continue }

        m.Set(vertexIdx, deg)
    }
}

// Roots returns an iterator that generates only the indexes of vertexes that
// have no parents i.e. root vertexes in a directed graph.
//
// A degree function can be made by passing an appropriate method from an
// [AdjacencyMatrix] or [DegreeMatrix].
func Roots(
    vertexes func() VertexIterator,
    indegree func(VertexIndex) int,
) VertexIterator {
    x := VertexIndex(0)
    limit := vertexIndexLimit(vertexes)
    return func() (VertexIndex, bool) {
        // for each vertex we haven't looked at yet...
        for i := x; i < limit; i++ {
            if (indegree(i) == 0) {
                x = i + 1 // start next iteration after this point
                return i, true
            }
        }
        return 0, false
    }
}

// Leaves returns an iterator that generates indexes of vertexes that have
// exactly one parent and exactly zero children in a directed graph.
//
// A degree function can be made by passing an appropriate method from an
// [AdjacencyMatrix] or [DegreeMatrix].
func Leaves(
    vertexes func() VertexIterator,
    indegree func(VertexIndex) int,
    outdegree func(index VertexIndex) int,
) VertexIterator {
    x := VertexIndex(0)
    limit := vertexIndexLimit(vertexes)
    return func() (VertexIndex, bool) {
        // for each vertex we haven't looked at yet...
        for i := x; i < limit; i++ {
            if (indegree(i) == 1) && (outdegree(i) == 0) {
                x = i + 1 // start next iteration after this point
                return i, true
            }
        }
        return 0, false
    }
}
