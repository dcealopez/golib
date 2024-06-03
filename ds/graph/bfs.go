package graph

import (
    "github.com/tawesoft/golib/v2/ks"
)

type vertexBFS struct{
    predecessor VertexIndex // -1 if none
    discovered bool         // "search colour" in the CLRS algorithm; true is "gray"
    distance int            // cumulative number of edges crossed from root
}

// BfsTree represents a (unweighted) breadth-first tree of the reachable
// graph from a given start vertex, taking the shortest number of edges.
//
// A BfsTree is itself a graph, and implements the [Iterator] interface.
type BfsTree struct {
    vertexes []vertexBFS
    start    VertexIndex
    queue    []VertexIndex
}

// NewBfsTree returns a new (empty) breadth-first search tree object for
// storing results.
func NewBfsTree() *BfsTree {
    return &BfsTree{
        vertexes: make([]vertexBFS, 0),
        queue:    make([]VertexIndex, 0),
    }
}

// Resize updates the BfsTree, if necessary, so that it has at least capacity
// for n vertexes. It reuses underlying memory from the existing tree where
// possible. Note that this will clear the tree.
func (t *BfsTree) Resize(n int) {
    t.vertexes = ks.SetLength(t.vertexes, n)
}

func (t *BfsTree) Clear() {
    t.start = 0
    clear(t.vertexes)
    clear(t.queue)
    t.queue = t.queue[0:0]
    for i := 0; i < len(t.vertexes); i++ {
        t.vertexes[i].predecessor = -1
    }
}

// Reachable returns true if the given vertex is reachable from the root of
// the BfsTree.
func (t BfsTree) Reachable(vertex VertexIndex) bool {
    if (vertex < 0) || (int(vertex) >= len(t.vertexes)) { return false }
    // Every reachable vertexBFS has a predecessor, except the root.
    return (t.vertexes[vertex].predecessor >= 0) || (vertex == t.start)
}

// Predecessor returns the predecessor of the vertex in the search tree.
// If the vertex is not reachable, or if the vertex is the search start
// vertex, the boolean return value is false.
func (t BfsTree) Predecessor(vertex VertexIndex) (VertexIndex, bool) {
    if (vertex < 0) || (int(vertex) >= len(t.vertexes)) { return 0, false }
    predecessor := t.vertexes[vertex].predecessor
    if predecessor < 0 { return 0, false }
    return predecessor, true
}

// Distance returns the cumulative number of edges crossed from the search
// start vertex to the given target. If the vertex is not reachable, the
// boolean return value is false.
func (t BfsTree) Distance(vertex VertexIndex) (int, bool) {
    if (vertex < 0) || (int(vertex) >= len(t.vertexes)) { return 0, false }
    return t.vertexes[vertex].distance, true
}

// Vertexes implements the graph [Iterator] Vertexes method.
func (t BfsTree) Vertexes() VertexIterator {
    i, width := 0, len(t.vertexes)
    return func() (_ VertexIndex, _ bool) {
        for {
            if i >= width { return }
            idx := VertexIndex(i)
            i++
            if !t.Reachable(idx) { continue }
            return idx, true
        }
    }
}

// Calculate computes a (unweighted) breadth-first tree of the reachable graph
// from a given start vertex, taking the shortest number of edges. The
// resulting search tree gives useful properties.
//
// The search stores a result in the provided result object, resizing its
// underlying buffer if necessary.
func (t *BfsTree) Calculate(graph Iterator, start VertexIndex) {
    t.Resize(int(vertexIndexLimit(graph.Vertexes)))
    t.Clear()

    t.start = start
    t.vertexes[start].discovered = true
    t.queue = append(t.queue, t.start)

    for len(t.queue) != 0 {
        // dequeue
        source := t.queue[len(t.queue)-1]
        t.queue = t.queue[:len(t.queue)-1]
        u := t.vertexes[source]

        edgesIter := graph.Edges(source)
        for {
            target, _, ok := edgesIter()
            if !ok { break }
            v := &(t.vertexes[target])

            if v.discovered { continue }
            v.discovered  = true
            v.distance    = u.distance + 1
            v.predecessor = source

            t.queue = append(t.queue, target)
        }
    }
}
