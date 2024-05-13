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
// Security model: note that keys are predictable and keys may be leaked
// through timing side-channel attacks. Do not treat these keys as secret
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
    "github.com/tawesoft/golib/v2/must"
    "github.com/tawesoft/golib/v2/operator"
)

var ErrNotFound = errors.New("not found")
var ErrRange    = errors.New("value out of range")
var ErrLimit    = errors.New("index exceeds limit")
var ErrConflict = errors.New("key conflict")

// A Key uniquely references a value in a Store.
type Key struct { index, generation uint64 }

// Write encodes and writes a key as an opaque 16-byte value. Provided a Store
// uses the same mask each time, this value is stable across different
// processes.
func (k Key) Write(w io.Writer) error {
    var buf [16]byte
    k.Bytes(buf[:])
    _, err := w.Write(buf[:])
    return err
}

// ReadKey Read reads and decodes a key from an opaque 16-byte value. Provided a Store
// uses the same mask each time, this value is stable across different
// processes.
func ReadKey(r io.Reader) (Key, error) {
    var buf [16]byte
    if _, err := io.ReadFull(r, buf[:]); err != nil {
        return Key{}, err
    }
    return KeyFromBytes(buf[:]), nil
}

// Bytes encodes a key as an opaque 16-byte value. Provided a Store uses the
// same mask each time, this value is stable across different processes.
//
// If dest is not large enough to receive 16 bytes, panics with
// [io.ErrShortBuffer].
func (k Key) Bytes(dest []byte) {
    if cap(dest) < 16 { panic(io.ErrShortBuffer) }
    binary.LittleEndian.PutUint64(dest[ 0: 8], k.index)
    binary.LittleEndian.PutUint64(dest[ 8:16], k.generation)
}

// KeyFromBytes decodes a key from an opaque 16-byte value. Provided a Store uses the
// same mask each time, this value is stable across different processes.
//
// If src is not at least 16 bytes long, panics with [io.ErrShortBuffer].
func KeyFromBytes(src []byte) Key {
    if len(src) < 16 { panic(io.ErrShortBuffer) }
    idx := binary.LittleEndian.Uint64(src[ 0: 8])
    gen := binary.LittleEndian.Uint64(src[ 8:16])
    return Key{idx, gen}
}

func encodeKey(index int, generation uint64) Key {
    if index < 0 { panic(ErrRange) }
    return Key{uint64(index), generation}
}

func decodeIndex(key Key) (int, bool) {
    index := key.index
    if index > math.MaxInt { return -1, false }
    return int(index), true // never returns a value < 0 on success
}

func lookup[ValueT any](s *Store[ValueT], key Key) (index int, ok bool) {
    index, ok = decodeIndex(key)
    if !ok || !s.filled.Get(index) {
        return -1, false
    }

    generation := key.generation
    if generation == 0 || s.generations[index] != generation {
        return -1, false
    }

    return index, true
}

// Store is a collection of values, indexed by a unique key. The zero-value
// for a store is a useful value, but see also [Store.Init].
type Store[ValueT any] struct {
    generations []uint64
    values      []ValueT
    filled      bitseq.Store // fast lookup for finding gaps
    active      int
    gaps        int
}

// Count returns the number of values currently in the Store.
func (s *Store[ValueT]) Count() int {
    return s.active
}

// Clear (re)initialises a Store so that it is empty and any backing storage
// is released.
func (s *Store[ValueT]) Clear() {
    s.generations = nil
    s.values      = nil
    s.filled      = bitseq.Store{}
    s.gaps        = 0
    s.active      = 0
}

// ReadKeys (re)initialises a store from a binary serialisation, clearing
// its current contents and repopulating it with keys referencing zero values.
//
// Limit, if greater than zero, sets an upper limit on the size (in number of
// elements) of the backing array. A small but maliciously crated input could
// otherwise consume a large amount of memory by encoding sparsely distributed
// keys.
//
// It is left to the caller to deserialise values and associate them with their
// matching key using [Store.Update].
//
// This function reads until EOF. [io.ReadLimiter] may be useful.
//
// The return value, if not nil, may be [ErrLimit], [ErrConflict] if there
// is a duplicate key, or may represent an [io] read error.
func (s *Store[ValueT]) ReadKeys(r io.Reader, limit int) error {
    s.Clear()

    for {
        key, err := ReadKey(r)
        if (err != nil) && errors.Is(err, io.EOF) { return nil }
        if err != nil { return err }

        index, ok := decodeIndex(key)
        if (limit > 0) && (index > limit) { return ErrLimit }

        generation := key.generation
        if !ok { return ErrLimit }

        if index > cap(s.generations) {
            s.Grow(cap(s.generations) - index)
        }

        s.generations[index] = generation
        s.filled.Set(index, true)
        s.active++
        // TODO count gaps between 0 and cap(s.generations)
    }
}

