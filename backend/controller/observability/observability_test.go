package observability

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestLogBucket(t *testing.T) {
	tests := []struct {
		name string
		base int
		num  int
		want string
	}{
		{
			name: "<1",
			base: 8,
			num:  0,
			want: "<1",
		},
		{
			name: "EqualLowEndOfRange",
			base: 8,
			num:  1,
			want: "[1,8)",
		},
		{
			name: "HigherEndOfRange",
			base: 8,
			num:  7,
			want: "[1,8)",
		},
		{
			name: "BigInputNum",
			base: 8,
			num:  16800000,
			want: "[16777216,134217728)",
		},
		{
			name: "Base2",
			base: 2,
			num:  8,
			want: "[8,16)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, logBucket(test.base, int64(test.num)))
		})
	}
}
