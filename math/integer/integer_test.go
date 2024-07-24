package integer_test

import (
    "math/bits"
    "testing"

    "github.com/tawesoft/golib/v2/internal/test"
    "github.com/tawesoft/golib/v2/math/integer"
)

func TestPower(t *testing.T) {
    tests := []struct{
        x, n int64
        expected int64
    }{
        { 0,  0,        1},
        { 0,  2,        0},
        { 1, 50,        1},
        { 2,  0,        1},
        { 2, -1,        0},
        { 2,  1,        2},
        { 2,  2,        4},
        { 2,  3,        8},
        { 2, 10,     1024},
        {10,  5,   100000},
        {13,  7, 62748517},
        {13,  1,       13},
    }

    for _, tt := range tests {
        actual := integer.Pow(tt.x, tt.n)
        if actual != tt.expected {
            t.Errorf("Pow(%d, %d) = %d; expected %d", tt.x, tt.n, actual, tt.expected)
        }
    }
}

func TestIsPowTwo(t *testing.T) {
    tests := []struct{
        value uint
        expected bool
    }{
        {0, false},
        {1, true},
        {2, true},
        {3, false},
        {4, true},
        {1 << (bits.UintSize - 1), true},
        {1 + (1 << (bits.UintSize - 1)), false},
        {^uint(0), false},
    }

    for _, tt := range tests {
        actual := integer.IsPowTwo(tt.value)
        if actual != tt.expected {
            t.Errorf("IsPowTwo(%d) = %v; expected %v", tt.value, actual, tt.expected)
        }
    }
}

    func TestIsPowTwo32(t *testing.T) {
        tests := []struct{
            value uint32
            expected bool
        }{
            {0, false},
            {1, true},
            {2, true},
            {3, false},
            {4, true},
            {1 << 31, true},
            {1 + (1 << 31), false},
            {^uint32(0), false},
        }

        for _, tt := range tests {
            actual := integer.IsPowTwo32(tt.value)
            if actual != tt.expected {
                t.Errorf("IsPowTwo32(%d) = %v; expected %v", tt.value, actual, tt.expected)
            }
        }
    }

    func TestIsPowTwo64(t *testing.T) {
        tests := []struct{
            value uint64
            expected bool
        }{
            {0, false},
            {1, true},
            {2, true},
            {3, false},
            {4, true},
            {1 << 63, true},
            {1 + (1 << 63), false},
            {^uint64(0), false},
        }

        for _, tt := range tests {
            actual := integer.IsPowTwo64(tt.value)
            if actual != tt.expected {
                t.Errorf("IsPowTwo64(%d) = %v; expected %v", tt.value, actual, tt.expected)
            }
        }
    }

func TestAlignPowTwo(t *testing.T) {
    tests := []struct{
        value uint
        expected uint // if 0, expect a panic
    }{
        {0, 1},
        {1, 1},
        {2, 2},
        {3, 4},
        {4, 4},
        {(1 << (bits.UintSize - 1)) - 1, 1 << (bits.UintSize - 1)},
        {(1 << (bits.UintSize - 1)),     1 << (bits.UintSize - 1)},
        {(1 << (bits.UintSize - 1)) + 1, 0}, // panics
        {^uint(0), 0}, // panics
    }

    for _, tt := range tests {
        if tt.expected == 0 {
            panics := test.Panics(t, func() { integer.AlignPowTwo(tt.value) }, integer.ErrOverflow)
            if !panics {
                t.Errorf("AlignPowTwo(%d): expected panic", tt.value)
            }
        } else {
            actual := integer.AlignPowTwo(tt.value)
            if actual != tt.expected {
                t.Errorf("AlignPowTwo(%d) = %d; expected %d", tt.value, actual, tt.expected)
            }
        }
    }
}

    func TestAlignPowTwo32(t *testing.T) {
        tests := []struct{
            value uint32
            expected uint32 // if 0, expect a panic
        }{
            {0, 1},
            {1, 1},
            {2, 2},
            {3, 4},
            {4, 4},
            {(1 << 31) - 1, 1 << 31},
            {(1 << 31),     1 << 31},
            {(1 << 31) + 1, 0}, // panics
            {^uint32(0), 0}, // panics
        }

        for _, tt := range tests {
            if tt.expected == 0 {
                panics := test.Panics(t, func() { integer.AlignPowTwo32(tt.value) }, integer.ErrOverflow)
                if !panics {
                    t.Errorf("AlignPowTwo32(%d): expected panic", tt.value)
                }
            } else {
                actual := integer.AlignPowTwo32(tt.value)
                if actual != tt.expected {
                    t.Errorf("AlignPowTwo32(%d) = %d; expected %d", tt.value, actual, tt.expected)
                }
            }
        }
    }

    func TestAlignPowTwo64(t *testing.T) {
        tests := []struct{
            value uint64
            expected uint64 // if 0, expect a panic
        }{
            {0, 1},
            {1, 1},
            {2, 2},
            {3, 4},
            {4, 4},
            {(1 << 63) - 1, 1 << 63},
            {(1 << 63),     1 << 63},
            {(1 << 63) + 1, 0}, // panics
            {^uint64(0), 0}, // panics
        }

        for _, tt := range tests {
            if tt.expected == 0 {
                panics := test.Panics(t, func() { integer.AlignPowTwo64(tt.value) }, integer.ErrOverflow)
                if !panics {
                    t.Errorf("AlignPowTwo64(%d): expected panic", tt.value)
                }
            } else {
                actual := integer.AlignPowTwo64(tt.value)
                if actual != tt.expected {
                    t.Errorf("AlignPowTwo64(%d) = %d; expected %d", tt.value, actual, tt.expected)
                }
            }
        }
    }

