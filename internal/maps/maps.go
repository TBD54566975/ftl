package maps

func FromSlice[K comparable, V any, T any](slice []T, kv func(el T) (K, V)) map[K]V {
	out := make(map[K]V, len(slice))
	for _, el := range slice {
		k, v := kv(el)
		out[k] = v
	}
	return out
}

// MapValues transforms a map[X]Y to a map[X]Z using the provided transformation function.
func MapValues[X comparable, Y any, Z any](input map[X]Y, transform func(X, Y) Z) map[X]Z {
	output := make(map[X]Z, len(input))
	for k, v := range input {
		output[k] = transform(k, v)
	}
	return output
}
