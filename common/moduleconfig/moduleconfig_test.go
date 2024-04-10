package moduleconfig

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl"
)

func TestLoadModuleConfig(t *testing.T) {
	tests := []struct {
		v       string
		wantErr bool
	}{
		{"dev", false},
		{"1.2.4", false},
		{"0.0.4", true},
	}
	for _, test := range tests {
		ftl.Version = test.v
		_, err := LoadModuleConfig("testdata/projects/sample")
		if test.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