func TestNextPowTwo(t *testing.T) {
    tests := []struct{
        value uint
        expected uint // if 0, expect a panic
    }{
        {0, 1},
        {1, 2},
        {2, 4},
        {3, 4},
        {4, 8},
        {(1 << (bits.UintSize - 1)) - 1, 1 << (bits.UintSize - 1)},
        {(1 << (bits.UintSize - 1)), 0}, // panics
        {^uint(0), 0}, // panics
    }

    for _, tt := range tests {
        if tt.expected == 0 {
            panics := test.Panics(t, func() { integer.NextPowTwo(tt.value) }, integer.ErrOverflow)
            if !panics {
                t.Errorf("NextPowTwo(%d): expected panic", tt.value)
            }
        } else {
            actual := integer.NextPowTwo(tt.value)
            if actual != tt.expected {
                t.Errorf("NextPowTwo(%d) = %d; expected %d", tt.value, actual, tt.expected)
            }
        }
    }
}

    func TestNextPowTwo32(t *testing.T) {
        tests := []struct{
            value uint32
            expected uint32
        }{
            {0, 1},
            {1, 2},
            {2, 4},
            {3, 4},
            {4, 8},
            {(1 << 31) - 1, 1 << 31},
            {(1 << 31), 0}, // panics
            {^uint32(0), 0}, // panics
        }

        for _, tt := range tests {
            if tt.expected == 0 {
                panics := test.Panics(t, func() { integer.NextPowTwo32(tt.value) }, integer.ErrOverflow)
                if !panics {
                    t.Errorf("NextPowTwo32(%d): expected panic", tt.value)
                }
            } else {
                actual := integer.NextPowTwo32(tt.value)
                if actual != tt.expected {
                    t.Errorf("NextPowTwo32(%d) = %d; expected %d", tt.value, actual, tt.expected)
                }
            }
        }
    }

    func TestNextPowTwo64(t *testing.T) {
        tests := []struct{
            value uint64
            expected uint64
        }{
            {0, 1},
            {1, 2},
            {2, 4},
            {3, 4},
            {4, 8},
            {(1 << 63) - 1, 1 << 63},
            {(1 << 63), 0}, // panics
            {^uint64(0), 0}, // panics
        }

        for _, tt := range tests {
            if tt.expected == 0 {
                panics := test.Panics(t, func() { integer.NextPowTwo64(tt.value) }, integer.ErrOverflow)
                if !panics {
                    t.Errorf("NextPowTwo64(%d): expected panic", tt.value)
                }
            } else {
                actual := integer.NextPowTwo64(tt.value)
                if actual != tt.expected {
                    t.Errorf("NextPowTwo64(%d) = %d; expected %d", tt.value, actual, tt.expected)
                }
            }
        }
    }
