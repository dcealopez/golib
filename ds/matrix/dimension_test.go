package matrix_test

import (
    "slices"
    "testing"

    "github.com/tawesoft/golib/v2/ds/matrix"
)

func TestDimensions_Index(t *testing.T) {
    tests := []struct {
        // args is the parameters passed to NewDimensions
        args []int
        // values is sequence of tests, each encoded as len(args) integer
        // offsets, followed by the expected integer index return value,
        // followed by the actual offsets after any wrapping.
        values []int
        // size is each dimension length multiplied
        size int
    }{
        {
            []int{6}, // 1-Dimensional
            []int{
            //  X, IDX, X'
                0, 0, 0,
                1, 1, 1,
                5, 5, 5,
                6, 0, 0, // wrap-around
                7, 1, 1, // wrap-around
            },
            6,
        },
        {
            []int{3, 3}, // 2-Dimensional
            //   X 0 1 2
            // Y
            // 0   0 1 2
            // 1   3 4 5
            // 2   6 7 8
            []int{
            //  X, Y, IDX, X',Y'
                0, 0, 0, 0, 0,
                1, 0, 1, 1, 0,
                0, 1, 3, 0, 1,
                1, 1, 4, 1, 1,
                2, 1, 5, 2, 1,
                3, 1, 3, 0, 1, // x wraps-around
                1, 3, 1, 1, 0, // y wraps-around
            },
            9,
        },
        {
            []int{4, 3, 2}, // 3-Dimensional
            //   X 0 1 2 3   0 1 2 3
            // Y
            // 0   0 1 2 3   2'3'4'5'
            // 1   4 5 6 7   6'7'8'9'
            // 2   8 9 0'1'  0"1"2"3"
            []int{
            //  X, Y, Z, IDX, X',Y',Z'
                0, 0, 0,  0, 0, 0, 0,
                1, 0, 0,  1, 1, 0, 0,
                2, 2, 1, 22, 2, 2, 1,
                1, 1, 0,  5, 1, 1, 0,
                3, 2, 1, 23, 3, 2, 1,
                4, 1, 1, 16, 0, 1, 1, // x wraps-around
                1, 4, 2,  5, 1, 1, 0, // y,z wraps-around
            },
            24,
        },
        {
            []int{4, 3, 2, 2}, // 4-Dimensional
            //   X 0 1 2 3   0 1 2 3   |  0 1 2 3   0 1 2 3
            // Y
            // 0   0 1 2 3   2'3'4'5'  | 4"5"6"7"   6^7^8^9^
            // 1   4 5 6 7   6'7'8'9'  | 8"9"0^1^   0:1:2:3:
            // 2   8 9 0'1'  0"1"2"3"  | 2^3^4^5^   4:5:6:7:
            []int{
            //  X, Y, Z, W, IDX, X',Y',Z',W'
                0, 0, 0, 0,  0, 0, 0, 0, 0,
                1, 0, 0, 0,  1, 1, 0, 0, 0,
                2, 2, 1, 0, 22, 2, 2, 1, 0,
                1, 1, 0, 0,  5, 1, 1, 0, 0,
                3, 2, 1, 0, 23, 3, 2, 1, 0,
                4, 1, 1, 0, 16, 0, 1, 1, 0, // x wraps-around
                1, 4, 2, 0,  5, 1, 1, 0, 0, // y,z wraps-around
                0, 0, 0, 1, 24, 0, 0, 0, 1,
                1, 0, 0, 1, 25, 1, 0, 0, 1,
                2, 2, 1, 1, 46, 2, 2, 1, 1,
                1, 1, 0, 1, 29, 1, 1, 0, 1,
                3, 2, 1, 1, 47, 3, 2, 1, 1,
                0, 0, 0, 2,  0, 0, 0, 0, 0, // w wraps-around
                1, 0, 0, 2,  1, 1, 0, 0, 0, // w wraps-around
                2, 2, 1, 2, 22, 2, 2, 1, 0, // w wraps-around
                1, 1, 0, 2,  5, 1, 1, 0, 0, // w wraps-around
                3, 2, 1, 2, 23, 3, 2, 1, 0, // w wraps-around
                4, 1, 1, 2, 16, 0, 1, 1, 0, // w,x wraps-around
                1, 4, 2, 2,  5, 1, 1, 0, 0, // w,y,z wraps-around
            },
            48,
        },
        {
            []int{4, 3, 2, 2, 2}, // 5-Dimensional! (tests N-dimensional case)
            //   X 0 1 2 3   0 1 2 3   |  0 1 2 3   0 1 2 3
            // Y
            // 0   0 1 2 3   2'3'4'5'  | 4"5"6"7"   6^7^8^9^ |
            // 1   4 5 6 7   6'7'8'9'  | 8"9"0^1^   0:1:2:3: |
            // 2   8 9 0'1'  0"1"2"3"  | 2^3^4^5^   4:5:6:7: |

            //   X 0 1 2 3   0 1 2 3   |  0 1 2 3   0 1 2 3
            // Y
            // 0   8:9:0~1~  0!1!2!3!  | 2/3/4/5/   4*5*6*7* |
            // 1   2~3~4~5~  4!5!6!7!  | 6/7/8/9/   8*9*0$1$ |
            // 2   6~7~8~9~  8!9!0/1/  | 0*1*2*3*   2$3$4$5$ |
            []int{
            //  X, Y, Z, W, U, IDX, X',Y',Z',W',U'
                3, 2, 1, 0, 0, 23,  3, 2, 1, 0, 0,
                3, 2, 0, 1, 0, 35,  3, 2, 0, 1, 0,
                3, 2, 1, 0, 1, 71,  3, 2, 1, 0, 1,
                3, 2, 0, 1, 1, 83,  3, 2, 0, 1, 1,
                5, 5, 3, 4, 5, 69,  1, 2, 1, 0, 1, // x,y,z,w,u wraps-around
            },
            96,
        },
    }
    for _, tt := range tests {
        dim := matrix.NewDimensions(tt.args...)
        stride := (len(tt.args) * 2) + 1
        numEncodedTests := len(tt.values) / stride

        for i := 0; i < numEncodedTests; i++ {
            offsets := tt.values[i * stride: (i * stride) + len(tt.args)]
            wrappedOffsets := tt.values[(i * stride) + len(tt.args) + 1: ((i + 1) * stride)]
            idx := tt.values[(i * stride) + len(tt.args)]

            if got := dim.Index(offsets...); got != idx {
                t.Errorf("matrix.NewDimensions(%v).Index(%v) = %v, want %v", tt.args, offsets, got, idx)
            }

            offsetsFromIdx := make([]int, len(tt.args))
            dim.Offsets(offsetsFromIdx, idx)
            if !slices.Equal(wrappedOffsets, offsetsFromIdx) {
                t.Errorf("matrix.NewDimensions(%v).Offsets(dest, %v); got %v, want %v;\noriginal offsets %v",
                    tt.args, idx, offsetsFromIdx, wrappedOffsets, offsets)
            }

            if !dim.Contains(wrappedOffsets...) {
                t.Errorf("expected matrix.NewDimensions(%v).Contains(%v)", tt.args, wrappedOffsets)
            }

            if dim.Contains(offsets...) && !slices.Equal(wrappedOffsets, offsets) {
                t.Errorf("expected !matrix.NewDimensions(%v).Contains(%v)", tt.args, offsets)
            }
        }

        if size := dim.Size(); tt.size != size {
            t.Errorf("matrix.NewDimensions(%v).Size(); got %v, want %v", tt.args, size, tt.size)
        }

        if dim.Dimensionality() != len(tt.args) {
            t.Errorf("expected !matrix.NewDimensions(%v).Dimensionality() == %d", tt.args, len(tt.args))
        }

        lengths := make([]int, len(tt.args))
        dim.Lengths(lengths)
        if !slices.Equal(lengths, tt.args) {
            t.Errorf("matrix.NewDimensions(%v).Lengths(); got %v", tt.args, lengths)
        }
    }
}
