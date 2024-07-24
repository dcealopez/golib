package matrix_test

import (
    "slices"
    "testing"

    "github.com/tawesoft/golib/v2/ds/matrix"
)

func TestM_Next(t *testing.T) {
    tests := []struct {
        name string
        matrix matrix.M[int]
        expectedIndexes []int // non-zero
        expectedValues []int // at expectedIndexes
        init func(m matrix.M[int])
    }{
        {
            "grid 0",
            matrix.NewGrid[int](4, 4),
            []int{},
            []int{},
            nil,
        },
        {
            "grid 1",
            matrix.NewSharedGrid([]int{4, 4}, []int{
                0, 0, 1, 0,
                2, 3, 4, 5,
                6, 0, 0, 0,
                0, 0, 0, 1,
                }),
            []int{2, 4, 5, 6, 7, 8, 15},
            []int{1, 2, 3, 4, 5, 6,  1},
            nil,
        },
        {
            "grid 2",
            matrix.NewGrid[int](3, 3),
            []int{1, 3, 7},
            []int{10, 20, 30},
            func(m matrix.M[int]) {
                m.Set(1, 10)
                m.Set(3, 20)
                m.Set(7, 30)
            },
        },
        {
            "bit 0",
            matrix.NewBit(4, 4),
            []int{},
            []int{},
            nil,
        },
        {
            "bit 1",
            matrix.NewBit(4, 4),
            []int{2, 4, 5, 6, 7, 15},
            []int{1, 1, 1, 1, 1, 1},
            func(m matrix.M[int]) {
                m.Set( 2, 10)
                m.Set( 4, 20)
                m.Set( 5, 30)
                m.Set( 6, 40)
                m.Set( 7, 50)
                m.Set(15, 60)
            },
        },
        {
            "diagonal 0",
            matrix.NewSharedDiagonal(2, []int{0, 0, 0, 0}),
            []int{},
            []int{},
            nil,
        },
        {
            "diagonal 1",
            matrix.NewSharedDiagonal(2, []int{0, 0, 1, 2}),
            []int{10, 15},
            []int{ 1,  2},
            nil,
        },
        {
            "hashmap 0",
            matrix.NewSharedHashmap([]int{4, 4}, map[int]int{
            }),
            []int{},
            []int{},
            nil,
        },
        {
            "hashmap 1",
            matrix.NewSharedHashmap([]int{4, 4}, map[int]int{
                 0: 11,
                 3: 12,
                15: 13,
            }),
            []int{0, 3, 15},
            []int{11, 12, 13},
            nil,
        },
        {
            "view 1",
            matrix.Region[int](matrix.NewSharedGrid([]int{4, 4}, []int{
                0, 0, 1, 0,
                2, 3, 4, 5,
                6, 0, 0, 0,
                0, 0, 0, 1,
            }), 5, 3, 2),
            []int{0, 1, 2},
            []int{3, 4, 5},
            nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := tt.matrix
            if tt.init != nil { tt.init(m) }
            m = matrix.Const(m)
            indexes := make([]int, 0)
            values := make([]int, 0)
            for idx, ok := -1, true; ok; idx, ok = m.Next(idx) {
                if idx < 0 { continue }
                indexes = append(indexes, idx)
                values = append(values, m.Get(idx))
            }
            if !slices.Equal(indexes, tt.expectedIndexes) {
                t.Errorf("matrix.Next: got indexes %v, want %v", indexes, tt.expectedIndexes)
            }
            if !slices.Equal(values, tt.expectedValues) {
                t.Errorf("matrix.Next: got values %v, want %v", values, tt.expectedValues)
            }
        })
    }
}
