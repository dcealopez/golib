// Package graph implements general-purpose graph algorithms.
//
// This includes some "online algorithms" for dynamic graphs i.e. algorithms
// that can efficiently give new results as a graph changes.
//
// This package does not implement any specific graph data structures: bring
// your own implementation!
//
// In general, this package is written & composed in such a way as to reduce
// runtime memory (re)allocations by encouraging the caller to reuse buffers.
//
// References
//
// Many of the algorithms and definitions are thanks to CLRS "Introduction to
// Algorithms", 3rd edition.
//
// The slides from Danupon Nanongkai (KTH, Sweden)'s ADFOCS 2018 talk,
// "Introduction to Dynamic Graph Algorithms", were also helpful.
//
// The dynamic BFS algorithm is from "Semi-dynamic breadth-first search in
// digraphs", Franciosa, Frigioni & Giaccio, Theoretical Computer Science, Vol.
// 250, Issue 1-2, Jan 6 2001, pp 201–217.
//
// To do
//
//
package graph

// Weight represents the "cost" of a weighted edge from one vertex to another.
// Weights may be negative, except where indicated by certain algorithms.
//
// Real number (floating point) weights should be scaled to an integer range,
// for example by multiplying by some constant and rounding.
type Weight int

// Dev note: While it was tempting to make the Weight type generic, a
// word-sized int is enough. Real numbers face rounding even when converted to
// IEEE 754 floating point representation, so can also survive rounding to an
// integer representation. Similarly, if a range of integer weights exceeds
// 2^32, or a floating point representation needs a full 64 bits of precision,
// then the problem space is probably large enough that the difference between
// two almost-identical weighted paths doesn't matter.

// VertexIndex is a non-negative index that uniquely identifies each vertex
// in a graph. Vertex indexes do not have to be consecutive or sorted, may be
// sparse, and do not have to start at zero, but it may be more efficient where
// that is the case.
//
// Many algorithms in this package will have computational complexity
// proportionate to the maximum VertexIndex, and require memory proportionate
// to the maximum VertexIndex squared.
//
// The optimum assignment of a VertexIndex to each vertex is called the Minimum
// Linear Arrangement, but in the general case this is NP-Hard and even an
// approximation has quadratic complexity. You can get decent results just
// using the indexes of a generational array (see the `genarray` sibling
// package).
type VertexIndex int

// Iterator is the basic interface that a graph data type must implement to be
// able to use many of the algorithms in this package.
//
// A graph can have many equivalent implementations: a data structure, perhaps
// an adjacency list or an adjacency matrix; or an algorithm. The mapping
// between this general-purpose graph interface and a given graph
// implementation is achieved through two iterator-style functions.
//
// Graphs may contain cycles, loops, be disjoint, be multigraphs, etc.
// without restriction except where noted. In general, algorithms in this
// package operate on finite graphs only.
//
// Edges are directed, but an undirected graph can always be implemented
// efficiently from a directed graph - e.g. from its adjacency matrix.
//
// The Weight method calculates a weight of the edge between source and target.
// In the case of a multigraph, multiple edges must be reduced by some
// appropriate method e.g. by picking the minimum edge weight, by calculating
// the sum of edge weights, or by the count of edges, etc. Algorithms will only
// call Weight if a directed edge already exists from the source to the target
// vertex, so it is not necessary to validate the arguments again. If the graph
// is unweighted, you can return a unit Weight of 1.
type Iterator interface {
    Vertexes() VertexIterator
    Edges(source VertexIndex) EdgeIterator
    Weight(source, target VertexIndex) Weight
}

// Incremental is an interface for any dynamic graph implementation or online
// algorithm that can efficiently update to give a useful result if a graph
// changes by the addition of a vertex or edge, or if an edge weight decreases.
//
// Multiple related graph representations can be kept in-sync with
// [TeeIncremental].
type Incremental interface {
    AddVertex(VertexIndex)
    AddEdge(source VertexIndex, target VertexIndex)
    DecreaseWeight(source VertexIndex, weight Weight)
}

// Decremental is an interface for any dynamic graph implementation or online
// algorithm that can efficiently update to give a useful result if a graph
// changes by the removal of a vertex or edge, or if an edge weight increases.
//
// Multiple related graph representations can be kept in-sync with
// [TeeDecremental].
type Decremental interface {
    RemoveVertex(VertexIndex)
    RemoveEdge(source VertexIndex, target VertexIndex)
    IncreaseWeight(source VertexIndex, weight Weight)
}

// Dynamic is an interface for any dynamic graph implementation or online
// algorithm that can efficiently update to give a useful result if a graph
// changes.
//
// Multiple related graph representations can be kept in-sync with
// [TeeDynamic].
type Dynamic interface {
    Incremental
    Decremental
}

// WeightFunc is the type of a function that gives a weight between two
// vertexes in a graph. Algorithms will only call this if a directed edge
// already exists from the source to the target vertex, so it is not necessary
// to validate the arguments again.
type WeightFunc func(source, target VertexIndex) Weight

// VertexIterator is the type of a generator function that, for some particular
// graph, generates a VertexIndex for each vertex in the graph.
//
// The last return value controls the iteration - if false, the iteration has
// finished and the other return value is not useful.
type VertexIterator = func() (vertex VertexIndex, ok bool)

// EdgeIterator is the type of a generator function that, for some particular
// source vertex, generates each target vertex connected by at least one
// directed edge, and a count of the number of edges between them. If the graph
// is not a multigraph, this count will always be one.
//
// The last return value controls the iteration - if false, the iteration has
// finished and the other return values are not useful.
type EdgeIterator = func() (target VertexIndex, count int, ok bool)

// vertexIndexLimit returns the size of a slice needed to hold the largest
// VertexIndex produced by a VertexIterator.
func vertexIndexLimit(g func() VertexIterator) VertexIndex {
    it := g()
    highest := VertexIndex(-1)
    for {
        idx, ok := it()
        if !ok { break }
        if idx > highest { highest = idx }
    }
    return highest + 1
}
