package duration

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestGoodDurations(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"", 0},
		{"1d", time.Hour * 24},
		{"1h", time.Hour},
		{"1m", time.Minute},
		{"1s", time.Second},
		{"1d1h1m1s", time.Hour*24 + time.Hour + time.Minute + time.Second},
	}
	for _, test := range tests {
		duration, err := Parse(test.input)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, duration)
	}
}

func TestBadDurations(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"1"},
		{"1x"},
		{"1d1d"},
		{"1d1h1m1s1"},
		{"1s5d"},
		{"-1s"},
	}
	for _, test := range tests {
		duration, err := Parse(test.input)
		assert.Error(t, err, test.input)
		assert.Zero(t, duration)
	}
}
