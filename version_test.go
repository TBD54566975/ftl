package ftl

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestIsVersionAtLeastMin(t *testing.T) {
	tests := []struct {
		v          string
		minVersion string
		want       bool
	}{
		// Test case for minFTLVersion being undefined
		{"1.2.3", "", true},

		// Test cases for !IsRelease
		{"dev", "dev", true},
		{"dev", "1.2.3", true},
		{"1.2.3", "dev", true},
		{"1.2.3", "1.2", true},
		{"2.0", "1.2.3", true},
		{"a.b.c", "1.2.3", true},

		// Test cases for comparator loop
		{"1.2.3", "1.2.3", true},
		{"2.2.3", "1.2.3", true},
		{"1.3.3", "1.2.3", true},
		{"1.2.4", "1.2.3", true},
		{"1.2.3", "2.2.3", false},
		{"1.2.3", "1.3.3", false},
		{"1.2.3", "1.2.4", false},
	}
	for _, test := range tests {
		got, err := IsVersionAtLeastMin(test.v, test.minVersion)
		assert.Equal(t, test.want, got)
		assert.NoError(t, err)
	}
}
