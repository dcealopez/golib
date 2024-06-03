// Package graph implements general-purpose graph algorithms.
//
// This package does not implement any specific graph data structures: bring
// your own implementation!
//
// Many of the algorithms and definitions are thanks to CLRS "Introduction to
// Algorithms", 3rd edition.
//
// In general, this package is written & composed in such a way as to reduce
// runtime memory (re)allocations by encouraging the caller to reuse buffers.
package graph

import (
    "golang.org/x/exp/constraints"
)

// Number represents any integer or float type. This is used to define weights.
type Number interface { constraints.Integer | constraints.Float }

// VertexIndex is a non-negative index that uniquely identifies each vertex
// in a graph. Vertex indexes do not have to be consecutive or sorted, may be
// sparse, and do not have to start at zero, but it may be more efficient where
// that is the case.
type VertexIndex int

// Iterator is the basic interface that a graph data type must implement to be
// able to use many of the algorithms in this package.
//
// A graph can have many equivalent implementations: a data structure, perhaps
// an adjacency list or an adjacency matrix; or an algorithm. The mapping
// between this general-purpose graph interface and a given graph
// implementation is achieved through two iterator-style functions.
//
// Graphs may contain cycles, loops, be disjoint, be multigraphs, be infinite,
// etc. without restriction except where noted. These properties can be tested
// for by algorithms implemented in this package.
//
// Edges are directed, but an undirected graph can always be implemented
// efficiently by defining an "undirected edge" as a directed edge from a
// lower VertexIndex to a higher VertexIndex.
type Iterator interface {
    Vertexes() VertexIterator
    Edges(source VertexIndex) EdgeIterator
}

type WeightedIterator[Weight Number] interface {
    Iterator

    // Weight returns the weight of the edge(s) between source and target.
    // If the boolean return value is false, the weight is infinite.
    // In the case of a multigraph, multiple edges must be reduced into
    // a single value e.g. by picking the smallest. Weights may be negative,
    // except where indicated by certain algorithms.
    Weight(source, target VertexIndex) (weight Weight, ok bool)
}

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

// Adder is a basic interface for adding to a graph or matrix. With [TeeAdder],
// this can be used to keep generated graphs and/or matrices, like an
// [AdjacencyMatrix], in sync with the graph they are generated from.
/*
type Adder[T any] interface {
    AddVertex() VertexIndex
    AddEdge(source VertexIndex, target VertexIndex)
}

// Remover is a basic interface for removing from a graph or matrix. With
// [TeeRemover], this can be used to keep generated graphs and/or matrices,
// like an [AdjacencyMatrix], in sync with the graph they are generated from.
type Remover[T any] interface {
    // Remove all edges first!
    RemoveVertex(target VertexIndex)
    RemoveEdge(source VertexIndex, target VertexIndex)
}
TODO
func TeeAdder[T any](adders ... Adder[T]) { ... }
func TeeRemover[T any](removers ... Remover[T]) { ... }
*/

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
