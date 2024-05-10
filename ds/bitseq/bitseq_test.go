package bitseq_test

import (
    "testing"

    "github.com/tawesoft/golib/v2/ds/bitseq"
)

func expect(t *testing.T, q bool, format string, args ... any) {
    if !q { t.Errorf(format, args...) }
}

func expectNoPanic(t *testing.T, q func() bool, format string, args ... any) {
    defer func() {
        if r := recover(); r != nil {
            t.Log(r)
            t.Logf(format, args...)
            t.Fatalf("unexpected panic on test")
        }
    }()
    if !q() { t.Errorf(format, args...) }
}

func expectPanic(t *testing.T, f func(), format string, args ... any) {
    defer func() {
        if r := recover(); r == nil {
            t.Errorf(format, args...)
        }
    }()
    f()
}

func TestStore_zeroValue(t *testing.T) {
    var s bitseq.Store
    expect(t, s.Length() == 0, "zero-value has zero length")
    expectPanic(t, func() { s.Pop() }, "zero-value pop should panic")
    expectPanic(t, func() { s.Peek() }, "zero-value peek should panic")
}

func TestStore_pushTrue(t *testing.T) {
    var s bitseq.Store

    for i := 0; i < 1000; i++ {
        s.Push(true)
        expect(t, s.Length() == i + 1, "push length is wrong")
        expect(t, s.Peek(),            "push then peek is wrong")
    }
}

func TestStore_pushFalse(t *testing.T) {
    var s bitseq.Store

    for i := 0; i < 1000; i++ {
        s.Push(false)
        expect(t, s.Length() == i + 1,  "push length is wrong")
        expect(t, s.Peek() == false,    "push then peek is wrong")
    }
}

func TestStore_pushAlternating(t *testing.T) {
    var s bitseq.Store

    for i := 0; i < 1000; i++ {
        q := (i % 2 == 0)
        s.Push(q)
        expect(t, s.Length() == i + 1, "push length is wrong")
        expect(t, s.Peek()   == q,     "push then peek is wrong")
    }

    for i := 0; i < 1000; i++ {
        q := (i % 2 == 0)
        expect(t, s.Get(i) == q,      "s.Get(%d) is wrong", i)
    }
}

func TestStore_sparse(t *testing.T) {
    var s bitseq.Store

    indexes := []int{5, 63, 75, 255, 256, 256, 257, 550, 511, 259}
    lengths := []int{6, 64, 76, 256, 257, 257, 258, 551, 551, 551}

    for i := 0; i < len(indexes); i++ {
        expectNoPanic(t, func() bool { s.Set(indexes[i], true); return true },
            "s.Set() panic at index %d", i)
        expectNoPanic(t, func() bool { return s.Length() == lengths[i] },
            "length is wrong at index %d (got %d, expected %d)",
            i, s.Length(), lengths[i])
    }

    for _, i := range indexes {
        expectNoPanic(t, func() bool { return s.Get(i) }, "s.Get(%d) is wrong", i)
    }

    checks := []int{0, 1, 4, 6, 62, 64, 74, 76, 254, 258, 549, 510, 512, 260}

    for _, i := range checks {
        expectNoPanic(t, func() bool { return !s.Get(i) }, "!s.Get(%d) is wrong", i)
    }

    expectPanic(t, func() { s.Get(552) }, "s.Get() should panic when out of range")

    s.Resize(260)
    expect(t, s.Length() == 260, "s.Length() should be 260 after first resize")
    s.Resize(255)
    expect(t, s.Length() == 255, "s.Length() should be 259 after second resize")
    s.Resize(260)
    expect(t, s.Length() == 260, "s.Length() should be 260 after third resize")

    indexes = []int{5, 63, 75, 255, 256}
    checks  = []int{0, 1, 4, 6, 62, 64, 74, 76, 254, 259}

    for _, i := range indexes {
        expectNoPanic(t, func() bool { return s.Get(i) }, "s.Get(%d) after resize is wrong", i)
    }
    for _, i := range checks {
        expectNoPanic(t, func() bool { return !s.Get(i) }, "!s.Get(%d) after resize is wrong", i)
    }

    expectPanic(t, func() { s.Get(260) }, "s.Get() after resize should panic when out of range")
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

        // populate the bitseq
        for j := 0; j < rle[i]; j++ {
            s.Push(q)
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
            s.Push(q)
            written++
        }

        // perform all checks up to that point
        current := -1
        for j := 0; j < len(indexes); j++ {
            idx := indexes[j]
            if idx >= written { break }

            next, ok := s.NextFalse(current)
            expect(t, ok, "expected NextFalse(%d) to be true inside loop after %d written", current, written)
            if !ok { break }
            expect(t, next == idx, "expected NextFalse(%d) to return %d, but got %d after %d written", current, idx, next, written)
            current = next
        }
        _, ok := s.NextFalse(current)
        expect(t, !ok, "expected NextFalse to be false after loop")
    }

    // test implicit trailing zeros
    s.Set(1000, false)
    next, ok := s.NextFalse(285)
    expect(t, ok, "expected NextFalse to be true after resize")
    expect(t, next == 386, "expected NextFalse(285) to return %d, but got %d after %d written", 386, next, written)
}
