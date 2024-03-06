package schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// AliasKind is the kind of alias.
//
//go:generate enumer -type AliasKind -trimprefix AliasKind -transform=lower -json -text
type AliasKind int

const (
	AliasKindJSON AliasKind = iota
)

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

func (m *MetadataAlias) ToProto() protoreflect.ProtoMessage {
	return &schemapb.MetadataAlias{
		Pos:   posToProto(m.Pos),
		Kind:  int64(m.Kind),
		Alias: m.Alias,
	}
}

func (m *MetadataAlias) schemaChildren() []Node { return nil }
func (m *MetadataAlias) schemaMetadata()        {}
