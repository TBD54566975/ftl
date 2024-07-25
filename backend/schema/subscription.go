package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Subscription struct {
	Pos Position `parser:"" protobuf:"1,optional"`

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
	return &schemapb.Subscription{
		Pos: posToProto(s.Pos),

		Name:     s.Name,
		Topic:    s.Topic.ToProto().(*schemapb.Ref), //nolint: forcetypeassert
		Comments: s.Comments,
	}
}

func SubscriptionFromProto(s *schemapb.Subscription) *Subscription {
	return &Subscription{
		Pos: posFromProto(s.Pos),

		Name:     s.Name,
		Topic:    RefFromProto(s.Topic),
		Comments: s.Comments,
	}
}
