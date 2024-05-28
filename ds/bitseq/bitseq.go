// Package bitseq efficiently implements a general-purpose infinite sequence of
// bits.
package bitseq

import (
    "encoding/binary"
    "errors"
    "io"
    "math/bits"
    "slices"
    "strings"

    "github.com/tawesoft/golib/v2/ks"
)

var ErrRange = errors.New("value out of range")

// Store is an "infinite" sequence of bits. Trailing zero bits do not
// necessarily consume any memory.
//
// The zero-value Store is a useful value. The store is not suitable for
// concurrent use without additional synchronization.
type Store struct {
    // buckets is the sequence of bits packed into a slice of uint64 values.
    // Trailing zero bits are not necessarily backed by a real bucket.
    buckets []uint64

    // numTrue is the count of true bits
    numTrue int
}

// CountTrue returns the number of true bits in the store.
func (s Store) CountTrue() int {
    return s.numTrue
}

// String implements the stringer interface, and prints a bit sequence as '1'
// and '0' characters from left-to-right. Trailing zeroes are omitted.
func (s Store) String() string {
    var buf strings.Builder
    var trailingZeros = 0

    for _, bucket := range s.buckets {
        for i := 0; i < 64; i++ {
            q := (bucket & (1 << i)) != 0
            if q {
                // delay writing zeroes until a true bit proves that they are
                // not trailing zeroes.
                for j := 0; j < trailingZeros; j++ {
                    buf.WriteByte('0')
                }
                trailingZeros = 0
                buf.WriteByte('1')
            } else {
                trailingZeros++
            }
        }
    }
    return buf.String()
}

// Clear resets the sequence to zeroes.
func (s *Store) Clear() {
    for i := 0; i < cap(s.buckets); i++ {
        s.buckets[i] = 0
    }
    s.numTrue = 0
}

// Crop attempts to reclaim any surplus backing memory consumed by trailing
// zero bits.
func (s *Store) Crop() {
    cropped := s.croppedBuckets()
    if len(cropped) != cap(s.buckets) {
        s.buckets = make([]uint64, len(cropped))
    }
    copy(s.buckets, cropped)
    for i := len(cropped); i < cap(s.buckets); i++ {
        s.buckets[i] = 0
    }
}

// croppedBuckets returns a subslice of buckets. The subslice excludes any
// buckets that are solely trailing zeros.
func (s Store) croppedBuckets() []uint64 {
    if s.buckets == nil { return nil }

    end := len(s.buckets)
    for i := len(s.buckets) - 1; i >= 0; i-- {
        if s.buckets[i] == 0 {
            end = i
        } else {
            break
        }
    }

    return s.buckets[0:end]
}

// magic bytes in the header
const magic = uint64(
    (uint64('B') <<  0) +
    (uint64('i') <<  8) +
    (uint64('t') << 16) +
    (uint64('s') << 24) +
    (uint64('e') << 32) +
    (uint64('q') << 40) +
    (uint64('V') << 48) +
    (uint64('1') << 56))

// Write writes an opaque binary representation of the Store into w.
func(s Store) Write(w io.Writer) error {
    var err error
    var crc uint64

    write := ks.LiftErrorFunc(func(value uint64) error {
        crc = ks.Checksum64(crc, value)
        return binary.Write(w, binary.LittleEndian, value)
    })

    buckets := s.croppedBuckets()
    stats := uint64(len(buckets))
    err = write(err, magic)
    err = write(err, stats)
    for _, c := range buckets {
        err = write(err, c)
    }
    err = write(err, crc)
    return err
}

// Read reads an opaque binary representation from r into the provided Store,
// replacing its existing contents iff successful.
//
// Important: While relatively robust against corrupt data, care should be
// taken when parsing arbitrary input. A malicious actor could craft an input
// that would allocate a large amount of memory, or attempt to extract
// information by continuing to consume from the reader. [io.LimitReader] may
// be helpful here.
func Read(dest *Store, r io.Reader) error {
    return ks.ErrTODO // TODO
}

func fromIndex(index int) (bucket int, offset int) {
    bucket = index / 64
    offset = index % 64
    return
}

