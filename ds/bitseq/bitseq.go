// Package bitseq efficiently implements a general-purpose and variable-length
// sequence of bits.
package bitseq

import (
    "encoding/binary"
    "errors"
    "fmt"
    "io"
    "math/bits"
    "slices"
    "strings"

    "github.com/tawesoft/golib/v2/ks"
)

var ErrRange = errors.New("value out of range")

// Store is a variable-length container of bits.
//
// The zero-value is a useful value. The store is not suitable for concurrent
// use without additional synchronization.
type Store struct {
    // length records the logical number of bits stored.
    //
    // Note: This may be less than or greater than the number
    // of bits actually physically available in buckets.
    length int

    // buckets is the sequence of bits packed into a slice of uint64 values.
    // Trailing zero bits are not necessarily backed by a real bucket.
    buckets []uint64
}

// String implements the stringer interface, and prints a bit sequence as '1'
// and '0' characters from left-to-right.
func (s *Store) String() string {
    var buf strings.Builder
    var n = 0
    for _, bucket := range s.buckets {
        for i := 0; i < 64; i++ {
            q := (bucket & (1 << i)) != 0
            if q {
                buf.WriteByte('1')
            } else {
                buf.WriteByte('0')
            }
            n++
            if n >= s.length { break }
        }
        if n >= s.length { break }
    }
    // trailing zeros not backed by buckets
    for i := n; i < s.length; i++ {
        buf.WriteByte('0')
    }
    return fmt.Sprintf("<Store length=%d, bits=%s>", s.length, buf.String())
}

// Resize resizes the length of the sequence of bits. If growing, trailing
// bits are set to zero. If shrinking, may be able to free some of the surplus
// backing memory.
func (s *Store) Resize(length int) {
    if length == s.length {
        return
    } else if length > s.length {
        s.Set(length - 1, false)
    } else {
        cropped := s.croppedBuckets()
        if len(cropped) != cap(s.buckets) {
            s.buckets = make([]uint64, len(cropped))
        }
        copy(s.buckets, cropped)
        s.buckets = s.buckets[:cap(s.buckets)]

        // clear trailing data in last bucket
        last := length / 64
        if last > 0 {
            offset := 64 - (length % 64)

            for i := offset; i < 64; i++ {
                s.buckets[last] = s.buckets[last] & (^(1 << offset))
            }
        }

        s.length = length
    }
}

// croppedBuckets returns a subslice of buckets that excludes any buckets that
// are solely trailing zeros.
func (s *Store) croppedBuckets() []uint64 {
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
func(s *Store) Write(w io.Writer) error {
    var err error
    var crc uint64

    write := ks.LiftErrorFunc(func(value uint64) error {
        crc = ks.Checksum64(crc, value)
        return binary.Write(w, binary.LittleEndian, value)
    })

    buckets := s.croppedBuckets()
    stats := (uint64(s.Length()) << 32) + (uint64(len(buckets) << 0))
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
// information by continuing to consume from the reader.
func Read(dest *Store, r io.Reader) error {
    return ks.ErrTODO
}

// Length returns the number of bits stored.
func (s *Store) Length() int {
    // Note: This may be less than or greater than the number
    // of bits actually physically available in buckets.
    return s.length
}

// Push stores a true or false bit at the end of the sequence.
func (s *Store) Push(bit bool) {
    s.Set(s.length, bit)
}

// Pop returns and removes the bit at the end of the sequence.
func (s *Store) Pop() bool {
    result := s.Get(s.length - 1)
    s.Set(s.length - 1, false) // zero unused memory
    s.length--
    return result
}

// Peek returns the bit at the end of the sequence.
func (s *Store) Peek() bool {
    return s.Get(s.length - 1)
}

func fromIndex(index int) (bucket int, offset int) {
    bucket = index / 64
    offset = index % 64
    return
}

// Set sets a bit to true or false at given index, growing the capacity of
// the backing array as necessary if the index exceeds its current size.
// Intermediate values are automatically initialised with false bits.
func (s *Store) Set(index int, bit bool) {
    if index < 0 { panic(ErrRange) }
    if index >= s.length { s.length = index + 1 }

    bucket, offset := fromIndex(index)
    if bucket >= cap(s.buckets) {
        if !bit { return } // trailing zeros are implied
        s.buckets = slices.Grow(s.buckets, 1 + bucket - len(s.buckets))
        s.buckets = s.buckets[:cap(s.buckets)]
    }
    if bit {
        s.buckets[bucket] = s.buckets[bucket] | (1 << offset)
    } else {
        s.buckets[bucket] = s.buckets[bucket] & (^(1 << offset))
    }
}

// Get looks up a bit at a given index, returning true iff it has been set.
// Panics if index is out of range.
func (s *Store) Get(index int) bool {
    if (index < 0) || (index >= s.length) { panic(ErrRange) }
    return s.getFromBucket(fromIndex(index))
}

func (s *Store) getFromBucket(bucket, offset int) bool {
    if bucket < len(s.buckets) {
        return (s.buckets[bucket] & (1 << offset)) != 0
    } else {
        return false // trailing zeros are implied
    }
}

// NextFalse returns the index of the next false bit found after the given
// index. To start at the beginning, start with NextFalse(-1). If the second
// return value is false, then the search has got to the end of the sequence
// without finding any false bits.
func (s *Store) NextFalse(after int) (int, bool) {
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
            if index >= s.length { return -1, false }
            if !s.getFromBucket(i, j) { return index, true }
        }
    }

    return -1, false
}

// NextTrue returns the index of the next true bit found after the given
// index. To start at the beginning, start with NextTrue(-1). If the second
// return value is false, then the search has got to the end of the sequence
// without finding any true bits.
func (s *Store) NextTrue(after int) (int, bool) {
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
            if index >= s.length { return -1, false }
            if s.getFromBucket(i, j) { return index, true }
        }
    }

    return -1, false
}
