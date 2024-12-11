package testdata

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/cmd/go2proto/testdata/testdatapb"
	"github.com/TBD54566975/ftl/internal/model"
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
		RepeatedInt:    []int{1, 2, 3},
		RepeatedMsg:    []*Message{&Message{Time: time.Now()}, &Message{Time: time.Now()}},
		URL:            must.Get(url.Parse("http://127.0.0.1")),
		Key:            model.NewDeploymentKey("echo"),
	}
	pb := model.ToProto()
	fmt.Println(pb)
	data, err := proto.Marshal(pb)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(data, []byte("http://127.0.0.1")), "missing url")
	assert.True(t, bytes.Contains(data, []byte("dpl-echo-")), "missing deployment key")
	assert.True(t, bytes.Contains(data, []byte("bar")), "missing sum type value")
	out := &testdatapb.Root{}
	err = proto.Unmarshal(data, out)
	assert.NoError(t, err)
	assert.Equal(t, pb.String(), out.String())
}
