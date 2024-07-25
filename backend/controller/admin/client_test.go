package admin

import (
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestIsEndpointLocal(t *testing.T) {
	tests := []struct {
		Name     string
		Endpoint string
		Want     bool
	}{
		{
			Name:     "DefaultLocalhost",
			Endpoint: "http://localhost:8892",
			Want:     true,
		},
		{
			Name:     "NumericLocalhost",
			Endpoint: "http://127.0.0.1:8892",
			Want:     true,
		},
		{
			Name:     "TooLow",
			Endpoint: "http://126.255.255.255:8892",
			Want:     false,
		},
		{
			Name:     "TooHigh",
			Endpoint: "http://128.0.0.1:8892",
			Want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			u, err := url.Parse(test.Endpoint)
			assert.NoError(t, err)
			got, err := isEndpointLocal(u)
			assert.NoError(t, err)
			assert.Equal(t, got, test.Want)
		})
	}
}
