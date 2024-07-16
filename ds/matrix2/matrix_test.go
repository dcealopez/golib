package matrix2_test

import (
    "fmt"
    "testing"

    "github.com/tawesoft/golib/v2/ds/matrix/indexes"
    "github.com/tawesoft/golib/v2/fun/slices"
)

func FuzzInterface(f *testing.F) {
    implementations := map[string](matrix2.Constructor[string]){
        "grid": matrix2.NewGrid[string],
    }

    // each test case is an encoding of dimensionality and each (x,y,z,w,...)
    // dimension.
    f.Add([]byte{
        3, 1, 2, 3,         // 3D, 1x2x3
        3, 1, 2, 2,         // 3D, 1x2x2
        3, 1, 2, 3,         // 3D, 1x2x3
        2, 4, 4,            // 2D, 4x4
        3, 4, 4, 2,         // 3D, 4x4x2
        5, 2, 2, 2, 2, 2,   // 5D, 2x2x2x2x2
        3, 3, 4, 2,         // 3D, 3x4x2
        3, 4, 4, 2,         // 3D, 4x4x2
        2, 4, 4,            // 2D, 4x4
    })
    f.Add([]byte{
        2, 2, 2, // 2D, 2x2
        2, 5, 2, // 2D, 5x2
    })

    // set assigns a string-encoding of the coordinate in each dimension
    set := func(m matrix2.Interface[string], lengths ... int) {
        iter := indexes.Forwards(lengths...)
        dest := make([]int, len(lengths))
        for iter(dest) {
            s := fmt.Sprintf("%v", dest)
            m.SetN(s, dest...)
        }
    }

    log := func(t *testing.T, name string, m matrix2.Interface[string]) {
        t.Logf("---matrix<%s>%v---", name, m.Dimensions())
        iter := indexes.Forwards(m.Dimensions()...)
        dest := make([]int, m.Dimensionality())
        for iter(dest) {
            t.Logf("%v: %q", dest, m.GetN(dest...))
        }
        t.Logf("---end---")
    }

    // check
    check := func(t *testing.T, name string, m matrix2.Interface[string], lengths ... int) {
        iter := indexes.Forwards(lengths...)
        dest := make([]int, len(lengths))
        outer: for iter(dest) {
            expected := fmt.Sprintf("%v", dest)
            for i := 0; i < len(dest); i++ {
                if i > m.Dimensionality() && (dest[i] != 0) { continue outer }
                if dest[i] >= lengths[i] { continue outer }
                if dest[i] >= m.DimensionN(i) { continue outer }
            }
            // fmt.Printf("%v %v %v\n", indexes, lengths, m.Dimensions())
            actual := m.GetN(dest...)
            if actual != expected {
                log(t, name, m)
                t.Errorf("%q matrix(%v; previously %v): access %v: expected %v, got %v",
                    name, m.Dimensions(), lengths, dest, expected, actual)
                break
            }
        }
    }

    f.Fuzz(func(t *testing.T, bdimensions []byte) {
        if len(bdimensions) == 0 { return }
        if len(bdimensions) > 50 { return }
        dimensions := slices.Map(func(x byte) int { return int(x) }, bdimensions)
        dimensionsCopy := dimensions
        for name, constructor := range implementations {
            dimensions = dimensionsCopy
            resizing := false
            var m matrix2.Interface[string]
            for {
                if len(dimensions) < 2 { return }
                dimension := dimensions[0]
                dimensions = dimensions[1:]
                if (dimension < 1) || (dimension > 8) { return }
                if len(dimensions) < dimension { return }
                lengths := append([]int{}, dimensions[0:dimension]...)
                dimensions = dimensions[dimension:]
                for i := 0; i < dimension; i++ {
                    length := lengths[i]
                    if length < 1 { return }
                    if length > 16 { return }
                }
                volume := slices.Reduce(1, func(a, b int) int { return a * b }, lengths)
                if volume == 0 { return }
                if volume > 1024 { return }

                if resizing == false {
                    t.Logf("CREATE")
                    m = matrix2.NewN(constructor, lengths...)
                    set(m, lengths...)
                    check(t, name, m, lengths...)
                    resizing = true
                } else {
                    t.Logf("RESIZE")
                    oldLengths := append([]int{}, m.Dimensions()...)
                    m.ResizeN(lengths...)
                    check(t, name, m, oldLengths...)
                    set(m, lengths...)
                    check(t, name, m, lengths...)
                }
                // 1/3rd the time
                if lengths[len(lengths)-1] % 3 == 0 {
                    m.Compact()
                    check(t, name, m, lengths...)
                }
            }
        }
    })
}
