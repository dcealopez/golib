// Package graph implements general-purpose graph algorithms.
//
// This package does not implement any specific graph data structures: bring
// your own implementation!
//
// Many of the algorithms and definitions are thanks to CLRS "Introduction to
// Algorithms", 3rd edition.
package graph

import (
    "github.com/tawesoft/golib/v2/iter"
    "golang.org/x/exp/constraints"
)

// Number represents any integer or float type.
type Number interface { constraints.Integer | constraints.Float }

// VertexIndex is a non-negative index that uniquely identifies each vertex
// in a graph. Vertex indexes do not have to be consecutive or sorted, may be
// sparse, and do not have to start at zero, but it may be more efficient where
// that is the case.
type VertexIndex int

// EdgeIndex optionally uniquely identifies a directed edge from a source
// vertex to a target vertex, for a unique (source, target) vertex pair in a
// multigraph. Edge indexes do not have to be consecutive or sorted, may be
// sparse, and do not have to start at zero, but it may be more efficient where
// that is the case.
//
// An EdgeIndex is only particularly useful in a multigraph. In normal graphs,
// each edge from a source can always be uniquely identified by the target.
// In this case, EdgeIndex is equal to the VertexIndex of the target.
//
// In some graph representations, unique information about each edge is lost.
// By convention, use -1 to represent a missing index.
//
// Note that while edges are always directed, an undirected graph can be
// implemented efficiently by representing an undirected edge as a directed
// edge from a lower-indexed vertex to a higher-indexed vertex, or vice-versa.
type EdgeIndex int

// Iterator is the basic interface that a graph data type must implement to be
// able to use many of the algorithms in this package.
//
// The mapping between this general-purpose graph interface and a given graph
// implementation is achieved through iterator-style functions that generate a
// [VertexIndex]. The Edges iterator function generates a count of the number
// of edges between the two vertexes. If the graph is not a multigraph, this
// count will always be one. Each iterator-style function returns false as the
// last argument iff they have completed.
//
// A graph can have many equivalent implementations: a data structure, perhaps
// an adjacency list or an adjacency matrix; or an algorithm, for example an
// algorithm that generates an infinite line of vertexes with a directed edge
// from each vertex to its successor.
//
// Graphs may contain cycles, loops, be disjoint, be multigraphs, be infinite,
// etc. without restriction except where noted. These properties can be tested
// for by algorithms implemented in this package.
type Iterator interface {
    Vertexes() VertexIterator
    Edges(source VertexIndex) EdgeIterator
}

type VertexIterator = func() (vertex VertexIndex, ok bool)
type EdgeIterator = func() (target VertexIndex, count int, ok bool)

// Builder is a basic interface for building a graph.
/*
type Builder[T any] interface {
    Iterator
    AddVertex() VertexIndex
    AddEdge(source VertexIndex, target VertexIndex)
}
TODO
func TeeBuilder[T any](builders ... Builder[T]) { ... }
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

// FilterVertexes implements the graph [Iterator] interface and represents a
// subgraph of a parent graph of only vertexes that satisfy the given filter
// function.
//
// The Iterator implemented by FilterVertexes is only a view of the parent graph
// and is computed on-the-fly as the parent changes.
type FilterVertexes struct {
    Parent Iterator
    Filter func(vertex VertexIndex) bool
}

func (f FilterVertexes) Vertexes() func() (VertexIndex, bool) {
    if f.Filter == nil { return f.Parent.Vertexes() }
    return iter.Filter(f.Filter, f.Parent.Vertexes())
}

func (f FilterVertexes) Edges(source VertexIndex) func() (VertexIndex, int, bool) {
    return f.Parent.Edges(source)
}

// FilterEdges implements the graph [Iterator] interface and represents a
// subgraph of a parent graph of only edges that satisfy the given filter
// function.
//
// The Iterator implemented by FilterEdges is only a view of the parent graph and
// is computed on-the-fly as the parent changes.
type FilterEdges struct {
    Parent Iterator
    Filter func(source VertexIndex, target VertexIndex) bool
}

func (f FilterEdges) Vertexes() func() (VertexIndex, bool) {
    return f.Parent.Vertexes()
}

func (f FilterEdges) Edges(source VertexIndex) func() (VertexIndex, int, bool) {
    if f.Filter == nil { return f.Parent.Edges(source) }
    it := f.Parent.Edges(source)
    return func() (_ VertexIndex, _ int, _ bool) {
        target, count, ok := it()
        if !ok || !f.Filter(source, target) { return }
        return target, count, true
    }
}
