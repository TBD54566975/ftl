package schema

//protobuf:3
type String struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Str bool `parser:"@'String'" protobuf:"-"`
}

var _ Type = (*String)(nil)
var _ Symbol = (*String)(nil)

func (s *String) Equal(other Type) bool { _, ok := other.(*String); return ok }
func (s *String) Position() Position    { return s.Pos }
func (*String) schemaChildren() []Node  { return nil }
func (*String) schemaType()             {}
func (*String) schemaSymbol()           {}
func (*String) String() string          { return "String" }
func (*String) GetName() string         { return "String" }
