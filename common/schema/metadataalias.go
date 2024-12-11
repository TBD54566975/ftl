package schema

import (
	"fmt"
)

// AliasKind is the kind of alias.
//
//go:generate enumer -type AliasKind -trimprefix AliasKind -transform=lower -json -text
type AliasKind int

const (
	AliasKindUnspecified AliasKind = iota
	AliasKindJson                  //nolint
)

//protobuf:5
type MetadataAlias struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Kind  AliasKind `parser:"'+' 'alias' @Ident" protobuf:"2"`
	Alias string    `parser:"@String" protobuf:"3"`
}

var _ Metadata = (*MetadataAlias)(nil)

func (m *MetadataAlias) Position() Position { return m.Pos }

func (m *MetadataAlias) String() string {
	return fmt.Sprintf("+alias %s %q", m.Kind, m.Alias)
}

func (m *MetadataAlias) schemaChildren() []Node { return nil }
func (m *MetadataAlias) schemaMetadata()        {}
