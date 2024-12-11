package schema

//protobuf:10
type Unit struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Unit bool `parser:"@'Unit'" protobuf:"-"`
}

var _ Type = (*Unit)(nil)
var _ Symbol = (*Unit)(nil)

func (u *Unit) Equal(other Type) bool  { _, ok := other.(*Unit); return ok }
func (u *Unit) Position() Position     { return u.Pos }
func (u *Unit) schemaType()            {}
func (u *Unit) schemaSymbol()          {}
func (u *Unit) String() string         { return "Unit" }
func (u *Unit) schemaChildren() []Node { return nil }
func (u *Unit) GetName() string        { return "Unit" }
