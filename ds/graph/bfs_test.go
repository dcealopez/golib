package graph_test

import (
    "testing"

    "github.com/tawesoft/golib/v2/ds/graph"
    "github.com/tawesoft/golib/v2/iter"
)

type TestGraph struct {
    // mapping maps Source Vertices to a Set of Target Vertices with a
    // given weight.
    mapping map[graph.VertexIndex](map[graph.VertexIndex]graph.Weight)
    // values maps vertexes to a string
    values map[graph.VertexIndex]string
    count int
}

func NewTestGraph() *TestGraph {
    return &TestGraph{
        mapping: make(map[graph.VertexIndex](map[graph.VertexIndex]graph.Weight)),
        values:  make(map[graph.VertexIndex]string),
        count:   0,
    }
}

func (g *TestGraph) Vertexes() graph.VertexIterator {
    return iter.Keys(iter.FromMap(g.mapping))
}

func (g *TestGraph) Edges(source graph.VertexIndex) graph.EdgeIterator {
    edges, edgesOk := g.mapping[source]
    keys := iter.Keys(iter.FromMap(edges))
    return func() (_ graph.VertexIndex, _ int, _ bool) {
        if !edgesOk { return }
        edge, ok := keys()
        if !ok { return }
        return edge, 1, true
    }
}

func (g *TestGraph) Vertex(s string) graph.VertexIndex {
    value := graph.VertexIndex(g.count)
    g.count++
    g.values[value]  = s
    g.mapping[value] = make(map[graph.VertexIndex]graph.Weight)
    return value
}

func (g *TestGraph) Edge(from, to graph.VertexIndex, weight graph.Weight) {
    g.mapping[from][to] = weight
}

func (g *TestGraph) Weight(from, to graph.VertexIndex) graph.Weight {
    return g.mapping[from][to]
}

func TestBfsTree_CalculateUnweighted(t *testing.T) {
    g := NewTestGraph()

    // graph abcde
    //
    //      /--------------|
    //      |   /-\        |
    //      V   \ V        |
    // a -> b -> c -> d -> e
    //      |              ^
    //      \-------------/

    a := g.Vertex("a") // 0
    b := g.Vertex("b") // 1
    c := g.Vertex("c") // 2
    d := g.Vertex("d") // 3
    e := g.Vertex("e") // 4

    // disconnected subgraph xyz
    x := g.Vertex("x") // 5
    y := g.Vertex("y") // 6
    z := g.Vertex("z") // 7

    g.Edge(a, b, 1) // a->b
    g.Edge(b, c, 2) // b->c
    g.Edge(b, e, 10) // b->e
    g.Edge(c, c, 1) // c->c (self-loo)
    g.Edge(c, d, -1) // c->d -- negative weight allowed because it's not a cycle
    g.Edge(d, e, 5) // d->e
    g.Edge(e, b, 1) // e->b (cycle)

    g.Edge(x, y, 1) // x->y
    g.Edge(y, z, 1) // y->z
    g.Edge(z, x, 1) // z->x (cycle)

    bfst := graph.NewBfsTree()

    // BFS from start vertex b
    bfst.CalculateUnweighted(g, b)

    type row struct{
        vertex      graph.VertexIndex
        predecessor graph.VertexIndex
        distance    graph.Weight
        reachable   bool
    }
    stat := func(v graph.VertexIndex) row {
        var actual row
        actual.vertex = v
        predecessor, hasPredecessor := bfst.Predecessor(v)
        if hasPredecessor {
            actual.predecessor = predecessor
        } else {
            actual.predecessor = -1
        }
        actual.distance, actual.reachable = bfst.Distance(v)
        return actual
    }
    errorRow := func(msg string, expected, actual row) {
        t.Errorf("%s: expected %+v, got %+v", msg, expected, actual)
    }
    test := func(rows []row) {
        for _, expected := range rows {
            actual := stat(expected.vertex)
            if actual.predecessor != expected.predecessor {
                errorRow("wrong predecessor", expected, actual)
            }
            if expected.reachable != actual.reachable {
                errorRow("wrong reachable", expected, actual)
            }
            if expected.reachable && (actual.distance != expected.distance ) {
                errorRow("wrong distance", expected, actual)
            }
        }
    }

    test([]row{
        {a, -1, 0, false},
        {b, -1, 0, true},
        {c,  b, 1, true}, // b->c
        {d,  c, 2, true}, // b->c->d
        {e,  b, 1, true},  // b->e
        {x, -1, 0, false},
        {y, -1, 0, false},
        {z, -1, 0, false},
    })

    // BFS (weighted) from start vertex b
    // negative weights are allowed, but negative weight cycles are not
    bfst.CalculateWeightedGeneral(g, b, g.Weight)

    test([]row{
        {a, -1, 0, false},
        {b, -1, 0, true},
        {c,  b, 2, true}, // b->c
        {d,  c, 1, true}, // b->c->d (2 + -1)
        {e,  d, 6, true},  // b->c->d->e (2 + -1 + 5)
        {x, -1, 0, false},
        {y, -1, 0, false},
        {z, -1, 0, false},
    })

    // TODO give the self-loop c->c a negative weight and check it is detected

    // TODO bfst.CalculateWeighted
}
