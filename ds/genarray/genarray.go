// Package genarray implements a generational array data structure.
//
// A generational array is a contiguous block of memory, like a vanilla array,
// but the array index is extended with an incrementing "generation" that
// allows elements to be safely invalidated and removed, and their space in the
// array reused for future insertions.
//
// Each array slot supports a 64-bit generation count. In rare cases this
// may eventually overflow, raising a panic with [ErrRange] on insertion.
//
// Security model: note that key lookups are not constant-time, and may be
// leaked through timing side-channel attacks. Do not treat keys as secret
// values.
package genarray

import (
    "encoding/binary"
    "errors"
    "io"
    "math"
    "slices"

    "github.com/tawesoft/golib/v2/ds/bitseq"
    "github.com/tawesoft/golib/v2/iter"
    "github.com/tawesoft/golib/v2/ks"
    "github.com/tawesoft/golib/v2/must"
    "github.com/tawesoft/golib/v2/operator"
)

var ErrNotFound = errors.New("not found")
var ErrRange    = errors.New("value out of range")
var ErrConflict = errors.New("key conflict")
var ErrLimit    = errors.New("index exceeds limit")

// A Key uniquely references a value in a Store.
type Key struct { maskedIndex, maskedGeneration uint64 }

// Write encodes a key as an opaque 16-byte value. Provided a Store uses the
// same mask each time, this value is stable across processes, and may be
// (de)serialised to save/restore/transmit a Store.
func (k Key) Write(w io.Writer) error {
    var buf [16]byte
    k.Bytes(buf[:])
    _, err := w.Write(buf[:])
    return err
}

// Bytes encodes a key as an opaque 16-byte value. Provided a Store uses the
// same mask each time, this value is stable across processes, and may be
// (de)serialised to save/restore/transmit a Store.
func (k Key) Bytes(dest []byte) {
    if cap(dest) < 16 { panic(io.ErrShortBuffer) }
    binary.LittleEndian.PutUint64(dest[ 0: 8], k.maskedIndex)
    binary.LittleEndian.PutUint64(dest[ 8:16], k.maskedGeneration)
}

func encodeKey(mask uint64, index int, generation uint64) Key {
    if index < 0 { panic(ErrRange) }
    return Key{uint64(index) & mask, generation & mask}
}

func decodeIndex(mask uint64, key Key) (int, bool) {
    index := key.maskedIndex & mask
    if index > math.MaxInt { return -1, false }
    return int(index), true // never returns a value < 0 on success
}

func decodeGeneration(mask uint64, key Key) uint64 {
    return key.maskedGeneration & mask
}

type pair[ValueT any] struct {
    generation uint64
    value      ValueT
}

// Store is a collection of values, indexed by a unique key. The zero-value
// for a store is a useful value, but see also [Store.Init].
type Store[ValueT any] struct {
    mask        uint64
    generations []uint64
    values      []ValueT
    filled      bitseq.Store // fast lookup for finding gaps
    active      int
    gaps        int
    limit       int
}

// Mask returns the mask value used in the Store. See [Store.Init].
func (s *Store[ValueT]) Mask() uint64 {
    return s.mask
}

// Count returns the number of entries in the Store.
func (s *Store[ValueT]) Count() int {
    return s.active
}

// Init (re)initialises a Store with a mask that guards against programming
// errors. Use a unique (or randomly-generated) mask for each store to avoid a
// key returned from one store being reused incorrectly to access an element in
// another store.
//
// Additionally, limit, if greater than zero, places an upper limit on the
// number of values the Store may hold at any one time. This is helpful to
// guard against excessive memory consumption caused by corrupted, erroneous,
// or malicious keys passed to [Store.Put].
func (s *Store[ValueT]) Init(mask uint64, limit int) {
    s.mask        = mask
    s.generations = nil
    s.values      = nil
    s.filled      = bitseq.Store{}
    s.gaps        = 0
    s.active      = 0
    s.limit       = limit
}

// Grow increases the store's capacity, if necessary, to guarantee space for
// another n elements. After Grow(n), at least n elements can be appended to
// the store without another allocation. This is an optional optimisation.
//
// Panics with [ErrLimit] if the new capacity would exceed the limit set in
// [Store.Init].
func (s *Store[ValueT]) Grow(n int) {
    if n <= s.gaps { return }
    n = n - s.gaps

    if cap(s.generations) + n > s.limit {
        panic(ErrLimit)
    }

    capBefore := cap(s.generations)
    s.generations = slices.Grow(s.generations, n)
    s.values = slices.Grow(s.values, n)
    capAfter := cap(s.generations)
    if capBefore == capAfter { return }
    s.filled.Set(capAfter - 1, false)
    s.gaps += (capAfter - capBefore)
}

