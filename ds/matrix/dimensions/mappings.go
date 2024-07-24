package dimensions

import (
    "errors"
    "fmt"
)

// Map describes a new set of dimensions formed by applying a mapping operation
// to an original set of dimensions. For example, this could be rotation,
// translation, cropping, mirroring, or different projections of dimensions.
//
// Map itself implements the [D] interface, but extends it with methods that
// map offsets and indexes back to the original shape.
//
// MapOffsets describes how to convert offsets on the new shape (the source
// arguments) back to offsets on the original shape (the dest arguments).
// len(dest) == original.Dimensionality() and len(offsets) ==
// new.Dimensionality().
//
// MapIndex describes how to convert an index on the new shape (the idx
// argument) back to an index on the original shape (the return value).
type Map interface {
    D
    Original() D
    MapOffsets(dest []int, source ... int)
    MapIndex(idx int) int
}

type mapping struct {
    D
    original D
    offsets func([]int, ... int)
}
    func (m mapping) Original() D {
        return m.original
    }
    func (m mapping) MapOffsets(dest []int, source ... int) {
        m.offsets(dest, source...)
    }
    func (m mapping) MapIndex(idx int) int {
        originalDim := m.original.Dimensionality()
        newDim      := m.D.Dimensionality()

        // share a temporary array for new and original offsets
        offsets := make([]int, newDim + originalDim)

        // Convert new index into new offsets
        m.D.Offsets(offsets[0:newDim], idx)

        // Convert new offsets into original offsets
        m.MapOffsets(offsets[newDim:], offsets[0:newDim]...)

        // And finally convert original offsets into original index
        return m.original.Index(offsets[newDim:]...)
    }

// Mapper can be used to create a [Map] for any shape. If a Map describes a
// mapping between two specific shapes, a Mapper describes a mapping between
// two types of shapes.
type Mapper struct {
    Shapes func(original D) D
    Offsets func(original, new D) func(dest []int, source ... int)
}

    func (m Mapper) Bind(original D) Map {
        new := m.Shapes(original)
        return mapping{
            D: new,
            original: original,
            offsets: m.Offsets(original, new),
        }
    }

// Crop returns a [Map] of a sub-region of the target shape.
//
// It is specified by identifying the index of a start point on the target,
// and the lengths in each dimension.
//
// If the specified length in any direction would take an offset further than
// the target length in that respective dimension, then it is cropped to the
// dimension boundaries. Lengths must be positive.
func Crop(target D, startIdx int, lengths ... int) Map {
    lengths = append([]int{}, lengths...) // don't share memory
    dims := target.Dimensionality()

    startOffsets := make([]int, dims)
    target.Offsets(startOffsets, startIdx)

    // crop to dimension boundaries
    for i := 0; i < dims; i++ {
        if startOffsets[i] + lengths[i] > target.Length(i) {
            lengths[i] = target.Length(i) - startOffsets[i]
        }
    }

    return Mapper{
        Shapes: func(original D) D {
            return New(lengths...)
        },
        Offsets: func(original, new D) func(dest []int, source ... int) {
            return func(dest []int, source ... int) {
                for i := 0; i < original.Dimensionality(); i++ {
                    if i >= len(dest) { break }
                    if i >= len(source) {
                        dest[i] = 0
                        continue
                    }
                    dest[i] = (startOffsets[i] + (source[i] % new.Length(i))) % target.Length(i)
                }
            }
        },
    }.Bind(target)
}

