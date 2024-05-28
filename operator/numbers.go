package operator

import (
    "golang.org/x/exp/constraints"
)

// Number represents any number type that you can support arithmetic using
// standard Go operators (like a + b, or a ^ b) - i.e. integers & floats.
type Number interface {
     constraints.Integer | constraints.Float
}

// Signed represents any number type that may encode both positive and negative
// values and that support arithmetic on using standard Go operators (like a +
// b, or a ^ b) - i.e. signed integers and floats.
type Signed interface {
     constraints.Signed | constraints.Float
}

// Add returns a + b.
func Add[R Number](a R, b R) R {
    return a + b
}

// Sub returns a - b.
func Sub[R Number](a R, b R) R {
    return a - b
}

// Mul returns a * b.
func Mul[R Number](a R, b R) R {
    return a * b
}

// Div returns a / b.
func Div[R Number](a R, b R) R {
    return a / b
}

// IsPositive returns true iff r >= 0.
func IsPositive[R Number](r R) bool {
    return r >= 0
}

// IsNegative returns true iff r <= 0.
func IsNegative[R Number](r R) bool {
    return r <= 0
}

// IsStrictlyPositive returns true iff r > 0.
func IsStrictlyPositive[R Number](r R) bool {
    return r > 0
}

// IsStrictlyNegative returns true iff r < 0.
func IsStrictlyNegative[R Number](r R) bool {
    return r < 0
}

// Abs returns (0 - r) for r < 0, or r for r >= 0.
func Abs[R Signed](r R) R {
    if (r >= 0) { return r }
    return 0 - r
}

// Inv returns (-r)
func Inv[R Signed](r R) R {
    return 0 - r
}
