package schema

var _ Value = (*TypeValue)(nil)

//protobuf:3
type TypeValue struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Value Type `parser:"@@" protobuf:"2"`
}

func (t *TypeValue) Position() Position { return t.Pos }

func (t *TypeValue) schemaChildren() []Node { return []Node{t.Value} }

func (t *TypeValue) String() string {
	return t.Value.String()
}

func (t *TypeValue) GetValue() any { return t.Value.String() }

func (t *TypeValue) schemaValueType() Type { return t.Value }
