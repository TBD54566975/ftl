package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

func TestProtoRoundtrip(t *testing.T) {
	p := schema.ToProto()
	actual := Normalise(ProtoToSchema(p.(*pschema.Schema))) //nolint:forcetypeassert
	assert.Equal(t, schema, actual)
}
