package simple_test

import (
	"fmt"
	"testing"
)

type TestingError string

// T wraps testing.TB, trapping calls to Fatalf et al and panicking. The panics are caught by
type T struct {
	testing.TB
}

func (t T) Fatal(args ...interface{}) {
	panic(TestingError(fmt.Sprint(args...)))
}

func (t T) Fatalf(format string, args ...interface{}) {
	panic(TestingError(fmt.Sprintf(format, args...)))
}

func (t T) Error(args ...interface{}) {
	panic(TestingError(fmt.Sprint(args...)))
}

func (t T) Errorf(format string, args ...interface{}) {
	panic(TestingError(fmt.Sprintf(format, args...)))
}

func (t T) FailNow() {
	panic(TestingError("FailNow called"))
}

func (t T) Fail() {
	panic(TestingError("Fail called"))
}
