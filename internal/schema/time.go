package schema

//protobuf:6
type Time struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Time bool `parser:"@'Time'" protobuf:"-"`
}

var _ Type = (*Time)(nil)
var _ Symbol = (*Time)(nil)

func (t *Time) Equal(other Type) bool { _, ok := other.(*Time); return ok }
func (t *Time) Position() Position    { return t.Pos }
func (*Time) schemaChildren() []Node  { return nil }
func (*Time) schemaType()             {}
func (*Time) schemaSymbol()           {}
func (*Time) String() string          { return "Time" }
func (*Time) GetName() string         { return "Time" }
