package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type FromOffset int

const (
	FromOffsetBeginning FromOffset = iota
	FromOffsetLatest
)

func (o *FromOffset) Capture(values []string) error {
	switch values[0] {
	case "beginning":
		*o = FromOffsetBeginning
	case "latest":
		*o = FromOffsetLatest
	default:
		return fmt.Errorf("unexpected value %q", values[0])
	}
	return nil
}

func (o FromOffset) String() string {
	switch o {
	case FromOffsetBeginning:
		return "beginning"
	case FromOffsetLatest:
		return "latest"
	default:
		panic("unexpected value")
	}
}

func (o FromOffset) ToProto() schemapb.FromOffset {
	switch o {
	case FromOffsetBeginning:
		return schemapb.FromOffset_FROM_OFFSET_BEGINNING
	case FromOffsetLatest:
		return schemapb.FromOffset_FROM_OFFSET_LATEST
	default:
		panic("unexpected value")
	}
}

//protobuf:7
type MetadataSubscriber struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Topic      *Ref       `parser:"'+' 'subscribe' @@" protobuf:"2"`
	FromOffset FromOffset `parser:"'from' '=' @('beginning'|'latest')" protobuf:"3"`
	DeadLetter bool       `parser:"@'deadletter'?" protobuf:"4"`
}

var _ Metadata = (*MetadataSubscriber)(nil)

func (*MetadataSubscriber) schemaMetadata()          {}
func (m *MetadataSubscriber) schemaChildren() []Node { return []Node{m.Topic} }
func (m *MetadataSubscriber) Position() Position     { return m.Pos }
func (m *MetadataSubscriber) String() string {
	components := []string{
		"+subscribe",
		m.Topic.String(),
	}
	components = append(components, "from="+m.FromOffset.String())
	if m.DeadLetter {
		components = append(components, "deadletter")
	}
	return strings.Join(components, " ")
}

func (m *MetadataSubscriber) ToProto() proto.Message {
	return &schemapb.MetadataSubscriber{
		Pos: posToProto(m.Pos),

		Topic:      m.Topic.ToProto().(*schemapb.Ref),
		FromOffset: m.FromOffset.ToProto(),
		DeadLetter: m.DeadLetter,
	}
}
