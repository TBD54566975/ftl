package common

import "encoding/json"

// LoadDatabase loads a "database" from JSON.
func LoadDatabase[T any](data []byte) T {
	var out T
	err := json.Unmarshal(data, &out)
	if err != nil {
		panic(err)
	}
	return out
}
