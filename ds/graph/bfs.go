package graph

import (
    "math"

    "github.com/tawesoft/golib/v2/ks"
)

type vertexBFS struct {
    // predecessor is the vertex immediately previous to this one along one
    // possible shortest path. Unreachable and root verticies have this
    // set to -1.
    predecessor VertexIndex

    // discovered is the "search colour" in the CLRS algorithm; true is "gray"
    // or "black".
    discovered bool

    // distance is the sum of weights travelled from the root. In an unweighted
    // graph, or a graph where every edge has unit weight, this is the same
    // cumulative number of edges travelled from source to the destination
    // along the shortest path.
    distance Weight
}

// BfsTree represents a (unweighted) breadth-first tree of the reachable
// graph from a given start vertex, taking the shortest number of edges.
//
// A BfsTree is itself a graph (the predecessor subgraph of the graph it
// was constructed from), and implements the [Iterator] interface.
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
    maximum := Weight(math.MaxInt)
    t.start = 0
    clear(t.vertexes)
    clear(t.queue)
    t.queue = t.queue[0:0]
    for i := 0; i < len(t.vertexes); i++ {
        t.vertexes[i].predecessor = -1
        t.vertexes[i].distance = maximum
        t.vertexes[i].discovered = false
    }
}

// Reachable returns true if the given vertex is reachable from the root of
// the BfsTree.
func (t BfsTree) Reachable(vertex VertexIndex) bool {
    if (vertex < 0) || (int(vertex) >= len(t.vertexes)) { return false }
    // Every reachable vertexBFS has a predecessor, except the root.
    return (t.vertexes[vertex].predecessor >= 0) || (vertex == t.start)
    // alternatively, every reachable vertexBFS is discovered.
}

// Predecessor returns the predecessor of the vertex in the search tree.
// If the vertex is not reachable, or if the vertex is the search start
// vertex, the boolean return value is false.
func (t BfsTree) Predecessor(vertex VertexIndex) (VertexIndex, bool) {
    if !t.Reachable(vertex) { return 0, false }
    predecessor := t.vertexes[vertex].predecessor
    if predecessor < 0 { return 0, false }
    return predecessor, true
}

// Distance returns the cumulative number of edges crossed from the search
// start vertex to the given target. If the vertex is not reachable, the
// boolean return value is false.
func (t BfsTree) Distance(vertex VertexIndex) (Weight, bool) {
    if !t.Reachable(vertex) { return 0, false }
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

// Edges implements the graph [Iterator] Edges method.
//
// Each vertex has exactly one directed edge, to its predecessor, except the
// root vertex, which has none.
func (t BfsTree) Edges(source VertexIndex) EdgeIterator {
    return func() (_ VertexIndex, _ int, _ bool) {
        target, ok := t.Predecessor(source)
        if (!ok) { return }
        return target, 1, true
    }
}

// CalculateUnweighted computes an unweighted breadth-first tree of the
// reachable graph from a given start vertex, taking the shortest number of
// edges. This is equivalent to a weighted search where every edge has a unit
// weight of one. The resulting search tree has useful properties.
//
// The search stores a result in the provided result object, resizing its
// underlying buffer if necessary.
//
// There may be many possible breadth-first trees of an input graph. This
// procedure always visits vertexes in the order given by a graph's
// [EdgeIterator].
func (t *BfsTree) CalculateUnweighted(graph Iterator, start VertexIndex) {
    t.Resize(int(vertexIndexLimit(graph.Vertexes)))
    t.Clear()

    t.start = start
    t.vertexes[start].discovered = true
    t.vertexes[start].distance = 0
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

// CalculateWeightedGeneral computes a weighted breadth-first tree of the
// reachable graph from a given start vertex, taking the shortest distance
// calculated as a cumulative sum of weights along a path. The resulting search
// tree has useful properties.
//
// This search is performed in the "general case" where edges may have negative
// weights, but not negative-weight cycles. [CalculateWeighted] is more
// efficient, but does not support negative edge weights.
//
// The search stores a result in the provided result object, resizing its
// underlying buffer if necessary.
//
// There may be many possible breadth-first trees of an input graph. The order
// taken by this procedure, when two paths have the same weight, is arbitrary
// and subject to change.
func (t *BfsTree) CalculateWeightedGeneral(
    graph Iterator,
    start VertexIndex,
    weight WeightFunc,
) {
    // Bellman-Ford algorithm

    limit := int(vertexIndexLimit(graph.Vertexes))
    maxDistance := Weight(math.MaxInt)
    t.Resize(limit)
    t.Clear()

    t.start = start
    t.vertexes[start].discovered = true
    t.vertexes[start].distance = 0
    t.queue = append(t.queue, t.start)

    // repeat limit-1 times
    for i := 0; i < limit - 1; i++ {
        uIter := graph.Vertexes()
        for {
            u, ok := uIter()
            if !ok { break }

            vIter := graph.Edges(u)
            for {
                v, _, ok := vIter()
                if !ok { break }

                w := weight(u, v)
                if t.vertexes[u].distance >= maxDistance { continue }
                if (t.vertexes[u].distance + w) < t.vertexes[v].distance {
                    t.vertexes[v].distance = t.vertexes[u].distance + w
                    t.vertexes[v].predecessor = u
                }
            }
        }
    }

    // TODO check negative cycles?
}

// TODO CalculateWeighted
// (Dijkstra's algorithm)
