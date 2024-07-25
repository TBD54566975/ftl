package maps

func FromSlice[K comparable, V any, T any](slice []T, kv func(el T) (K, V)) map[K]V {
	out := make(map[K]V, len(slice))
	for _, el := range slice {
		k, v := kv(el)
		out[k] = v
	}
	return out
}
