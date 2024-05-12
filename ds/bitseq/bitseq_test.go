package bitseq_test

import (
    "testing"

    "github.com/tawesoft/golib/v2/ds/bitseq"
)

func expect(t *testing.T, q bool, format string, args ... any) {
    if !q { t.Errorf(format, args...) }
}

func TestStore_String(t *testing.T) {
    const T = true
    const F = false
    var s bitseq.Store
    var args = []struct{
        index    int
        value    bool
        expected string
    }{
        /* 0 */ {  5, F, ""},
        /* 1 */ {  2, T, "001"},
        /* 2 */ {  2, F, ""},
        /* 3 */ {  3, T, "0001"},
        /* 4 */ {  1, T, "0101"},
        /* 5 */ {  8, T, "010100001"},
    }

    for i, arg := range args {
        s.Set(arg.index, arg.value)
        got := s.String()
        expect(t, got == arg.expected, "%d: got %q, expected %q", i, got, arg.expected)
    }
}

func TestStore_NextTrue(t *testing.T) {
    var s bitseq.Store

    // run-length encoded alternating 0s and 1s
    //               0  1  0  1   0  1    0  1    0
    var rle = []int{10, 1, 1, 2, 70, 1, 200, 1, 100}

    // (inclusive)
    // 000 - 009 : 0, 0, 0...,
    // 010 - 010 : 1,
    // 011 - 011 : 0,
    // 012 - 013 : 1, 1,
    // 014 - 083 : 0, 0, 0...,
    // 084 - 084 : 1,
    // 085 - 284 : 0, 0, 0...,
    // 285 - 285 : 1,
    // 286 ...   : 0, 0, 0...,

    // offset to next occurrence of a true bit
    var indexes = []int{10, 12, 13, 84, 285}

    written := 0
    for i := 0; i < len(rle); i++ {
        q := (i % 2) == 1

        // populate the bitseq with the current rle rule
        for j := 0; j < rle[i]; j++ {
            s.Set(written, q)
            written++
        }

        // perform all checks up to that point
        current := -1
        for j := 0; j < len(indexes); j++ {
            idx := indexes[j]
            if idx >= written { break }

            next, ok := s.NextTrue(current)
            expect(t, ok, "expected NextTrue(%d) to be true inside loop after %d written", current, written)
            if !ok { break }
            expect(t, next == idx, "expected NextTrue(%d) to return %d, but got %d after %d written", current, idx, next, written)
            current = next
        }
        _, ok := s.NextTrue(current)
        expect(t, !ok, "expected NextTrue to be false after loop")
    }
}

func TestStore_NextFalse(t *testing.T) {
    var s bitseq.Store

    // run-length encoded alternating 0s and 1s
    //               1  0  1  0   1  0    1  0    1
    var rle = []int{10, 1, 1, 2, 70, 1, 200, 1, 100}
    // and later, s.Set(387, false)

    // (inclusive)
    // 000 - 009 : 1, 1, 1...,
    // 010 - 010 : 0,
    // 011 - 011 : 1,
    // 012 - 013 : 0, 0,
    // 014 - 083 : 1, 1, 1...,
    // 084 - 084 : 0,
    // 085 - 284 : 1, 1, 1...,
    // 285 - 285 : 0,
    // 286 - 385 : 1, 1, 1...,

    // offset to next occurrence of a false bit
    var indexes = []int{10, 12, 13, 84, 285}

    written := 0
    for i := 0; i < len(rle); i++ {
        q := (i % 2) == 0

        // populate the bitseq
        for j := 0; j < rle[i]; j++ {
            s.Set(written, q)
            written++
        }

        // perform all checks up to that point
        current := -1
        for j := 0; j < len(indexes); j++ {
            idx := indexes[j]
            if idx >= written { break }

            next := s.NextFalse(current)
            expect(t, next == idx, "expected NextFalse(%d) to return %d, but got %d after %d written", current, idx, next, written)
            current = next
        }
    }

    // trailing zeros
    current, idx := 285, 386
    next := s.NextFalse(current)
    expect(t, next == idx, "expected NextFalse(%d) to return %d after loop, but got %d after %d written", current, idx, next, written)
}