// Insert puts a copy of value in the Store, and returns a Key which uniquely
// identifies it for lookup later.
//
// In the unlikely event that the 64-bit generation counter for an entry would
// overflow, panics with ErrRange.
func (s *Store[ValueT]) Insert(value ValueT) Key {
    if s.gaps == 0 {
        // append directly to end of a full store
        index := len(s.generations)
        s.Grow(1)
        s.generations = append(s.generations, 1)
        s.filled.Set(index, true)
        s.gaps--
        s.active++
        return encodeKey(s.mask, index, 1)
    } else {
        // reuse a gap
        index := s.filled.NextFalse(-1)
        generation := s.generations[index]
        if generation == math.MaxUint64 { panic(ErrRange) }
        generation++
        s.generations[index] = generation
        s.values[index] = value
        s.filled.Set(index, true)
        s.gaps--
        s.active++
        return encodeKey(s.mask, index, generation)
    }
}

// Delete removes an entry from the Store, referenced by Key, and overwrites
// the value in the store with the zero value for its type in order to prevent
// dangling references. If not found (or already deleted previously), returns
// ErrNotFound. Once removed, that Key will never again be a valid reference to
// a value, even if the underlying memory in the Store gets reused for a new
// value.
func (s *Store[ValueT]) Delete(key Key) error {
    index, ok := decodeIndex(s.mask, key)
    if !ok || !s.filled.Get(index) {
        return ErrNotFound
    }

    generation := decodeGeneration(s.mask, key)
    if generation == 0 || s.generations[index] != generation {
        return ErrNotFound
    }

    s.values[index] = operator.Zero[ValueT]()
    s.filled.Set(index, false)
    s.gaps++
    s.active--
    must.True(s.active >= 0)
    return nil
}

// Get retrieves a copy of an entry from the Store, referenced by Key. The
// second return value is true iff found.
func (s *Store[ValueT]) Get(key Key) (ValueT, bool) {
    var zero ValueT

    index, ok := decodeIndex(s.mask, key)
    if !ok || !s.filled.Get(index) {
        return zero, false
    }

    generation := decodeGeneration(s.mask, key)
    if generation == 0 || s.generations[index] != generation {
        return zero, false
    }

    return s.values[index], true
}

// Update modifies an existing entry in the Store, referenced by Key. May
// return ErrNotFound if the key does not reference a valid current entry.
func (s *Store[ValueT]) Update(key Key, value ValueT) error {
    index, ok := decodeIndex(s.mask, key)
    if !ok || !s.filled.Get(index) {
        return ErrNotFound
    }

    return ks.ErrTODO
}

// Put puts a copy of value in the Store at a location and with a generation
// count specified by a previously returned Key. This function is designed for
// restoring contents of a Store e.g. from a serialised form on disk.
//
// If an entry already exists at that location (whatever its generation; even
// if it has been deleted), this function panics with [ErrConflict].
func (s *Store[ValueT]) Put(key Key, value ValueT) {
    index, ok := decodeIndex(s.mask, key)
    if !ok || s.filled.Get(index) {
        panic(ErrConflict)
    }

    panic(ks.ErrTODO)
    // growTo() // TODO
}

// Keys returns an iterator function that generates each stored key. The order
// of iteration is not defined. It is not safe to mutate the Store during this
// iteration.
func (s *Store[ValueT]) Keys() func()(Key, bool) {
    current := -1
    return func() (Key, bool) {
        idx, ok := s.filled.NextTrue(current)
        if !ok { return Key{}, false }
        current = idx
        key := encodeKey(s.mask, idx, s.generations[idx])
        return key, true
    }
}

// Values returns an iterator function that generates each stored value. The
// order of iteration is not defined. It is not safe to mutate the Store during
// this iteration.
func (s *Store[ValueT]) Values() func()(ValueT, bool) {
    current := -1
    return func() (ValueT, bool) {
        idx, ok := s.filled.NextTrue(current)
        if !ok { return operator.Zero[ValueT](), false }
        current = idx
        return s.values[idx], true
    }
}

// Pairs returns an iterator function that generates each stored (Key, Value)
// pair. The order of iteration is not defined. It is not safe to mutate the
// Store during this iteration.
func (s *Store[ValueT]) Pairs() func()(iter.Pair[Key, ValueT], bool) {
    current := -1
    return func() (pair iter.Pair[Key, ValueT], ok bool) {
        idx, ok := s.filled.NextTrue(current)
        if !ok { return iter.Pair[Key, ValueT]{}, false }
        current = idx
        key := encodeKey(s.mask, idx, s.generations[idx])
        return iter.Pair[Key, ValueT]{Key: key, Value: s.values[idx]}, true
    }
}
