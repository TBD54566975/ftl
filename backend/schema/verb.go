package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Visibility string

const (
	Public   Visibility = "public"
	Internal Visibility = "internal"
	Private  Visibility = "private"
)

type Verb struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments   []string   `parser:"@Comment*" protobuf:"4"`
	Visibility string     `parser:"@('public' | 'internal' | 'private')?" protobuf:"3"`
	Name       string     `parser:"'verb' @Ident" protobuf:"2"`
	Request    Type       `parser:"'(' @@ ')'" protobuf:"5"`
	Response   Type       `parser:"@@" protobuf:"6"`
	Metadata   []Metadata `parser:"@@*" protobuf:"7"`
}

var _ Decl = (*Verb)(nil)
var _ Symbol = (*Verb)(nil)

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

func (v *Verb) GetName() string { return v.Name }

func (v *Verb) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(v.Comments))
	fmt.Fprintf(w, "%s verb %s(%s) %s", v.Visibility, v.Name, v.Request, v.Response)
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
		Pos:        posToProto(v.Pos),
		Name:       v.Name,
		Visibility: v.Visibility,
		Comments:   v.Comments,
		Request:    typeToProto(v.Request),
		Response:   typeToProto(v.Response),
		Metadata:   metadataListToProto(v.Metadata),
	}
}

func VerbFromProto(s *schemapb.Verb) *Verb {
	return &Verb{
		Pos:        posFromProto(s.Pos),
		Name:       s.Name,
		Visibility: s.Visibility,
		Comments:   s.Comments,
		Request:    typeToSchema(s.Request),
		Response:   typeToSchema(s.Response),
		Metadata:   metadataListToSchema(s.Metadata),
	}
}
