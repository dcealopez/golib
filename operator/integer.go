package operator

import (
    "github.com/tawesoft/golib/v2/math/integer"
    "golang.org/x/exp/constraints"
)

// BitwiseAnd returns bitwise AND (i.e. `a & b`) of integer inputs.
func BitwiseAnd[I constraints.Integer](a I, b I) I {
    return a & b
}

// BitwiseOr returns bitwise OR (i.e. `a | b`) of integer inputs.
func BitwiseOr[I constraints.Integer](a I, b I) I {
    return a | b
}

// BitwiseXor returns bitwise XOR (i.e. `a ^ b`) of integer inputs.
func BitwiseXor[I constraints.Integer](a I, b I) I {
    return a ^ b
}

// BitwiseNot returns bitwise complement. This is `m ^ x` with m = "all bits
// set to 1" for unsigned x, and m = -1 for Signed x.
func BitwiseNot[I constraints.Integer](i I) I {
    return ^i
}

// Mod returns a mod b (i.e. `a % b`) of integer inputs.
func Mod[I constraints.Integer](a I, b I) I {
    return a % b
}

// Pow returns a to the power of b for integer inputs.
func Pow[I constraints.Integer](a, b I) I {
    return integer.Pow(a, b)
}

// ShiftLeft returns bitwise shift left (i.e. `a << b`) of integer inputs.
func ShiftLeft[I constraints.Integer](a I, b I) I {
    return a << b
}

// ShiftRight returns bitwise shift right (i.e. `a >> b`) of integer inputs.
func ShiftRight[I constraints.Integer](a I, b I) I {
    return a >> b
}
