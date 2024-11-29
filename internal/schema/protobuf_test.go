package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

func TestProtoRoundtrip(t *testing.T) {
	p := MustValidate(testSchema).ToProto()
	actual, err := FromProto(p.(*schemapb.Schema))
	assert.NoError(t, err)
	assert.Equal(t, Normalise(testSchema), Normalise(actual))
}
