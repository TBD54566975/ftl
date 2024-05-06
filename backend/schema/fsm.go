package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type FSM struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments    []string         `parser:"@Comment*" protobuf:"2"`
	Name        string           `parser:"'fsm' @Ident '{'" protobuf:"3"`
	Start       []*Ref           `parser:"('start' @@)*" protobuf:"4"` // Start states.
	Transitions []*FSMTransition `parser:"('transition' @@)* '}'" protobuf:"5"`
}

func FSMFromProto(pb *schemapb.FSM) *FSM {
	return &FSM{
		Pos:         posFromProto(pb.Pos),
		Name:        pb.Name,
		Start:       slices.Map(pb.Start, RefFromProto),
		Transitions: slices.Map(pb.Transitions, FSMTransitionFromProto),
	}
}

var _ Decl = (*FSM)(nil)
var _ Symbol = (*FSM)(nil)

func (f *FSM) GetName() string    { return f.Name }
func (f *FSM) IsExported() bool   { return false }
func (f *FSM) Position() Position { return f.Pos }
func (f *FSM) schemaDecl()        {}
func (f *FSM) schemaSymbol()      {}

func (f *FSM) String() string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "fsm %s {\n", f.Name)
	for _, s := range f.Start {
		fmt.Fprintf(w, "  start %s\n", s)
	}
	for _, t := range f.Transitions {
		fmt.Fprintf(w, "  transition %s\n", t)
	}
	fmt.Fprintf(w, "}")
	return w.String()
}

func (f *FSM) ToProto() protoreflect.ProtoMessage {
	return &schemapb.FSM{
		Pos:  posToProto(f.Pos),
		Name: f.Name,
		Start: slices.Map(f.Start, func(r *Ref) *schemapb.Ref {
			return r.ToProto().(*schemapb.Ref) //nolint: forcetypeassert
		}),
		Transitions: slices.Map(f.Transitions, func(t *FSMTransition) *schemapb.FSMTransition {
			return t.ToProto().(*schemapb.FSMTransition) //nolint: forcetypeassert
		}),
	}
}

func (f *FSM) schemaChildren() []Node {
	out := make([]Node, 0, len(f.Transitions))
	for _, s := range f.Start {
		out = append(out, s)
	}
	for _, t := range f.Transitions {
		out = append(out, t)
	}
	return out
}

type FSMTransition struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	From     *Ref     `parser:"@@" protobuf:"3"`
	To       *Ref     `parser:"'to' @@" protobuf:"4"`
}

func FSMTransitionFromProto(pb *schemapb.FSMTransition) *FSMTransition {
	return &FSMTransition{
		Pos:  posFromProto(pb.Pos),
		From: RefFromProto(pb.From),
		To:   RefFromProto(pb.To),
	}
}

var _ Node = (*FSMTransition)(nil)

func (f *FSMTransition) Position() Position { return f.Pos }

func (f *FSMTransition) String() string {
	return fmt.Sprintf("%s to %s", f.From, f.To)
}

func (f *FSMTransition) ToProto() protoreflect.ProtoMessage {
	return &schemapb.FSMTransition{
		Pos:  posToProto(f.Pos),
		From: f.From.ToProto().(*schemapb.Ref), //nolint: forcetypeassert
		To:   f.To.ToProto().(*schemapb.Ref),   //nolint: forcetypeassert
	}
}

func (f *FSMTransition) schemaChildren() []Node { return nil }
