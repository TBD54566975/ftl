package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Verb struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"3"`
	Name     string     `parser:"'verb' @Ident" protobuf:"2"`
	Request  *DataRef   `parser:"'(' @@ ')'" protobuf:"4"`
	Response *DataRef   `parser:"@@" protobuf:"5"`
	Metadata []Metadata `parser:"@@*" protobuf:"6"`
}

var _ Decl = (*Verb)(nil)

func (v *Verb) schemaDecl() {}
func (v *Verb) schemaChildren() []Node {
	children := make([]Node, 2+len(v.Metadata))
	children[0] = v.Request
	children[1] = v.Response
	for i, c := range v.Metadata {
		children[i+2] = c
	}
	return children
}
func (v *Verb) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(v.Comments))
	fmt.Fprintf(w, "verb %s(%s) %s", v.Name, v.Request, v.Response)
	fmt.Fprint(w, indent(encodeMetadata(v.Metadata)))
	return w.String()
}

// AddCall adds a call reference to the Verb.
func (v *Verb) AddCall(verb *VerbRef) {
	for _, c := range v.Metadata {
		if c, ok := c.(*MetadataCalls); ok {
			c.Calls = append(c.Calls, verb)
			return
		}
	}
	v.Metadata = append(v.Metadata, &MetadataCalls{Calls: []*VerbRef{verb}})
}

func (v *Verb) ToProto() proto.Message {
	return &schemapb.Verb{
		Pos:      posToProto(v.Pos),
		Name:     v.Name,
		Comments: v.Comments,
		Request:  v.Request.ToProto().(*schemapb.DataRef),  //nolint:forcetypeassert
		Response: v.Response.ToProto().(*schemapb.DataRef), //nolint:forcetypeassert
		Metadata: metadataListToProto(v.Metadata),
	}
}

func VerbToSchema(s *schemapb.Verb) *Verb {
	return &Verb{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Request:  dataRefToSchema(s.Request),
		Response: dataRefToSchema(s.Response),
		Metadata: metadataListToSchema(s.Metadata),
	}
}