// Sampler returns a new Mapper that can flip, drop, or reorder dimensions
// of shapes arbitrarily.
//
// The string argument encodes operations on each dimension and an ordering
// by the presence, or lack thereof, of a sequence of indexes and modifiers.
//
// The presence of the characters '0'-'9' and 'A'-'F' identify dimensions 0 to
// 15 on the parent. As syntax sugar, the characters in the string "xyzw" are
// respectively interchangeable with the characters "0123". Case is ignored
// in all cases.
//
// A dimension preceded by a negative sign flips or mirrors that
// dimension, so that instead of being read e.g. left to right, or top to
// bottom, it is instead read right to left, or bottom to top.
//
// Omitted dimensions are mapped to offset zero along that dimension in the
// parent. A dimension can be excluded and mapped to a constant offset by
// instead preceding it with an exclamation mark, in which case it is mapped to
// the next element in the optional "constants" argument, which encodes a
// constant integer offset into that dimension.
//
// ASCII whitespace is ignored. The syntax does not support more than 16
// dimensions. A dimension may not be referenced twice. May panic with
// [SwizzleSyntaxError].
//
// For example, given a 2D matrix d2, Reorder(matrix, "yx").Map(d2) returns a
// [Map] that rotates the x & y dimensions in d2, turning it from row major
// order into column major order. Similarly, Reorder(matrix, "-x -y").Map(d2)
// returns a Map that mirrors the matrix along x and y axes. For some 3D matrix
// d3, Reorder(matrix, "xy !z", 4).Map(d3) returns a Map that models a 2D slice
// of d3 along the axis z=4.
func Sampler(reorder string, constants ... int) Mapper {

    // note: this implementation is pretty horrible, but tests confirm it works

    // output[0:5]: parent dimension index 0b00001111
    // output[5]:   unused
    // output[6]:   mirror ?
    // output[7]:   constant ?
    const mask_index  = 0b00001111
    const mask_mirror = 0b01000000
    const mask_const  = 0b10000000

    var outputs [16]uint8
    var seen uint16
    const nothing = uint8(255)
    for i := 0; i < 16; i++ { outputs[i] = nothing }
    precede := nothing
    currentOutput := 0
    currentConstant := 0
    newDims := 0

    // first parse the encoded instructions into outputs
    for i := 0; i < len(reorder); i++ {
        idx := nothing
        c := reorder[i]

        if c == '\t' || c == '\n' || c == ' ' {
            continue
        } else  if ((c == '-') || (c == '!')) && precede == nothing {
            precede = c
        } else if c >= '0' && c <= '9' {
            idx = c - '0'
        } else if c >= 'a' && c <= 'f' {
            idx = c - 'a'
        } else if c >= 'A' && c <= 'F' {
            idx = c - 'A'
        } else {
            switch c {
                case 'x': idx = 0
                case 'y': idx = 1
                case 'z': idx = 2
                case 'w': idx = 3
                case 'X': idx = 0
                case 'Y': idx = 1
                case 'Z': idx = 2
                case 'W': idx = 3
                default:
                    panic(SamplerSyntaxError{
                        Offset:     i,
                        Input:      reorder,
                        Unexpected: c,
                    })
            }
        }
        if idx == nothing { continue }
        if seen & (1 << idx) == (1 << idx) {
            panic(SamplerSyntaxError{
                Offset:     i,
                Input:      reorder,
                Reason:     "dimension referenced twice",
            })
        }
        if (precede == '!') && (currentConstant >= len(constants)) {
            panic(SamplerSyntaxError{
                Offset:     i,
                Input:      reorder,
                Reason:     "constant index out of range",
            })
        }
        seen |= (1 << idx)

        v := idx
        if precede == '-' { v |= mask_mirror }
        if precede == '!' { v |= mask_const; currentConstant++ }
        if precede != '!' { newDims++ }
        precede = nothing
        outputs[currentOutput] = v
        currentOutput++
    }

    if newDims == 0 {
        panic(SamplerSyntaxError{
            Offset:     0,
            Input:      reorder,
            Reason:     "must have at least one non-constant output",
        })
    }

    // now implement a mapper that applies outputs
    constants = append([]int{}, constants[0:currentConstant]...) // don't share memory

    return Mapper{
        Shapes: func(original D) D {
            lengths := make([]int, newDims)
            for i, j := 0, 0; i < currentOutput; i++ {
                if outputs[i] & mask_const == mask_const { continue }
                lengths[j] = original.Length(int(outputs[i] & mask_index))
                j++
            }
            return New(lengths...)
        },
        Offsets: func(original, new D) func(dest []int, source ... int) {
            return func(dest []int, source ... int) {
                for i := 0; i < len(dest); i++ {
                    dest[i] = 0
                }
                for i, c, s := 0, 0, 0; i < 16; i++ {
                    output := outputs[i]
                    if output == nothing {
                        break
                    }
                    idx := output & mask_index
                    if output & mask_const == mask_const {
                        dest[idx] = constants[c]
                        c++
                        continue
                    }
                    var offset int
                    if s < len(source) {
                        offset = source[s] % original.Length(int(idx))
                        s++
                    } else {
                        offset = 0
                    }
                    if output & mask_mirror == mask_mirror {
                        dest[idx] = original.Length(int(idx)) - offset - 1;
                    } else {
                        dest[idx] = offset
                    }
                }
            }
        },
    }
}

type SamplerSyntaxError struct {
    Offset int // byte offset
    Input string
    Unexpected uint8 // character
    Reason string // if not unexpected
}

func (e SamplerSyntaxError) Is(err error) bool {
    var reorderSyntaxError SamplerSyntaxError
    ok := errors.As(err, &reorderSyntaxError)
    return ok
}

func (e SamplerSyntaxError) Error() string {
    if e.Unexpected != 0 {
        return fmt.Sprintf("error parsing dimension reorder string %q: at byte offset %d: unexpected byte 0x%x",
            e.Input, e.Offset, e.Unexpected)
    } else {
        return fmt.Sprintf("error parsing dimension reorder string %q: at byte offset %d: %s",
            e.Input, e.Offset, e.Reason)
    }
}
