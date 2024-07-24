package series_test

import (
    "math"
    "testing"

    "github.com/tawesoft/golib/v2/math/series"
)

func TestGeometric_integer(t *testing.T) {
    tests := []struct{
        coefficient, ratio int64
        terms, sums []int64
        converges bool
        limit int64
    }{
        {
            coefficient:  1,
            ratio:        2,
            // 1 + 1r + 1r^2 + 1r^3...
            terms:        []int64{1, 2, 4, 8},
            sums:         []int64{1, 1+2, 1+2+4, 1+2+4+8},
        },
        {
            coefficient:  2,
            ratio:        0,
            // 2 + 2 + 2 + 2...
            terms:        []int64{2, 2, 2, 2},
            sums:         []int64{2, 4, 6, 8},
            converges:    true,
            limit:        2,
        },
        {
            coefficient:  3,
            ratio:        4,
            // 3 + (3*4) + (3*16) + (3*64)...
            terms:        []int64{3, 12, 48, 192},
            sums:         []int64{3, 3+12, 3+12+48, 3+12+48+192},
        },
        {
            // Grandi's series
            coefficient:  1,
            ratio:       -1,
            terms:        []int64{ 1, -1,  1, -1,  1, -1},
            sums:         []int64{ 1,  0,  1,  0,  1,  0},
        },
    }

    for _, tt := range tests {
        g := series.NewGeometricInteger(tt.coefficient, tt.ratio)

        for i := 0; i < len(tt.terms); i++ {
            actual := g.Term(i)
            if actual != tt.terms[i] {
                t.Errorf("NewGeometricInteger(%d, %d).Term(%d): got %d, want %d",
                    tt.coefficient, tt.ratio, i, actual, tt.terms[i])
            }
        }

        for i := 0; i < len(tt.sums); i++ {
            actual := g.Sum(i)
            if actual != tt.sums[i] {
                t.Errorf("NewGeometricInteger(%d, %d).Sum(%d): got %d, want %d",
                    tt.coefficient, tt.ratio, i, actual, tt.sums[i])
            }
        }

        limit, converges := g.Limit()
        if converges != tt.converges {
            t.Errorf("NewGeometricInteger(%d, %d).Limit(): got %d, %t, wanted %t",
                tt.coefficient, tt.ratio, limit, converges, tt.converges)
        } else if converges && (limit != tt.limit) {
            t.Errorf("NewGeometricInteger(%d, %d).Limit(): got %d, wanted %d",
                tt.coefficient, tt.ratio, limit, tt.limit)
        }
    }
}

func TestGeometric_float(t *testing.T) {
    tests := []struct{
        coefficient, ratio float64
        terms, sums []float64
        converges bool
        limit float64
    }{
        {
            coefficient:  1,
            ratio:        2,
            terms:        []float64{1, 2, 4, 8},
            sums:         []float64{1, 1+2, 1+2+4, 1+2+4+8},
        },
        {
            coefficient:  2,
            ratio:        0,
            terms:        []float64{2, 2, 2, 2},
            sums:         []float64{2, 4, 6, 8},
            converges:    true,
            limit:        2,
        },
        {
            coefficient:  3,
            ratio:        4,
            terms:        []float64{3, 12, 48, 192},
            sums:         []float64{3, 3+12, 3+12+48, 3+12+48+192},
        },
        {
            // Grandi's series
            coefficient:  1,
            ratio:       -1,
            terms:        []float64{ 1, -1,  1, -1,  1, -1},
            sums:         []float64{ 1,  0,  1,  0,  1,  0},
        },
        {
            coefficient:  1,
            ratio:        0.5,
            terms:        []float64{1.000, 0.500, 0.250, 0.125},
            sums:         []float64{1.000, 1.500, 1.750, 1.875},
            converges:    true,
            limit:        2.0,
        },
        {
            coefficient:  1,
            ratio:       -0.5,
            terms:        []float64{1.000,-0.500, 0.250,-0.125},
            sums:         []float64{1.000, 0.500, 0.750, 0.625},
            converges:    true,
            limit:        2.0/3.0,
        },
    }

    near := func(a, b float64) bool {
        return math.Abs(a - b) < 0.001
    }

    for _, tt := range tests {
        g := series.NewGeometricFloat(tt.coefficient, tt.ratio)

        for i := 0; i < len(tt.terms); i++ {
            actual := g.Term(i)
            if !near(actual, tt.terms[i]) {
                t.Errorf("NewGeometricFloat(%f, %f).Term(%d): got %f, want %f",
                    tt.coefficient, tt.ratio, i, actual, tt.terms[i])
            }
        }

        for i := 0; i < len(tt.sums); i++ {
            actual := g.Sum(i)
            if !near(actual, tt.sums[i]) {
                t.Errorf("NewGeometricFloat(%f, %f).Sum(%d): got %f, want %f",
                    tt.coefficient, tt.ratio, i, actual, tt.sums[i])
            }
        }

        limit, converges := g.Limit()
        if converges != tt.converges {
            t.Errorf("NewGeometricFloat(%f, %f).Limit(): got %f, %t, want %t",
                tt.coefficient, tt.ratio, limit, converges, tt.converges)
        } else if converges && (limit != tt.limit) {
            t.Errorf("NewGeometricFloat(%f, %f).Limit(): got %f, want %f",
                tt.coefficient, tt.ratio, limit, tt.limit)
        }
    }
}
