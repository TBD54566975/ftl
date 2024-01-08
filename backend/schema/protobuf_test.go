package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

func TestProtoRoundtrip(t *testing.T) {
	p := MustValidate(schema).ToProto()
	actual, err := FromProto(p.(*schemapb.Schema))
	assert.NoError(t, err)
	assert.Equal(t, Normalise(schema), Normalise(actual))
}
