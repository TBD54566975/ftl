package schema

type Node interface {
	String() string
	// schemaChildren returns the children of this node.
	schemaChildren() []Node
}

type Type interface {
	Node
	// schemaType is a marker to ensure that all types implement the Type interface.
	schemaType()
}

type Int struct{}

type Float struct{}

type String struct{}

type Bool struct{}

type Array struct {
	Element Type `json:"element"`
}

type Map struct {
	Key   Type `json:"key"`
	Value Type `json:"value"`
}

type Field struct {
	Name string `json:"name"`
	Type Type   `json:"type"`
}

// DataRef is a reference to a data structure.
type DataRef struct {
	Name string `json:"name"`
}

// A Data structure.
type Data struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type VerbRef struct {
	Module string `json:"module"`
	Verb   string `json:"name"`
}

type Verb struct {
	Name     string    `json:"name"`
	Request  DataRef   `json:"request"`
	Response DataRef   `json:"response"`
	Calls    []VerbRef `json:"calls,omitempty"`
}

type Module struct {
	Name  string `json:"name"`
	Data  []Data `json:"data,omitempty"`
	Verbs []Verb `json:"verbs,omitempty"`
}
