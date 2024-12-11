package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestProtoRoundtrip(t *testing.T) {
	p := MustValidate(testSchema).ToProto()
	actual, err := FromProto(p)
	assert.NoError(t, err)
	assert.Equal(t, Normalise(testSchema), Normalise(actual))
}
