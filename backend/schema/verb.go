package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Verb struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"2"`
	Export   bool       `parser:"@'export'?" protobuf:"3"`
	Name     string     `parser:"'verb' @Ident" protobuf:"4"`
	Request  Type       `parser:"'(' @@ ')'" protobuf:"5"`
	Response Type       `parser:"@@" protobuf:"6"`
	Metadata []Metadata `parser:"@@*" protobuf:"7"`
}

var _ Decl = (*Verb)(nil)
var _ Symbol = (*Verb)(nil)

// VerbKind is the kind of Verb: verb, sink, source or empty.
type VerbKind string

const (
	// VerbKindVerb is a normal verb taking an input and an output of any non-unit type.
	VerbKindVerb VerbKind = "verb"
	// VerbKindSink is a verb that takes an input and returns unit.
	VerbKindSink VerbKind = "sink"
	// VerbKindSource is a verb that returns an output and takes unit.
	VerbKindSource VerbKind = "source"
	// VerbKindEmpty is a verb that takes unit and returns unit.
	VerbKindEmpty VerbKind = "empty"
)

// Kind returns the kind of Verb this is.
func (v *Verb) Kind() VerbKind {
	_, inIsUnit := v.Request.(*Unit)
	_, outIsUnit := v.Response.(*Unit)
	switch {
	case inIsUnit && outIsUnit:
		return VerbKindEmpty

	case inIsUnit:
		return VerbKindSource

	case outIsUnit:
		return VerbKindSink

	default:
		return VerbKindVerb
	}
}

func (v *Verb) Position() Position { return v.Pos }
func (v *Verb) schemaDecl()        {}
func (v *Verb) schemaSymbol()      {}
func (v *Verb) schemaChildren() []Node {
	children := make([]Node, 2+len(v.Metadata))
	children[0] = v.Request
	children[1] = v.Response
	for i, c := range v.Metadata {
		children[i+2] = c
	}
	return children
}

func (v *Verb) GetName() string  { return v.Name }
func (v *Verb) IsExported() bool { return v.Export }

func (v *Verb) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(v.Comments))
	if v.Export {
		fmt.Fprint(w, "export ")
	}
	fmt.Fprintf(w, "verb %s(%s) %s", v.Name, v.Request, v.Response)
	fmt.Fprint(w, indent(encodeMetadata(v.Metadata)))
	return w.String()
}

// AddCall adds a call reference to the Verb.
func (v *Verb) AddCall(verb *Ref) {
	for _, c := range v.Metadata {
		if c, ok := c.(*MetadataCalls); ok {
			c.Calls = append(c.Calls, verb)
			return
		}
	}
	v.Metadata = append(v.Metadata, &MetadataCalls{Calls: []*Ref{verb}})
}

func (v *Verb) GetMetadataIngress() optional.Option[*MetadataIngress] {
	for _, m := range v.Metadata {
		if m, ok := m.(*MetadataIngress); ok {
			return optional.Some(m)
		}
	}
	return optional.None[*MetadataIngress]()
}

func (v *Verb) GetMetadataCronJob() optional.Option[*MetadataCronJob] {
	for _, m := range v.Metadata {
		if m, ok := m.(*MetadataCronJob); ok {
			return optional.Some(m)
		}
	}
	return optional.None[*MetadataCronJob]()
}

func (v *Verb) ToProto() proto.Message {
	return &schemapb.Verb{
		Pos:      posToProto(v.Pos),
		Export:   v.Export,
		Name:     v.Name,
		Comments: v.Comments,
		Request:  TypeToProto(v.Request),
		Response: TypeToProto(v.Response),
		Metadata: metadataListToProto(v.Metadata),
	}
}

func VerbFromProto(s *schemapb.Verb) *Verb {
	return &Verb{
		Pos:      posFromProto(s.Pos),
		Export:   s.Export,
		Name:     s.Name,
		Comments: s.Comments,
		Request:  TypeFromProto(s.Request),
		Response: TypeFromProto(s.Response),
		Metadata: metadataListToSchema(s.Metadata),
	}
}
