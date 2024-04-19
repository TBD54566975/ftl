package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type SumType struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments    []string      `parser:"@Comment*" protobuf:"2"`
	Name        string        `parser:"'sumtype' @Ident" protobuf:"3"`
	AddendTypes []*AddendType `parser:"'{' @@* '}'" protobuf:"4"`
}

var _ Decl = (*SumType)(nil)
var _ Symbol = (*SumType)(nil)

func (s *SumType) Position() Position { return s.Pos }

func (s *SumType) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(s.Comments))
	fmt.Fprintf(w, "sumtype %s {\n", s.Name)
	for _, v := range s.AddendTypes {
		fmt.Fprintln(w, v.String())
	}
	fmt.Fprintf(w, "}")
	return w.String()
}

func (s *SumType) schemaDecl() {}
func (*SumType) schemaSymbol() {}
func (s *SumType) schemaChildren() []Node {
	children := make([]Node, len(s.AddendTypes))
	for i, v := range s.AddendTypes {
		children[i] = v
	}
	return children
}
func (s *SumType) ToProto() proto.Message {
	return &schemapb.SumType{
		Pos:         posToProto(s.Pos),
		Comments:    s.Comments,
		Name:        s.Name,
		AddendTypes: s.addendTypesToProto(),
	}
}
func (s *SumType) addendTypesToProto() []*schemapb.AddendType {
	protoAddendTypes := make([]*schemapb.AddendType, len(s.AddendTypes))
	for i, v := range s.AddendTypes {
		protoAddendTypes[i] = v.ToProto()
	}
	return protoAddendTypes
}

func (s *SumType) GetName() string { return s.Name }

func SumTypeFromProto(s *schemapb.SumType) *SumType {
	return &SumType{
		Pos:         posFromProto(s.Pos),
		Name:        s.Name,
		Comments:    s.Comments,
		AddendTypes: addendTypesToSchema(s.AddendTypes),
	}
}

type AddendType struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Type     Type     `parser:"@@" protobuf:"3"`
}

func (a *AddendType) ToProto() proto.Message {
	return &schemapb.AddendType{
		Pos:      posToProto(a.Pos),
		Comments: a.Comments,
		Type:     typeToProto(a.Type),
	}
}

func (a *AddendType) Position() Position     { return a.Pos }
func (a *AddendType) String() string         { return a.Type.String() }
func (a *AddendType) schemaChildren() []Node { return nil }

func addendTypesToSchema(pbAddendTypes []*schemapb.AddendType) []*AddendType {
	addendTypes := make([]*AddendType, len(pbAddendTypes))
	for i, v := range pbAddendTypes {
		addendTypes[i] = addendTypeToSchema(v)
	}
	return addendTypes
}

func addendTypeToSchema(addendType *schemapb.AddendType) *AddendType {
	return &AddendType{
		Pos:      posFromProto(addendType.Pos),
		Comments: addendType.Comments,
		Type:     typeToSchema(addendType.Type),
	}
}
