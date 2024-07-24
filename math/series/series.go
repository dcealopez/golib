package series

import (
    "math"

    "github.com/tawesoft/golib/v2/math/integer"
    "golang.org/x/exp/constraints"
)

type Number interface {
    constraints.Float | constraints.Integer
}

// abs returns the absolute value of x, even if x is unsigned.
func abs[T Number](x T) T {
    if x >= 0 { return x }
    return -x
}

// Geometric represents the series a + ar + ar^2 + ar^3 + ... for some
// coefficient a and some ratio r.
type Geometric[N Number] struct {
    coefficient, ratio N
    powerfunc func(a, b N) N
}

    func (g Geometric[N]) Coefficient() N { return g.coefficient }
    func (g Geometric[N]) Ratio() N       { return g.ratio }

    func NewGeometricFloat(coefficient, ratio float64) Geometric[float64] {
        return Geometric[float64]{
            coefficient: coefficient,
            ratio:       ratio,
            powerfunc:   math.Pow,
        }
    }

    func NewGeometricInteger[N constraints.Integer](coefficient, ratio N) Geometric[N] {
        return Geometric[N]{
            coefficient: coefficient,
            ratio:       ratio,
            powerfunc:   integer.Pow[N],
        }
    }

    // Sum returns the sum of the first n terms of the series g.
    func (g Geometric[N]) Sum(n int) N {
        if g.ratio == 0 { return N(n + 1) * g.coefficient }
        if g.ratio == 1 { return g.ratio * g.coefficient }
        rN := g.powerfunc(g.ratio, N(n+1))
        return g.coefficient * ((1 - rN) / (1 - g.ratio))
    }

    // Term returns the nth term of the series g, starting at n = 0.
    func (g Geometric[N]) Term(n int) N {
        if g.ratio == 0 { return g.coefficient }
        return g.coefficient * g.powerfunc(g.ratio, N(n))
    }

    // Limit returns the limit towards which the infinite series g converges.
    // Iff the second return value is false, then the series does not converge.
    func (g Geometric[N]) Limit() (N, bool) {
        if g.ratio == 0 { return g.coefficient, true }
        if abs(g.ratio) >= 1 { return 0, false }
        return g.coefficient / (1 - g.ratio), true
    }
