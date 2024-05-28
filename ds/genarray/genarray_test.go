package genarray_test

import (
    "fmt"

    "github.com/tawesoft/golib/v2/ds/genarray"
    "github.com/tawesoft/golib/v2/must"
)

func ExampleStore() {
    // declare a generational array Store
    var store genarray.Store[string]

    // insert "hello" and "world" into the Store and keep a reference
    hello := store.Insert("hello")
    world := store.Insert("world")

    // we can retrieve a value using the returned key as a reference
    value, ok := store.Get(hello)
    must.Truef(ok, "failed to lookup hello key in store")
    must.Equalf(value, "hello", "hello key lookup returned unexpected value %q", value)

    // or delete a value...
    err := store.Delete(world)
    must.Equal(err, nil)

    // after being deleted, cannot retrieve again
    value, ok = store.Get(world)
    must.Not(ok)

    // nor delete again
    err = store.Delete(world)
    must.Equal(err, genarray.ErrNotFound)

    // now insert a new element, reusing the space where "world" appeared
    // previously in the Store's backing array.
    everyone := store.Insert("everyone")

    // keys are not equal despite the space being reused
    must.Not(world == everyone)

    // let's check the contents:
    valueIterator := store.Values()
    for {
        value, ok := valueIterator()
        if !ok { break }
        fmt.Println(value)
    }
    fmt.Println(store.Count())

    // Output:
    // hello
    // everyone
    // 2
}
