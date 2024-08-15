package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type FSM struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments    []string         `parser:"@Comment*" protobuf:"2"`
	Name        string           `parser:"'fsm' @Ident " protobuf:"3"`
	Metadata    []Metadata       `parser:"@@*" protobuf:"6"`
	Start       []*Ref           `parser:"'{' ('start' @@)*" protobuf:"4"` // Start states.
	Transitions []*FSMTransition `parser:"('transition' @@)* '}'" protobuf:"5"`
}

func FSMFromProto(pb *schemapb.FSM) *FSM {
	return &FSM{
		Pos:         posFromProto(pb.Pos),
		Name:        pb.Name,
		Start:       slices.Map(pb.Start, RefFromProto),
		Transitions: slices.Map(pb.Transitions, FSMTransitionFromProto),
		Metadata:    metadataListToSchema(pb.Metadata),
	}
}

var _ Decl = (*FSM)(nil)
var _ Symbol = (*FSM)(nil)

// TerminalStates returns the terminal states of the FSM.
func (f *FSM) TerminalStates() []*Ref {
	var out []*Ref
	all := map[string]struct{}{}
	in := map[string]struct{}{}
	for _, t := range f.Transitions {
		all[t.From.String()] = struct{}{}
		all[t.To.String()] = struct{}{}
		in[t.From.String()] = struct{}{}
	}
	for key := range all {
		if _, ok := in[key]; !ok {
			ref, err := ParseRef(key)
			if err != nil {
				panic(err) // key must be valid
			}
			out = append(out, ref)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].String() < out[j].String()
	})
	return out
}

// NextState returns the next state, if any, given the current state and event.
//
// If currentState is None, the instance has not started.
func (f *FSM) NextState(sch *Schema, currentState optional.Option[RefKey], event Type) optional.Option[*Ref] {
	verb := Verb{}
	curState, ok := currentState.Get()
	if !ok {
		for _, start := range f.Start {
			// This shouldn't happen, but if it does, we'll just return false.
			if err := sch.ResolveToType(start, &verb); err != nil {
				return optional.None[*Ref]()
			}
			if verb.Request.Equal(event) {
				return optional.Some(start)
			}
		}
		return optional.None[*Ref]()
	}
	for _, transition := range f.Transitions {
		if transition.From.ToRefKey() != curState {
			continue
		}
		// This shouldn't happen, but if it does we'll just return false.
		if err := sch.ResolveToType(transition.To, &verb); err != nil {
			return optional.None[*Ref]()
		}
		if verb.Request.Equal(event) {
			return optional.Some(transition.To)
		}
	}
	return optional.None[*Ref]()
}

func (f *FSM) GetName() string    { return f.Name }
func (f *FSM) IsExported() bool   { return false }
func (f *FSM) Position() Position { return f.Pos }
func (f *FSM) schemaDecl()        {}
func (f *FSM) schemaSymbol()      {}

func (f *FSM) String() string {
	w := &strings.Builder{}
	if len(f.Metadata) == 0 {
		fmt.Fprintf(w, "fsm %s {\n", f.Name)
	} else {
		fmt.Fprintf(w, "fsm %s", f.Name)
		fmt.Fprint(w, indent(encodeMetadata(f.Metadata)))
		fmt.Fprintf(w, "\n{\n")
	}
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
		Metadata: metadataListToProto(f.Metadata),
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
	for _, m := range f.Metadata {
		out = append(out, m)
	}
	return out
}

type FSMTransition struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	From     *Ref     `parser:"@@" protobuf:"3,optional"`
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
