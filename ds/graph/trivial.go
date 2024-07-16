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

// TeeIncremental returns a new Incremental interface whose methods, when
// called, are also called on `a` and `b`. This can be used to keep two
// related graph implementations in sync with each-other.
//
// Multiple related graph representations can be kept in-sync with
// [TeeIncremental].
func TeeIncremental(a, b Incremental) Incremental {
    return teeIncremental{a, b}
}

type teeIncremental struct {
    a, b Incremental
}

func (t teeIncremental) AddVertex(index VertexIndex) {
    t.a.AddVertex(index)
    t.b.AddVertex(index)
}
func (t teeIncremental) AddEdge(source VertexIndex, target VertexIndex) {
    t.a.AddEdge(source, target)
    t.b.AddEdge(source, target)
}
func (t teeIncremental) DecreaseWeight(source VertexIndex, weight Weight) {
    if weight < 0 { panic("DecreaseWeight: weight must be positive") }
    t.a.DecreaseWeight(source, weight)
    t.b.DecreaseWeight(source, weight)
}

// TeeDecremental returns a new Decremental interface whose methods, when
// called, are also called on `a` and `b`. This can be used to keep two
// related graph implementations in sync with each-other.
func TeeDecremental(a, b Decremental) Decremental {
    return teeDecremental{a, b}
}

type teeDecremental struct {
    a, b Decremental
}

func (t teeDecremental) RemoveVertex(index VertexIndex) {
    t.a.RemoveVertex(index)
    t.b.RemoveVertex(index)
}
func (t teeDecremental) RemoveEdge(source VertexIndex, target VertexIndex) {
    t.a.RemoveEdge(source, target)
    t.b.RemoveEdge(source, target)
}
func (t teeDecremental) IncreaseWeight(source VertexIndex, weight Weight) {
    if weight < 0 { panic("IncreaseWeight: weight must be positive") }
    t.a.IncreaseWeight(source, weight)
    t.b.IncreaseWeight(source, weight)
}

// TeeDynamic returns a new Dynamic interface whose methods, when called, are
// also called on `a` and `b`. This can be used to keep two related graph
// implementations in-sync with each-other.
func TeeDynamic[W Weight] (a, b Dynamic) Dynamic {
    return teeDynamic[W]{
        teeIncremental: teeIncremental{a, b},
        teeDecremental: teeDecremental{a, b},
    }
}

type teeDynamic[W Weight] struct {
    teeIncremental
    teeDecremental
}
