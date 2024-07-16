package indexes_test

import (
    "fmt"

    "github.com/tawesoft/golib/v2/ds/matrix/indexes"
)

func ExampleForwards() {
    iter := indexes.Forwards(3, 3, 2)
    var idx [3]int
    for iter(idx[:]) {
        fmt.Printf("%v\n", idx)
    }

    // Output:
    // [0 0 0]
    // [1 0 0]
    // [2 0 0]
    // [0 1 0]
    // [1 1 0]
    // [2 1 0]
    // [0 2 0]
    // [1 2 0]
    // [2 2 0]
    // [0 0 1]
    // [1 0 1]
    // [2 0 1]
    // [0 1 1]
    // [1 1 1]
    // [2 1 1]
    // [0 2 1]
    // [1 2 1]
    // [2 2 1]
}

func ExampleBackwards() {
    iter := indexes.Backwards(3, 3, 2)
    var idx [3]int
    for iter(idx[:]) {
        fmt.Printf("%v\n", idx)
    }

    // Output:
    // [2 2 1]
    // [1 2 1]
    // [0 2 1]
    // [2 1 1]
    // [1 1 1]
    // [0 1 1]
    // [2 0 1]
    // [1 0 1]
    // [0 0 1]
    // [2 2 0]
    // [1 2 0]
    // [0 2 0]
    // [2 1 0]
    // [1 1 0]
    // [0 1 0]
    // [2 0 0]
    // [1 0 0]
    // [0 0 0]
}

func ExampleRange_Forwards() {
    iter := indexes.NewRange(
        []int{1, 2},
        []int{4, 4},
    ).Forwards()

    var idx [2]int
    for iter(idx[:]) {
        fmt.Printf("%v\n", idx)
    }

    // Output:
    // [1 2]
    // [2 2]
    // [3 2]
    // [1 3]
    // [2 3]
    // [3 3]
}

func ExampleRange_Backwards() {
    iter := indexes.NewRange(
        []int{1, 2},
        []int{4, 4},
    ).Backwards()

    var idx [2]int
    for iter(idx[:]) {
        fmt.Printf("%v\n", idx)
    }

    // Output:
    // [3 3]
    // [2 3]
    // [1 3]
    // [3 2]
    // [2 2]
    // [1 2]
}

func ExampleRange_Contains() {
    r := indexes.NewRange(
        []int{1, 2},
        []int{4, 4},
    )

    test := func(indexes ... int) {
        fmt.Printf("range%v contains %v? %t\n",
            r, indexes, r.Contains(indexes...))
    }
    test(0, 0)
    test(1, 2)
    test(3, 3)
    test(3, 3, 0)
    test(3, 3, 1)
    test(3, 3, 0, 0, 0)
    test(3, 3, 0, 0, 1)
    test(-1, -2)

    // Output:
    // range{[1 2] [4 4]} contains [0 0]? false
    // range{[1 2] [4 4]} contains [1 2]? true
    // range{[1 2] [4 4]} contains [3 3]? true
    // range{[1 2] [4 4]} contains [3 3 0]? true
    // range{[1 2] [4 4]} contains [3 3 1]? false
    // range{[1 2] [4 4]} contains [3 3 0 0 0]? true
    // range{[1 2] [4 4]} contains [3 3 0 0 1]? false
    // range{[1 2] [4 4]} contains [-1 -2]? false
}

func ExampleReorder() {
    iter := indexes.Forwards(2, 4)
    rotate := indexes.Reorder(1, 0)
    var idx [2]int
    for indexes.Map(rotate, iter)(idx[:]) {
        fmt.Printf("%v\n", idx)
    }

    // Output:
    // [0 0]
    // [0 1]
    // [1 0]
    // [1 1]
    // [2 0]
    // [2 1]
    // [3 0]
    // [3 1]
}

func ExampleFilter() {
    iter := indexes.Forwards(5, 7)
    var idx [2]int
    even := func(offsets []int) bool {
        for i := 0; i < len(offsets); i++ {
            if offsets[i] % 2 != 0 { return false }
        }
        return true
    }
    for indexes.Filter(even, iter)(idx[:]) {
        fmt.Printf("%v\n", idx)
    }

    // Output:
    // [0 0]
    // [2 0]
    // [4 0]
    // [0 2]
    // [2 2]
    // [4 2]
    // [0 4]
    // [2 4]
    // [4 4]
    // [0 6]
    // [2 6]
    // [4 6]
}