// WriteKeys writes a binary serialisation of a Store's keys that can later
// be deserialised and associated with values.
//
// It is left to the caller to serialise the values themselves. Note that
// because [Store.Keys] and [Store.Values] iterate in the same order, it is
// not strictly necessary to store redundant keys with the serialised values.
//
// The return value, if not nil, may represent an [io] write error.
func (s *Store[ValueT]) WriteKeys(w io.Writer) error {
    for i := 0; i < len(s.generations); i++ {
        key := encodeKey(i, s.generations[i])
        if err := key.Write(w); err != nil {
            return err
        }
    }
    return nil
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

    capBefore := cap(s.generations)
    s.generations = slices.Grow(s.generations, n)
    s.generations = s.generations[:cap(s.generations)]
    s.values = slices.Grow(s.values, n)
    s.values = s.values[:cap(s.values)]
    capAfter := cap(s.generations)
    if capBefore == capAfter { return }
    s.filled.Set(capAfter - 1, false)
    s.gaps += (capAfter - capBefore)
}

// Insert puts a copy of value in the Store, and returns a Key which uniquely
// identifies it for lookup later.
//
// In the unlikely event that the 64-bit generation counter for an entry would
// overflow, or in the case that the limit set in [Store.Init] is exceeded,
// panics with ErrRange or ErrLimit.
func (s *Store[ValueT]) Insert(value ValueT) Key {
    if s.gaps == 0 {
        // append directly to end of a full store
        index := cap(s.generations)
        s.Grow(1) // may raise ErrLimit
        s.generations[index] = 1
        s.values[index] = value
        s.filled.Set(index, true)
        s.gaps--
        s.active++
        return encodeKey(index, 1)
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
        return encodeKey(index, generation)
    }
}

// Delete removes an entry from the Store, referenced by Key, and overwrites
// the value in the store with the zero value for its type in order to prevent
// dangling references. If not found (or already deleted previously), returns
// ErrNotFound. Once removed, that Key will never again be a valid reference to
// a value, even if the underlying memory in the Store gets reused for a new
// value.
func (s *Store[ValueT]) Delete(key Key) error {
    index, ok := decodeIndex(key)
    if !ok || !s.filled.Get(index) {
        return ErrNotFound
    }

    generation := key.generation
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

// Contains returns true iff the key is a valid reference to a current value.
func (s *Store[ValueT]) Contains(key Key) bool {
    _, ok := lookup(s, key)
    return ok
}

// Get retrieves a copy of a value from the Store, referenced by Key. The
// second return value is true iff found.
func (s *Store[ValueT]) Get(key Key) (ValueT, bool) {
    if index, ok := lookup(s, key); ok {
        return s.values[index], true
    } else {
        var zero ValueT
        return zero, false
    }
}

// Update modifies an existing value in the Store, referenced by Key. May
// return ErrNotFound if the key does not reference a valid current entry.
// Otherwise, returns nil.
func (s *Store[ValueT]) Update(key Key, value ValueT) error {
    if index, ok := lookup(s, key); ok {
        s.values[index] = value
        return nil
    } else {
        return ErrNotFound
    }
}

// Keys returns an iterator function that generates each stored key. The order
// of iteration is not defined, except that [Store.Keys], [Store.Values] and
// [Store.Pairs] produce values in the same order. It is not safe to mutate the
// Store during this iteration.
func (s *Store[ValueT]) Keys() func()(Key, bool) {
    current := -1
    return func() (Key, bool) {
        idx, ok := s.filled.NextTrue(current)
        if !ok { return Key{}, false }
        current = idx
        key := encodeKey(idx, s.generations[idx])
        return key, true
    }
}

// Values returns an iterator function that generates each stored value. The
// order of iteration is not defined, except that [Store.Keys], [Store.Values]
// and [Store.Pairs] produce values in the same order. It is not safe to mutate
// the Store during this iteration.
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
// pair. The order of iteration is not defined, except that [Store.Keys],
// [Store.Values] and [Store.Pairs] produce values in the same order. It is not
// safe to mutate the Store during this iteration.
func (s *Store[ValueT]) Pairs() func()(iter.Pair[Key, ValueT], bool) {
    current := -1
    return func() (pair iter.Pair[Key, ValueT], ok bool) {
        idx, ok := s.filled.NextTrue(current)
        if !ok { return iter.Pair[Key, ValueT]{}, false }
        current = idx
        key := encodeKey(idx, s.generations[idx])
        return iter.Pair[Key, ValueT]{Key: key, Value: s.values[idx]}, true
    }
}
