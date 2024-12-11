package observability

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestLogBucket(t *testing.T) {
	tests := []struct {
		name string
		base int
		min  optional.Option[int]
		max  optional.Option[int]
		num  int
		want string
	}{
		// Without demarcating min/max buckets
		{
			name: "<1",
			base: 8,
			min:  optional.None[int](),
			max:  optional.None[int](),
			num:  0,
			want: "<1",
		},
		{
			name: "EqualLowEndOfRange",
			base: 8,
			min:  optional.None[int](),
			max:  optional.None[int](),
			num:  1,
			want: "[1,8)",
		},
		{
			name: "HigherEndOfRange",
			base: 8,
			min:  optional.None[int](),
			max:  optional.None[int](),
			num:  7,
			want: "[1,8)",
		},
		{
			name: "BigInputNum",
			base: 8,
			min:  optional.None[int](),
			max:  optional.None[int](),
			num:  16800000,
			want: "[16777216,134217728)",
		},
		{
			name: "Base2",
			base: 2,
			min:  optional.None[int](),
			max:  optional.None[int](),
			num:  8,
			want: "[8,16)",
		},

		// With min/max buckets
		{
			name: "LessThanMin",
			base: 2,
			min:  optional.Some(2),
			max:  optional.None[int](),
			num:  3,
			want: "<4",
		},
		{
			name: "EqualToMax",
			base: 2,
			min:  optional.None[int](),
			max:  optional.Some(2),
			num:  4,
			want: ">=4",
		},
		{
			name: "EqualToMaxWhenMinMaxEqual",
			base: 2,
			min:  optional.Some(2),
			max:  optional.Some(2),
			num:  4,
			want: ">=4",
		},
		{
			name: "GreaterThanMax",
			base: 2,
			min:  optional.Some(2),
			max:  optional.Some(2),
			num:  4000,
			want: ">=4",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, LogBucket(test.base, int64(test.num), test.min, test.max))
		})
	}
}
