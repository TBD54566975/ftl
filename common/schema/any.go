package schema

//protobuf:9
type Any struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Any bool `parser:"@'Any'" protobuf:"-"`
}

var _ Type = (*Any)(nil)
var _ Symbol = (*Any)(nil)

func (a *Any) Position() Position   { return a.Pos }
func (*Any) schemaChildren() []Node { return nil }
func (*Any) schemaType()            {}
func (*Any) schemaSymbol()          {}
func (*Any) String() string         { return "Any" }
func (*Any) Equal(other Type) bool  { _, ok := other.(*Any); return ok }
func (*Any) GetName() string        { return "Any" }
