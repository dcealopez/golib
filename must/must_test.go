package must_test

import (
    "errors"
    "fmt"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/tawesoft/golib/v2/must"
)

func TestCatchFunc(t *testing.T) {
    {
        failingFunctionWithString := func() string {
            panic("oops")
            return "something"
        }

        x, err := must.Try[string](failingFunctionWithString)()
        assert.Equal(t, "", x)
        assert.Error(t, err)
        assert.Equal(t, nil, errors.Unwrap(err))
    }

    {
        failingFunctionWithError := func() string {
            panic(fmt.Errorf("oops"))
            return "something"
        }

        x, err := must.Try[string](failingFunctionWithError)()
        assert.Equal(t, "", x)
        assert.Error(t, err)
        assert.Equal(t, "oops", err.Error())
    }
}

func TestMust(t *testing.T) {
    successfulFunction := func() (string, error) {
        return "success", nil
    }

    assert.NotPanics(t, func() {
        must.Result(successfulFunction())
    })

    failingFunction := func() (string, error) {
        return "", fmt.Errorf("oops")
    }

    assert.Panics(t, func() {
        must.Result(failingFunction())
    })
}

func TestMustFunc(t *testing.T) {
    assert.Panics(t, func() {
        opener := func() (*os.File, error) {
            return os.Open("doesnotexist")
        }
        f := must.Func(opener)()
        f.Close()
    })

    successfulFunction := func() (string, error) {
        return fmt.Sprintf("%d", 123), nil
    }

    f := must.Func(successfulFunction)
    assert.Equal(t, "123", f())
}
