package integer

import (
    "errors"
    "math/bits"

    "golang.org/x/exp/constraints"
)

var ErrOverflow = errors.New("integer overflow")

// Pow implements x^n for integers.
//
// Note that this function is not robust in the event of integer overflow.
func Pow[T constraints.Integer](x, n T) T {
    // https://en.wikipedia.org/wiki/Exponentiation_by_squaring
    if x == 0 && n == 0 {
        // 0^0 is usually either left undefined, or defined as 1.
        // https://en.wikipedia.org/wiki/Zero_to_the_power_of_zero#Treatment_on_computers
        return 1
    }
    if n  < 0 { return 0 } // 0 < x < 1, but truncated to zero.
    if n == 0 { return 1 }
    y := T(1)
    for n > 1 {
        if n % 2 == 1 {
            y = x * y
            n = n - 1
        }
        x = x * x
        n = n / 2
    }
    return x * y
}

// AlignPowTwo returns the next 2^n >= x, for some integer n.
//
// If x > 2^(bits.UintSize - 1), this function panics with [ErrOverflow].
func AlignPowTwo(x uint) uint {
    if x <= 1 { return 1 }
    if x > (1 << (bits.UintSize -1)) { panic(ErrOverflow) }

    // e.g. 0b000...0001000 -> 0x000...0000111 -> 3
    prefix := bits.UintSize - bits.LeadingZeros(x - 1)

    // e.g. 0b000...0000111 -> 0b000...00001000 (1 << 3)
    return 1 << prefix
}

    // AlignPowTwo32 returns the next 2^n >= x, for some integer n.
    //
    // If x > 2^31, this function panics with [ErrOverflow].
    func AlignPowTwo32(x uint32) uint32 {
        if x <= 1 { return 1 }
        if x > (1 << 31) { panic(ErrOverflow) }

        // e.g. 0b000...0001000 -> 0x000...0000111 -> 3
        prefix := 32 - bits.LeadingZeros32(x - 1)

        // e.g. 0b000...0000111 -> 0b000...00001000 (1 << 3)
        return 1 << prefix
    }

    // AlignPowTwo64 returns the next 2^n >= x, for some integer n.
    //
    // If x > 2^63, this function panics with [ErrOverflow].
    func AlignPowTwo64(x uint64) uint64 {
        if x <= 1 { return 1 }
        if x > (1 << 63) { panic(ErrOverflow) }

        // e.g. 0b000...0001000 -> 0x000...0000111 -> 3
        prefix := 64 - bits.LeadingZeros64(x - 1)

        // e.g. 0b000...0000111 -> 0b000...00001000 (1 << 3)
        return 1 << prefix
    }

// NextPowTwo returns the next 2^n > x, for some integer n.
//
// If x >= 2^(bits.UintSize - 1), this function panics with [ErrOverflow].
func NextPowTwo(x uint) uint {
    if x == 0 { return 1 }
    if x >= (1 << (bits.UintSize - 1)) { panic(ErrOverflow) }

    // e.g. 0b000...0001100 -> 4
    prefix := bits.UintSize - bits.LeadingZeros(x)

    // e.g. 0b000...0001100 -> 0b000...00010000 (1 << 4)
    return 1 << prefix
}

    // NextPowTwo32 returns the next 2^n > x, for some integer n.
    //
    // If x >= 2^31, this function panics with [ErrOverflow].
    func NextPowTwo32(x uint32) uint32 {
        if x == 0 { return 1 }
        if x >= (1 << 31) { panic(ErrOverflow) }

        // e.g. 0b000...0001100 -> 4
        prefix := 32 - bits.LeadingZeros32(x)

        // e.g. 0b000...0001100 -> 0b000...00010000 (1 << 4)
        return 1 << prefix
    }

    // NextPowTwo64 returns the next 2^n > x, for some integer n.
    //
    // If x >= 2^63, this function panics with [ErrOverflow].
    func NextPowTwo64(x uint64) uint64 {
        if x == 0 { return 1 }
        if x >= (1 << 63) { panic(ErrOverflow) }

        // e.g. 0b000...0001100 -> 4
        prefix := 64 - bits.LeadingZeros64(x)

        // e.g. 0b000...0001100 -> 0b000...00010000 (1 << 4)
        return 1 << prefix
    }

// IsPowTwo returns true iff x == 2^n for some integer n >= 0.
func IsPowTwo(x uint) bool {
    return bits.OnesCount(x) == 1
}

    // IsPowTwo32 returns true iff x == 2^n for some integer n >= 0.
    func IsPowTwo32(x uint32) bool {
        return bits.OnesCount32(x) == 1
    }

    // IsPowTwo64 returns true iff x == 2^n for some integer n >= 0.
    func IsPowTwo64(x uint64) bool {
        return bits.OnesCount64(x) == 1
    }
