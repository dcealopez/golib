// Package checked (operator/checked) implements operations bounded by
// limits that are robust in the event of integer overflow.
//
// Note: Float variants are not fully tested yet.
package checked

// Add returns (a + b, true) iff a, b, and the result all lie between min and
// max inclusive, otherwise returns (0, false). This calculation is robust in
// the event of integer overflow.
func Add[N Number](min N, max N, a N, b N) (N, bool) {
    if (a < min) || (a > max) || (b < min) || (b > max) { return 0, false }
    if (min > max) || (max < min) { return 0, false }
    if (b > 0) && (a > (max - b)) { return 0, false }
    if (b < 0) && (a < (min - b)) { return 0, false }
    return a + b, true
}

// Sub returns (a - b, true) iff a, b, and the result all lie between min and
// max inclusive, otherwise returns (0, false). This calculation is robust in
// the event of integer overflow.
func Sub[N Number](min N, max N, a N, b N) (N, bool) {
    if (a < min) || (a > max) || (b < min) || (b > max) { return 0, false }
    if (min > max) || (max < min) { return 0, false }
    if (b < 0) && (a > (max + b)) { return 0, false }
    if (b > 0) && (a < (min + b)) { return 0, false }
    return a - b, true
}

// Mul returns (a * b, true) iff a, b, and the result all lie between min and
// max inclusive, otherwise returns (0, false). This calculation is robust in
// the event of integer overflow.
func Mul[N Number](min N, max N, a N, b N) (N, bool) {
    if (a < min) || (a > max) || (b < min) || (b > max) { return 0, false }
    if (min > max) || (max < min) { return 0, false }

    x := a * b
    if (x < min) || (x > max) { return 0, false }
    if (x != 0) && (a != x/b) { return 0, false }
    return x, true
}

// Abs returns (positive i, true) iff both i and the result lie between min and
// max inclusive. Otherwise, returns (0, false).
func Abs[N Number](min N, max N, i N) (N, bool) {
    if (i < min) || (i > max) { return 0, false }
    if (min > max) || (max < min) { return 0, false }
    if (i < 0) && (0 > (max + i)) { return 0, false }
    if (i > 0) && (0 < (min + i)) { return 0, false }
    if i < 0 { return -i, true } else { return i, true }
}

// Inv returns (-i, true) iff both i and the result lie between min and max
// inclusive. Otherwise, returns (0, false).
func Inv[N Number](min N, max N, i N) (N, bool) {
    if (i < min) || (i > max) { return 0, false }
    if (min > max) || (max < min) { return 0, false }
    if (i < 0) && (0 > (max + i)) { return 0, false }
    if (i > 0) && (0 < (min + i)) { return 0, false }
    return -i, true
}
