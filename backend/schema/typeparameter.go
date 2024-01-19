package schema

import "google.golang.org/protobuf/reflect/protoreflect"

type TypeParameter struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"@Ident" protobuf:"2"`
}

var _ Type = (*TypeParameter)(nil)
var _ Decl = (*TypeParameter)(nil)

func (*TypeParameter) schemaType()          {}
func (t *TypeParameter) Position() Position { return t.Pos }
func (t *TypeParameter) String() string     { return t.Name }
func (t *TypeParameter) ToProto() protoreflect.ProtoMessage {
	panic("unimplemented")
}
func (t *TypeParameter) schemaChildren() []Node { return nil }
func (t *TypeParameter) schemaDecl()            {}
