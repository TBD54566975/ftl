package schema

//protobuf:4
type Bytes struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Bytes bool `parser:"@'Bytes'" protobuf:"-"`
}

var _ Type = (*Bytes)(nil)
var _ Symbol = (*Bytes)(nil)

func (b *Bytes) Equal(other Type) bool { _, ok := other.(*Bytes); return ok }
func (b *Bytes) Position() Position    { return b.Pos }
func (*Bytes) schemaChildren() []Node  { return nil }
func (*Bytes) schemaType()             {}
func (*Bytes) schemaSymbol()           {}
func (*Bytes) String() string          { return "Bytes" }
func (*Bytes) GetName() string         { return "Bytes" }
