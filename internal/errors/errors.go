package errors

import "errors"

// UnwrapAll recursively unwraps all errors in err, including all intermediate errors.
//
//nolint:errorlint
func UnwrapAll(err error) []error {
	out := []error{}
	if inner, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range inner.Unwrap() {
			out = append(out, UnwrapAll(e)...)
		}
		return out
	}
	if inner, ok := err.(interface{ Unwrap() error }); ok && inner.Unwrap() != nil {
		out = append(out, UnwrapAll(inner.Unwrap())...)
	}
	out = append(out, err)
	return out
}

// Innermost returns true if err cannot be further unwrapped.
//
//nolint:errorlint
func Innermost(err error) bool {
	if err, ok := err.(interface{ Unwrap() []error }); ok && len(err.Unwrap()) > 0 {
		return false
	}
	if err, ok := err.(interface{ Unwrap() error }); ok && err.Unwrap() != nil {
		return false
	}
	return true
}

func Join(errs ...error) error { return errors.Join(errs...) }

func New(text string) error { return errors.New(text) }

func As(err error, target interface{}) bool { return errors.As(err, target) }

func Is(err, target error) bool { return errors.Is(err, target) }

func Unwrap(err error) error { return errors.Unwrap(err) }
