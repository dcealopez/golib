package matrix_test

import (
    "slices"
    "testing"

    "github.com/tawesoft/golib/v2/ds/matrix"
)

func TestGrid_Next(t *testing.T) {
    tests := []struct {
        name string
        constructor matrix.Constructor[int]
        expected []int
        init func(m matrix.M[int])
    }{
        {
            "grid 0",
            matrix.NewGrid[int](4, 4),
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
            nil,
        },
        {
            "grid 2",
            matrix.NewGrid[int](3, 3),
            []int{1, 3, 7},
            func(m matrix.M[int]) {
                m.Set(1, 1)
                m.Set(3, 1)
                m.Set(7, 1)
            },
        },
        {
            "bit 0",
            matrix.NewBit(4, 4),
            []int{},
            nil,
        },
        {
            "bit 1",
            matrix.NewBit(4, 4),
            []int{2, 4, 5, 6, 7, 15},
            func(m matrix.M[int]) {
                m.Set( 2, 1)
                m.Set( 4, 1)
                m.Set( 5, 1)
                m.Set( 6, 1)
                m.Set( 7, 1)
                m.Set(15, 1)
            },
        },
        {
            "diagonal 0",
            matrix.NewSharedDiagonal(2, []int{0, 0, 0, 0}),
            []int{},
            nil,
        },
        {
            "diagonal 1",
            matrix.NewSharedDiagonal(2, []int{0, 0, 1, 1}),
            []int{10, 15},
            nil,
        },
        {
            "hashmap 0",
            matrix.NewSharedHashmap([]int{4, 4}, map[int]int{
            }),
            []int{},
            nil,
        },
        {
            "hashmap 1",
            matrix.NewSharedHashmap([]int{4, 4}, map[int]int{
                 0: 1,
                 3: 1,
                15: 1,
            }),
            []int{0, 3, 15},
            nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := matrix.New(tt.constructor)
            if tt.init != nil { tt.init(m) }
            m = matrix.Const(m)
            indexes := make([]int, 0)
            for idx, ok := -1, true; ok; idx, ok = m.Next(idx) {
                if idx < 0 { continue }
                indexes = append(indexes, idx)
            }
            if !slices.Equal(indexes, tt.expected) {
                t.Errorf("matrix.Next: got indexes %v, want %v", indexes, tt.expected)
            }
        })
    }
}
