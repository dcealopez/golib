package dimensions_test

import (
    "errors"
    "fmt"
    "slices"
    "testing"

    "github.com/tawesoft/golib/v2/ds/matrix/dimensions"
    "github.com/tawesoft/golib/v2/must"
)

func ExampleMapper() {
    // This example maps X,Y coordinates onto faces of a 3D box (size 4x3x6),
    // with the 4x3 side on the front and back, 4x6 sides on the top and
    // bottom, and 3x6 on the left and right sides.

    cube := dimensions.New(4, 3, 6)
    var cubeOffsets [3]int

    // define how to map the bottom of any cube to a 2D grid
    bottomMap := dimensions.Mapper{
        Shapes: func(target dimensions.D) dimensions.D {
            return dimensions.New(
                target.Length(0), // X
                target.Length(2), // Z
            )
        },
        Offsets: func(original, new dimensions.D) func(dest []int, source ...int) {
            return func(dest []int, source ...int) {
                dest[0] = source[0]
                dest[1] = original.Length(1) - 1
                dest[2] = original.Length(2) - 1 - source[1] // backwards
            }
        },
    }

    // map the bottom of *this specific* cube
    cubeBottom := bottomMap.Bind(cube)

    // For this example, we'll iterate over x and z for the 4x6 face.
    fmt.Println("Bottom:")
    for z := 0; z < cubeBottom.Length(1); z++ {
        for x := 0; x < cubeBottom.Length(0); x++ {
            // convert (x, z) offsets  on the face to offsets on the target
            cubeBottom.MapOffsets(cubeOffsets[:], x, z)

            fmt.Printf("(%d, %d) -> (%d, %d, %d)\n",
                x, z, cubeOffsets[0], cubeOffsets[1], cubeOffsets[2])
        }
    }

    // define how to map the right side of any cube to a 2D grid
    rightMap := dimensions.Mapper{
        Shapes: func(target dimensions.D) dimensions.D {
            return dimensions.New(
                target.Length(1), // Y
                target.Length(2), // Z
            )
        },
        Offsets: func(original, new dimensions.D) func(dest []int, source ...int) {
            return func(dest []int, source ...int) {
                dest[0] = original.Length(0) - 1 // last slice along X axis
                dest[1] = source[0]
                dest[2] = source[1]
            }
        },
    }

    // map the right side of *this specific* cube
    cubeRight := rightMap.Bind(cube)

    // For this example, we'll iterate over indexes 0 .. 18 for the 3x6 face.
    fmt.Println("\nRight:")
    for idx := 0; idx < cubeRight.Size(); idx++ {
        // convert indexes on the face to indexes on the target
        cubeIdx := cubeRight.MapIndex(idx)

        fmt.Printf("%d -> %d\n", idx, cubeIdx)
    }

    // Output:
    // Bottom:
    // (0, 0) -> (0, 2, 5)
    // (1, 0) -> (1, 2, 5)
    // (2, 0) -> (2, 2, 5)
    // (3, 0) -> (3, 2, 5)
    // (0, 1) -> (0, 2, 4)
    // (1, 1) -> (1, 2, 4)
    // (2, 1) -> (2, 2, 4)
    // (3, 1) -> (3, 2, 4)
    // (0, 2) -> (0, 2, 3)
    // (1, 2) -> (1, 2, 3)
    // (2, 2) -> (2, 2, 3)
    // (3, 2) -> (3, 2, 3)
    // (0, 3) -> (0, 2, 2)
    // (1, 3) -> (1, 2, 2)
    // (2, 3) -> (2, 2, 2)
    // (3, 3) -> (3, 2, 2)
    // (0, 4) -> (0, 2, 1)
    // (1, 4) -> (1, 2, 1)
    // (2, 4) -> (2, 2, 1)
    // (3, 4) -> (3, 2, 1)
    // (0, 5) -> (0, 2, 0)
    // (1, 5) -> (1, 2, 0)
    // (2, 5) -> (2, 2, 0)
    // (3, 5) -> (3, 2, 0)
    //
    // Right:
    // 0 -> 3
    // 1 -> 7
    // 2 -> 11
    // 3 -> 15
    // 4 -> 19
    // 5 -> 23
    // 6 -> 27
    // 7 -> 31
    // 8 -> 35
    // 9 -> 39
    // 10 -> 43
    // 11 -> 47
    // 12 -> 51
    // 13 -> 55
    // 14 -> 59
    // 15 -> 63
    // 16 -> 67
    // 17 -> 71
}

