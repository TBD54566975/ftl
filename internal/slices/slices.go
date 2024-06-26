package slices

import (
	"cmp"
	"sort"
)

func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func MapErr[T, U any](slice []T, fn func(T) (U, error)) ([]U, error) {
	result := make([]U, len(slice))
	for i, v := range slice {
		var err error
		result[i], err = fn(v)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func Filter[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// GroupBy groups the elements of a slice by the result of a function.
func GroupBy[T any, K comparable](slice []T, fn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}

func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// AppendOrReplace appends a value to a slice if the slice does not contain a
// value for which the given function returns true. If the slice does contain
// such a value, it is replaced.
func AppendOrReplace[T any](slice []T, value T, fn func(T) bool) []T {
	for i, v := range slice {
		if fn(v) {
			slice[i] = value
			return slice
		}
	}
	return append(slice, value)
}

// Sort returns a sorted clone of slice.
func Sort[T cmp.Ordered](slice []T) []T {
	out := make([]T, len(slice))
	copy(out, slice)
	sort.Slice(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}

func FlatMap[T, U any](slice []T, fn func(T) []U) []U {
	result := make([]U, 0, len(slice))
	for _, v := range slice {
		result = append(result, fn(v)...)
	}
	return result
}

func Find[T any](slice []T, fn func(T) bool) (T, bool) {
	for _, v := range slice {
		if fn(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func FindVariant[T any, U any](slice []U) (T, bool) {
	for _, el := range slice {
		if found, ok := any(el).(T); ok {
			return found, true
		}
	}
	var zero T
	return zero, false
}
