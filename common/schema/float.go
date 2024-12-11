package schema

//protobuf:2
type Float struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Float bool `parser:"@'Float'" protobuf:"-"`
}

var _ Type = (*Float)(nil)
var _ Symbol = (*Float)(nil)

func (f *Float) Equal(other Type) bool { _, ok := other.(*Float); return ok }
func (f *Float) Position() Position    { return f.Pos }
func (*Float) schemaChildren() []Node  { return nil }
func (*Float) schemaType()             {}
func (*Float) schemaSymbol()           {}
func (*Float) String() string          { return "Float" }
func (*Float) GetName() string         { return "Float" }