// Set sets a bit to true or false at given index.
func (s *Store) Set(index int, bit bool) {
    // Grows the capacity of the backing array as necessary if the index
    // exceeds its current size. Intermediate values are automatically
    // initialised with false bits.

    if index < 0 { panic(ErrRange) }

    bucket, offset := fromIndex(index)
    if bucket >= cap(s.buckets) {
        if !bit { return } // trailing zeros are implied
        s.buckets = slices.Grow(s.buckets, 1 + bucket - cap(s.buckets))
        s.buckets = s.buckets[0:cap(s.buckets)]
    }
    if bit {
        if 0 == (s.buckets[bucket] & (1 << offset)) { s.numTrue++ }
        s.buckets[bucket] = s.buckets[bucket] | (1 << offset)
    } else {
        if 0 != (s.buckets[bucket] & (1 << offset)) { s.numTrue-- }
        s.buckets[bucket] = s.buckets[bucket] & (^(1 << offset))
    }
}

// Get looks up a bit at a given index, returning true iff it has been set.
// Panics if index is less than zero.
func (s Store) Get(index int) bool {
    if (index < 0) { panic(ErrRange) }
    return s.getFromBucket(fromIndex(index))
}

func (s Store) getFromBucket(bucket, offset int) bool {
    if bucket < len(s.buckets) {
        return (s.buckets[bucket] & (1 << offset)) != 0
    } else {
        return false // trailing zeros are implied
    }
}

// NextFalse returns the index of the next false bit found after the given
// index. To start at the beginning, start with NextFalse(-1).
func (s Store) NextFalse(after int) int {
    buckets := len(s.buckets)
    start := after + 1
    bucket := start / 64
    var offset int

    for i := bucket; i < buckets; i++ {
        if s.buckets[i] == ^uint64(0) { continue }

        if i == bucket {
            // First bucket - start search midway through at given offset
            offset = start % 64
        } else {
            // Beginning of a subsequent bucket - start at beginning
            offset = 0
        }

        for j := offset; j < 64; j++ {
            index := (i * 64) + j
            if !s.getFromBucket(i, j) { return index }
        }
    }

    return start
}

// NextTrue returns the index of the next true bit found after the given
// index. To start at the beginning, start with NextTrue(-1). If the second
// return value is false, then the search has finished, and the remaining
// sequence is an infinite sequence of false bits.
func (s Store) NextTrue(after int) (int, bool) {
    buckets := len(s.buckets)
    start := after + 1
    bucket := start / 64

    for i := bucket; i < buckets; i++ {
        if s.buckets[i] == 0 { continue }

        var offset, limit int
        if i == bucket {
            // First bucket - start search midway through. If the entire
            // remainder is trailing zeros, we can skip early.
            offset = start % 64
            limit = 64 - bits.TrailingZeros64(s.buckets[i])
        } else {
            // Beginning of a subsequent bucket, skip zero prefix/suffixes.
            offset = bits.TrailingZeros64(s.buckets[i])
            limit = 64 - bits.LeadingZeros64(s.buckets[i])
        }

        for j := offset; j < limit; j++ {
            index := (i * 64) + j
            if s.getFromBucket(i, j) { return index, true }
        }
    }

    return -1, false
}

// PrevTrue returns the index of the previous true bit found before the given
// index. To start at the end, start with PrevTrue(-1). If the second
// return value is false, then the search has finished, and the remaining
// sequence prefix is either empty or all false bits.
func (s Store) PrevTrue(before int) (int, bool) {
    if before <= 0 { return -1, false }
    start := before - 1
    bucket := start / 64

    for i := bucket; i >= 0; i-- {
        if s.buckets[i] == 0 { continue }

        var offset, limit int
        if i == bucket {
            // First bucket - start search midway through. If the entire
            // prefix is trailing zeros, we can skip early.
            offset = start % 64
            limit = bits.LeadingZeros64(s.buckets[i])
        } else {
            // Beginning of a subsequent bucket, skip zero prefix/suffixes.
            offset = 64 - bits.TrailingZeros64(s.buckets[i])
            limit = bits.LeadingZeros64(s.buckets[i])
        }

        for j := offset; j >= limit; j-- {
            index := (i * 64) + j
            if s.getFromBucket(i, j) { return index, true }
        }
    }

    return -1, false
}
