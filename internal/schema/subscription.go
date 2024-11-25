package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

//protobuf:10
type Subscription struct {
	Pos     Position             `parser:"" protobuf:"1,optional"`
	Runtime *SubscriptionRuntime `parser:"" protobuf:"31634,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"'subscription' @Ident" protobuf:"3"`
	Topic    *Ref     `parser:"@@" protobuf:"4"`
}

var _ Decl = (*Subscription)(nil)
var _ Symbol = (*Subscription)(nil)

func (s *Subscription) Position() Position { return s.Pos }
func (*Subscription) schemaDecl()          {}
func (*Subscription) schemaSymbol()        {}
func (s *Subscription) schemaChildren() []Node {
	return []Node{s.Topic}
}

func (s *Subscription) GetName() string  { return s.Name }
func (s *Subscription) IsExported() bool { return false }

func (s *Subscription) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(s.Comments))
	fmt.Fprintf(w, "subscription %s %v", s.Name, s.Topic)
	return w.String()
}

func (s *Subscription) ToProto() proto.Message {
	pb := &schemapb.Subscription{ //nolint: forcetypeassert
		Pos: posToProto(s.Pos),

		Name:     s.Name,
		Topic:    s.Topic.ToProto().(*schemapb.Ref),
		Comments: s.Comments,
	}
	if s.Runtime != nil {
		pb.Runtime = s.Runtime.ToProto()
	}
	return pb
}

func SubscriptionFromProto(s *schemapb.Subscription) *Subscription {
	return &Subscription{
		Pos:     PosFromProto(s.Pos),
		Runtime: SubscriptionRuntimeFromProto(s.Runtime),

		Name:     s.Name,
		Topic:    RefFromProto(s.Topic),
		Comments: s.Comments,
	}
}

type SubscriptionRuntime struct {
	KafkaBrokers    []string `parser:"" protobuf:"1"`
	TopicID         string   `parser:"" protobuf:"2"`
	ConsumerGroupID string   `parser:"" protobuf:"3"`
}

func (s *SubscriptionRuntime) ToProto() *schemapb.SubscriptionRuntime {
	return &schemapb.SubscriptionRuntime{
		KafkaBrokers:    s.KafkaBrokers,
		TopicId:         s.TopicID,
		ConsumerGroupId: s.ConsumerGroupID,
	}
}

func SubscriptionRuntimeFromProto(s *schemapb.SubscriptionRuntime) *SubscriptionRuntime {
	if s == nil {
		return nil
	}
	return &SubscriptionRuntime{
		KafkaBrokers:    s.KafkaBrokers,
		TopicID:         s.TopicId,
		ConsumerGroupID: s.ConsumerGroupId,
	}
}
