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
