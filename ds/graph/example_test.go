package graph_test

import (
    "fmt"
    stdslices "slices"
    "strings"

    "github.com/tawesoft/golib/v2/ds/graph"
    "github.com/tawesoft/golib/v2/fun/slices"
    "github.com/tawesoft/golib/v2/iter"
    "github.com/tawesoft/golib/v2/must"
)

type Distance int

type Body struct {
    Name   string
    Orbits map[graph.VertexIndex]Distance
}

// Universe is a multigraph.
type Universe struct {
    Bodies map[graph.VertexIndex]Body
}

// Vertexes implements the graph.Iterator interface.
func (u Universe) Vertexes() func() (graph.VertexIndex, bool) {
    return iter.Keys(iter.FromMap(u.Bodies))
}

// Edges implements the graph.Iterator interface.
func (u Universe) Edges(source graph.VertexIndex) func() (target graph.VertexIndex, count int, ok bool) {
    vertex := u.Bodies[source]
    it := iter.Keys(iter.FromMap(vertex.Orbits))
    return func() (_ graph.VertexIndex, _ int, _ bool) {
        target, ok := it()
        if !ok { return }
        return target, 1, true
    }
}

// AddVertex implements the graph.Builder interface
func (u *Universe) AddVertex() graph.VertexIndex {
    index := graph.VertexIndex(len(u.Bodies))
    if u.Bodies == nil {
        u.Bodies = make(map[graph.VertexIndex]Body)
    }
    u.Bodies[index] = Body{}
    return index
}

// AddEdge implements the graph.Builder interface
func (u *Universe) AddEdge(from graph.VertexIndex, to graph.VertexIndex) {
    u.Bodies[from].Orbits[to] = Distance(0)
}

func (u *Universe) AddBody(name string) graph.VertexIndex {
    index := u.AddVertex()
    u.Bodies[index] = Body{
        Name: name,
        Orbits: make(map[graph.VertexIndex]Distance),
    }
    return index
}

func (u *Universe) AddOrbit(center graph.VertexIndex, body graph.VertexIndex, distance Distance) {
    u.AddEdge(center, body)
    u.Bodies[center].Orbits[body] = distance
}

func Example_solarSystem() {
    // define a custom multigraph of orbit distances and solar irradiation
    var d Universe

    // Alpha Centuri is trinary star system with a binary star center
    alphaCentauri := d.AddBody("Alpha Centauri")     // Alpha Centauri AB barycenter
    proximaCentauri := d.AddBody("Proxima Centauri") // Alpha Centauri C

    sun     := d.AddBody("Sol")

    mercury := d.AddBody("Mercury")
    venus   := d.AddBody("Venus")
    earth   := d.AddBody("The Earth")
    mars    := d.AddBody("Mars")
    jupiter := d.AddBody("Jupiter")
    saturn  := d.AddBody("Saturn")
    uranus  := d.AddBody("Uranus")
    neptune := d.AddBody("Neptune")

    moon    := d.AddBody("The Moon")
    demos   := d.AddBody("Demos")
    phobos  := d.AddBody("Phobos")

    // Average orbit distance (kilometres)
    // (figures from NASA)
    d.AddOrbit(alphaCentauri, proximaCentauri, 1_937_292_425_565)
    d.AddOrbit(sun,           mercury,                57_000_000)
    d.AddOrbit(sun,           venus,                 108_000_000)
    d.AddOrbit(sun,           earth,                 149_000_000)
    d.AddOrbit(sun,           mars,                  228_000_000)
    d.AddOrbit(sun,           jupiter,               780_000_000)
    d.AddOrbit(sun,           saturn,              1_437_000_000)
    d.AddOrbit(sun,           uranus,              2_871_000_000)
    d.AddOrbit(sun,           neptune,             4_530_000_000)
    d.AddOrbit(earth,         moon,                      384_400)
    d.AddOrbit(mars,          demos,                      23_460)
    d.AddOrbit(mars,          phobos,                      6_000)

    // construct an adjacency matrix for efficient adjacency lookup
    orbitMatrix := graph.NewAdjacencyMatrix(nil)
    orbitMatrix.Calculate(d)

    // construct degree matrices for efficient in-/out- degree lookup
    indegreeMatrix := graph.NewDegreeMatrix()
    outdegreeMatrix := graph.NewDegreeMatrix()
    indegreeMatrix.Calculate(d.Vertexes, orbitMatrix.Indegree)
    outdegreeMatrix.Calculate(d.Vertexes, orbitMatrix.Outdegree)

    // efficiently returns true iff a orbits b
    orbits := func(a graph.VertexIndex, b graph.VertexIndex) bool {
        return orbitMatrix.Get(b, a) > 0
    }

    // efficiently returns the number of satellites orbiting a
    numSatellites := func(a graph.VertexIndex) int {
        return outdegreeMatrix.Get(a)
    }

    // produces target vertexes from orbit edges
    satellites := func(source graph.VertexIndex) func() (graph.VertexIndex, bool) {
        edges := d.Edges(source)
        return func() (graph.VertexIndex, bool) {
            for {
                targetIdx, _, ok := edges()
                return targetIdx, ok
            }
        }
    }

    must.Truef(orbits(earth, sun),   "expected: earth orbits the sun")
    must.Truef(orbits(moon,  earth), "expected: moon orbits the earth")
    must.Falsef(orbits(sun,  earth), "expected: sun does not orbit the earth")

    // efficiently find all graph roots
    roots := iter.ToSlice(graph.Roots(orbitMatrix.Vertexes, indegreeMatrix.Get))

    vertexName := func(x graph.VertexIndex) string {
        vertex := d.Bodies[x]
        return vertex.Name
    }

    sorted := func(xs []string) []string {
        stdslices.Sort(xs)
        return xs
    }

    fmt.Printf("The graph contains data for %d solar systems: %s.\n",
        len(roots),
        strings.Join(sorted(slices.Map(vertexName, roots)), " and "))
    fmt.Printf("The sun is orbited by %d satellites.\n",
        numSatellites(sun))
    fmt.Printf("Mars is orbited by %d satellites: %s.\n",
        numSatellites(mars),
        strings.Join(sorted(slices.Map(vertexName, iter.ToSlice(satellites(mars)))), " and "))

    // Output:
    // The graph contains data for 2 solar systems: Alpha Centauri and Sol.
    // The sun is orbited by 8 satellites.
    // Mars is orbited by 2 satellites: Demos and Phobos.
}
