package graph

import (
    "github.com/tawesoft/golib/v2/iter"
)

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

func (f FilterVertexes) Vertexes() VertexIterator {
    if f.Filter == nil { return f.Parent.Vertexes() }
    return iter.Filter(f.Filter, f.Parent.Vertexes())
}

func (f FilterVertexes) Edges(source VertexIndex) EdgeIterator {
    return f.Parent.Edges(source)
}

// FilterEdges implements the graph [Iterator] interface and represents a
// subgraph of a parent graph of only edges that satisfy the given filter
// function.
//
// The Iterator implemented by FilterEdges is only a view of the parent graph
// and is computed on-the-fly as the parent changes.
type FilterEdges struct {
    Parent Iterator
    Filter func(source VertexIndex, target VertexIndex) bool
}

func (f FilterEdges) Vertexes() VertexIterator {
    return f.Parent.Vertexes()
}

func (f FilterEdges) Edges(source VertexIndex) EdgeIterator {
    if f.Filter == nil { return f.Parent.Edges(source) }
    it := f.Parent.Edges(source)
    return func() (_ VertexIndex, _ int, _ bool) {
        target, count, ok := it()
        if !ok || !f.Filter(source, target) { return }
        return target, count, true
    }
}

// WeightedEdges implements [WeightedIterator] interface by attaching the
// provided WeightFunc to an existing [Iterator].
//
// A simple WeightFunc for unit-weights can be created using the UnitWeightFunc
// method on an [AdjacencyMatrix].
//
// The WeightedIterator implemented by WeightedEdges is only a view of the
// parent graph and is computed on-the-fly as the parent changes.
type WeightedEdges[Weight Number] struct {
    Parent Iterator
    WeightFunc func(source, target VertexIndex) (weight Weight, ok bool)
}

func (v WeightedEdges[W]) Vertexes() VertexIterator {
    return v.Parent.Vertexes()
}

func (v WeightedEdges[W]) Edges(source VertexIndex) EdgeIterator {
    return v.Parent.Edges(source)
}

func (v WeightedEdges[W]) Weight(source, target VertexIndex) (weight W, ok bool) {
    return v.WeightFunc(source, target)
}
