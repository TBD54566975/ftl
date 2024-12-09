package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/slices"
)

//protobuf:9
type Topic struct {
	Pos     Position      `parser:"" protobuf:"1,optional"`
	Runtime *TopicRuntime `parser:"" protobuf:"31634,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Export   bool     `parser:"@'export'?" protobuf:"3"`
	Name     string   `parser:"'topic' @Ident" protobuf:"4"`
	Event    Type     `parser:"@@" protobuf:"5"`
}

var _ Decl = (*Topic)(nil)
var _ Symbol = (*Topic)(nil)
var _ Provisioned = (*Topic)(nil)

func (t *Topic) Position() Position { return t.Pos }
func (*Topic) schemaDecl()          {}
func (*Topic) schemaSymbol()        {}
func (t *Topic) schemaChildren() []Node {
	if t.Event == nil {
		return nil
	}
	return []Node{t.Event}
}

func (t *Topic) GetName() string  { return t.Name }
func (t *Topic) IsExported() bool { return t.Export }

func (t *Topic) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(t.Comments))
	if t.Export {
		fmt.Fprint(w, "export ")
	}
	fmt.Fprintf(w, "topic %s %s", t.Name, t.Event)
	return w.String()
}

func (t *Topic) ToProto() proto.Message {
	pb := &schemapb.Topic{
		Pos: posToProto(t.Pos),

		Name:     t.Name,
		Export:   t.Export,
		Event:    TypeToProto(t.Event),
		Comments: t.Comments,
	}
	if t.Runtime != nil {
		pb.Runtime = t.Runtime.ToProto()
	}
	return pb
}

func (t *Topic) GetProvisioned() ResourceSet {
	return ResourceSet{
		{
			Kind:   ResourceTypeTopic,
			Config: &Topic{Name: t.Name},
		},
	}
}

func (t *Topic) ResourceID() string {
	return t.Name
}

func TopicFromProto(t *schemapb.Topic) *Topic {
	return &Topic{
		Pos:     PosFromProto(t.Pos),
		Runtime: TopicRuntimeFromProto(t.Runtime),

		Name:     t.Name,
		Export:   t.Export,
		Event:    TypeFromProto(t.Event),
		Comments: t.Comments,
	}
}

type TopicRuntime struct {
	KafkaBrokers []string `parser:"" protobuf:"1"`
	TopicID      string   `parser:"" protobuf:"2"`
}

func (t *TopicRuntime) ToProto() *schemapb.TopicRuntime {
	return &schemapb.TopicRuntime{
		KafkaBrokers: t.KafkaBrokers,
		TopicId:      t.TopicID,
	}
}

func TopicRuntimeFromProto(t *schemapb.TopicRuntime) *TopicRuntime {
	if t == nil {
		return nil
	}
	return &TopicRuntime{
		KafkaBrokers: t.KafkaBrokers,
		TopicID:      t.TopicId,
	}
}

type TopicRuntimeEvent struct {
	ID      string        `parser:"" protobuf:"1"`
	Payload *TopicRuntime `parser:"" protobuf:"2"`
}

func (t *TopicRuntimeEvent) ToProto() protoreflect.ProtoMessage {
	return &schemapb.TopicRuntimeEvent{
		Id:      t.ID,
		Payload: t.Payload.ToProto(),
	}
}

func TopicRuntimeEventFromProto(t *schemapb.TopicRuntimeEvent) *TopicRuntimeEvent {
	return &TopicRuntimeEvent{
		ID:      t.Id,
		Payload: TopicRuntimeFromProto(t.Payload),
	}
}

func (t *TopicRuntimeEvent) ApplyTo(m *Module) {
	for topic := range slices.FilterVariants[*Topic](m.Decls) {
		if topic.Name == t.ID {
			topic.Runtime = t.Payload
		}
	}
}