func TestD_Index(t *testing.T) {
    tests := []struct {
        // args is the parameters passed to New
        args []int
        // values is sequence of tests, each encoded as len(args) integer
        // offsets, followed by the expected integer index return value,
        // followed by the actual offsets after any wrapping.
        //
        // For example, X, Y, IDX, X', Y' encodes offsets X and Y, followed
        // by the expected index after converting offsets X and Y to an index,
        // followed by X' and Y', the expected values for X and Y after
        // wrapping modulo dimension length.
        values []int
        // size is each dimension length multiplied e.g. width times height
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
        dim := dimensions.New(tt.args...)
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

func TestMapping(t *testing.T) {
    tests := []struct {
        name string
        // args are the parameters passed to New
        args []int
        // mapper performs some mapping e.g. crop, reorder
        mapper func(in dimensions.D) dimensions.Map
        // lengths is the expected dimensions of the new shape
        lengths []int
        // values is sequence of tests, each encoded as len(lengths) offsets on
        // the new shape, and the expected len(args) offsets on the old shape.
        // e.g. for a mapping 3D -> 2D, []int{X,Y, X', Y', Z'}.
        values []int
        // err, if not nil, checks a panic raised by mapper
        err error
    }{
        {
            "crop/identity",
            []int{5, 6}, // 2D
            func(in dimensions.D) dimensions.Map {
                // a bigger crop just gets clamped to boundaries
                return dimensions.Crop(in, 0, 6, 7)
            },
            []int{5, 6},
            []int{
             // X, Y, X',Y'
                0, 0, 0, 0,
                4, 4, 4, 4,
                5, 6, 0, 0, // wrap
                6, 7, 1, 1, // wrap
            },
            nil,
        },
        {
            "crop/center",
            []int{9, 9}, // 2D
            func(in dimensions.D) dimensions.Map {
                idx := in.Index(2, 3)
                return dimensions.Crop(in, idx, 4, 5)
            },
            []int{4, 5},
            []int{
             // X, Y, X',Y'
                0, 0, 2, 3,
                3, 3, 5, 6,
                4, 5, 2, 3, // wraps on the mapped shape
                5, 6, 3, 4, // wraps on the mapped shape

                // X: 2 3 4 5 | 2 3 4 5 | 2 3 <- 10th
                // Y: 3 4 5 6 7 | 3 4 5 6 7 <- 10th
                9, 9, 3, 7, // wraps even on the original shape
            },
            nil,
        },
        {
            "reorder/no-dimensions",
            []int{2, 3, 4}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("").Bind(in)
            },
            []int{},
            []int{},
            dimensions.SamplerSyntaxError{},
        },
        {
            "reorder/all-const",
            []int{2, 3, 4}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("!x !y !z", 1, 2, 3).Bind(in)
            },
            []int{},
            []int{},
            dimensions.SamplerSyntaxError{},
        },
        {
            "reorder/duplicate-dimension",
            []int{2, 3, 4}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("xxy").Bind(in)
            },
            []int{},
            []int{},
            dimensions.SamplerSyntaxError{},
        },
        {
            "reorder/identity",
            []int{2, 3, 4}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("xyz").Bind(in)
            },
            []int{2, 3, 4},
            []int{
             // X, Y, Z, X',Y',Z'
                0, 0, 0, 0, 0, 0,
                1, 2, 3, 1, 2, 3,
                2, 3, 4, 0, 0, 0, // wraps
                3, 4, 5, 1, 1, 1, // wraps
            },
            nil,
        },
        {
            "reorder/mirror",
            []int{2, 3, 4}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("-x -y -z").Bind(in)
            },
            []int{2, 3, 4},
            []int{
             // X, Y, Z, X',Y',Z'
                0, 0, 0, 1, 2, 3,
                1, 1, 1, 0, 1, 2,
                1, 2, 3, 0, 0, 0,
                2, 3, 4, 1, 2, 3, // wraps
            },
            nil,
        },
        {
            "reorder/rotate",
            []int{2, 3, 4}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("zyx").Bind(in)
            },
            []int{4, 3, 2},
            []int{
             // X, Y, Z, X',Y',Z'
                0, 0, 0, 0, 0, 0,
                3, 2, 1, 1, 2, 3,
                4, 3, 2, 0, 0, 0, // wraps
                5, 4, 3, 1, 1, 1, // wraps
            },
            nil,
        },
        {
            "reorder/face",
            []int{3, 4, 5}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("xy").Bind(in)
            },
            []int{3, 4},
            []int{
             // X, Y, X',Y',Z'
                0, 0, 0, 0, 0,
                2, 3, 2, 3, 0,
                3, 4, 0, 0, 0, // wraps
                4, 5, 1, 1, 0, // wraps
            },
            nil,
        },
        {
            "reorder/face",
            []int{3, 4, 5}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("xy !z", 4).Bind(in)
            },
            []int{3, 4},
            []int{
             // X, Y, X',Y',Z'
                0, 0, 0, 0, 4,
                2, 3, 2, 3, 4,
                3, 4, 0, 0, 4, // wraps
                4, 5, 1, 1, 4, // wraps
            },
            nil,
        },
        {
            "reorder/row",
            []int{3, 4, 5}, // 3D
            func(in dimensions.D) dimensions.Map {
                return dimensions.Sampler("x !z", 4).Bind(in)
            },
            []int{3},
            []int{
             // X, X',Y',Z'
                0, 0, 0, 4,
                2, 2, 0, 4,
                3, 0, 0, 4, // wraps
                4, 1, 0, 4, // wraps
            },
            nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            original := dimensions.New(tt.args...)

            mapped, err := must.Try(func() dimensions.Map {
                return tt.mapper(original)
            })()
            if err != nil && tt.err == nil {
                t.Errorf("unexpected error: %v", err)
                return
            }
            if err == nil && tt.err != nil {
                t.Errorf("expected error, but got none")
                return
            }
            if err != nil && tt.err != nil && !errors.Is(err, tt.err) {
                t.Errorf("expected errors.Is(%T), but got %T (%v)", tt.err, err, err)
                return
            }
            if err != nil { return }

            if mapped.Dimensionality() != len(tt.lengths) {
                t.Errorf("dimensionality %v, want %v", mapped.Dimensionality(), len(tt.lengths))
            } else {
                for i := 0; i < mapped.Dimensionality(); i++ {
                    if mapped.Length(i) != tt.lengths[i] {
                        t.Errorf("length(%d) %v, want %v", i, mapped.Length(i), tt.lengths[i])
                    }
                }
            }
            stride := len(tt.args) + len(tt.lengths)
            numEncodedTests := len(tt.values) / stride
            for i := 0; i < numEncodedTests; i++ {
                row := tt.values[i * stride : (i * stride) + stride]
                viewOffsets := row[0:len(tt.lengths)]
                expectedOffsets := row[len(tt.lengths):]
                originalOffsets := make([]int, len(tt.args))
                mapped.MapOffsets(originalOffsets, viewOffsets...)

                if !slices.Equal(expectedOffsets, originalOffsets) {
                    t.Errorf("map offsets %v: got %v, want %v",
                        viewOffsets, originalOffsets, expectedOffsets)
                }
            }
        })
    }
}
