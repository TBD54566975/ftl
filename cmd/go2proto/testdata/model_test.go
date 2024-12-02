package testdata

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/cmd/go2proto/testdata/testdatapb"
)

func TestModel(t *testing.T) {
	intv := 1
	model := Root{
		Int:            1,
		String:         "foo",
		MessagePtr:     &Message{Time: time.Now()},
		Enum:           EnumA,
		SumType:        &SumTypeA{A: "bar"},
		OptionalInt:    2,
		OptionalIntPtr: &intv,
		OptionalMsg:    &Message{Time: time.Now()},
	}
	pb := model.ToProto()
	data, err := proto.Marshal(pb)
	assert.NoError(t, err)
	out := &testdatapb.Root{}
	err = proto.Unmarshal(data, out)
	assert.NoError(t, err)
	assert.Equal(t, pb.String(), out.String())
}
