package checked

import (
    "math"

    "github.com/tawesoft/golib/v2/must"
    "golang.org/x/exp/constraints"
)

// Limits are a pair of integer values defining a range between an (inclusive)
// minimum value and an (inclusive) maximum value.
type Limits[I constraints.Integer] struct {
    Min I
    Max I
}

// GetLimits returns a filled-in [Limit] representing the widest possible
// minimum and maximum values for a generic integer type.
func GetLimits[I constraints.Integer]() Limits[I] {
    var n Limits[I]
    switch x := any(&n).(type) {
        case *Limits[int]:    *x = Int
        case *Limits[int8]:   *x = Int8
        case *Limits[int16]:  *x = Int16
        case *Limits[int32]:  *x = Int32
        case *Limits[int64]:  *x = Int64
        case *Limits[uint]:   *x = Uint
        case *Limits[uint8]:  *x = Uint8
        case *Limits[uint16]: *x = Uint16
        case *Limits[uint32]: *x = Uint32
        case *Limits[uint64]: *x = Uint64
        default:
            must.Neverf("Limits are not defined for type %T", n)
    }
    return n
}

// Filled-in [Limits] about different integer types with minimum and maximum
// set to the largest range supported by the limit.
//
// For signed integers, these are the appropriately sized math.MinInt and
// math.MaxInt constants. For unsigned integers, these are zero and the
// appropriately sized math.MaxUint constants.
var (
    Int   = Limits[int]  {math.MinInt,   math.MaxInt}
    Int8  = Limits[int8] {math.MinInt8,  math.MaxInt8}
    Int16 = Limits[int16]{math.MinInt16, math.MaxInt16}
    Int32 = Limits[int32]{math.MinInt32, math.MaxInt32}
    Int64 = Limits[int64]{math.MinInt64, math.MaxInt64}

    Uint   = Limits[uint]  {0, math.MaxUint}
    Uint8  = Limits[uint8] {0, math.MaxUint8}
    Uint16 = Limits[uint16]{0, math.MaxUint16}
    Uint32 = Limits[uint32]{0, math.MaxUint32}
    Uint64 = Limits[uint64]{0, math.MaxUint64}
)

// Add returns (a + b, true) iff a, b, and the result all lie between the Limit
// min and max inclusive, otherwise returns (0, false). This calculation is
// robust in the event of integer overflow.
func (l Limits[I]) Add(a I, b I) (I, bool) {
    return Add(l.Min, l.Max, a, b)
}

// Sub returns (a - b, true) iff a, b, and the result all lie between the Limit
// min and max inclusive, otherwise returns (0, false). This calculation is
// robust in the event of integer overflow.
func (l Limits[I]) Sub(a I, b I) (I, bool) {
    return Sub(l.Min, l.Max, a, b)
}

// Mul returns (a * b, true) iff a, b, and the result all lie between the Limit
// min and max inclusive, otherwise returns (0, false). This calculation is
// robust in the event of integer overflow.
func (l Limits[I]) Mul(a I, b I) (I, bool) {
    return Mul(l.Min, l.Max, a, b)
}

// Abs returns (positive i, true) iff both i and the result lie between the
// Limit min and max inclusive. Otherwise, returns (0, false).
func (l Limits[I]) Abs(i I) (I, bool) {
    return Abs(l.Min, l.Max, i)
}

// Inv returns (-i, true) iff both i and the result lie between the Limit min
// and max inclusive. Otherwise, returns (0, false).
func (l Limits[I]) Inv(i I) (I, bool) {
    return Inv(l.Min, l.Max, i)
}
