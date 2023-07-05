package schema

import (
	"github.com/alecthomas/errors"
	"google.golang.org/protobuf/proto"

	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// FromProto converts a protobuf Schema to a Schema and validates it.
func FromProto(s *pschema.Schema) (*Schema, error) {
	modules, err := moduleListToSchema(s.Modules)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	schema := &Schema{
		Modules: modules,
	}
	return schema, Validate(schema)
}

func moduleListToSchema(s []*pschema.Module) ([]*Module, error) {
	var out []*Module
	for _, n := range s {
		module, err := ModuleFromProto(n)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		out = append(out, module)
	}
	return out, nil
}

// ModuleFromProto converts a protobuf Module to a Module and validates it.
func ModuleFromProto(s *pschema.Module) (*Module, error) {
	module := &Module{
		Name:     s.Name,
		Comments: s.Comments,
		Decls:    declListToSchema(s.Decls),
	}
	return module, ValidateModule(module)
}

func ModuleFromBytes(b []byte) (*Module, error) {
	s := &pschema.Module{}
	if err := proto.Unmarshal(b, s); err != nil {
		return nil, errors.WithStack(err)
	}
	return ModuleFromProto(s)
}

// VerbRefFromProto converts a protobuf VerbRef to a VerbRef.
func VerbRefFromProto(s *pschema.VerbRef) *VerbRef {
	return &VerbRef{
		Module: s.Module,
		Name:   s.Name,
	}
}

func declListToSchema(s []*pschema.Decl) []Decl {
	var out []Decl
	for _, n := range s {
		switch n := n.Value.(type) {
		case *pschema.Decl_Verb:
			out = append(out, verbToSchema(n.Verb))
		case *pschema.Decl_Data:
			out = append(out, dataToSchema(n.Data))
		}
	}
	return out
}

func verbToSchema(s *pschema.Verb) *Verb {
	return &Verb{
		Name:     s.Name,
		Comments: s.Comments,
		Request:  dataRefToSchema(s.Request),
		Response: dataRefToSchema(s.Response),
		Metadata: metadataListToSchema(s.Metadata),
	}
}

func dataToSchema(s *pschema.Data) *Data {
	return &Data{
		Name:     s.Name,
		Fields:   fieldListToSchema(s.Fields),
		Comments: s.Comments,
	}
}

func fieldListToSchema(s []*pschema.Field) []*Field {
	var out []*Field
	for _, n := range s {
		out = append(out, fieldToSchema(n))
	}
	return out
}

func fieldToSchema(s *pschema.Field) *Field {
	return &Field{
		Name:     s.Name,
		Comments: s.Comments,
		Type:     typeToSchema(s.Type),
	}
}

func typeToSchema(s *pschema.Type) Type {
	switch s := s.Value.(type) {
	case *pschema.Type_VerbRef:
		return verbRefToSchema(s.VerbRef)
	case *pschema.Type_DataRef:
		return dataRefToSchema(s.DataRef)
	case *pschema.Type_Int:
		return &Int{}
	case *pschema.Type_Float:
		return &Float{}
	case *pschema.Type_String_:
		return &String{}
	case *pschema.Type_Time:
		return &Time{}
	case *pschema.Type_Bool:
		return &Bool{}
	case *pschema.Type_Array:
		return arrayToSchema(s.Array)
	case *pschema.Type_Map:
		return mapToSchema(s.Map)
	}
	panic("unreachable")
}

func verbRefToSchema(s *pschema.VerbRef) *VerbRef {
	return &VerbRef{
		Name:   s.Name,
		Module: s.Module,
	}
}

func dataRefToSchema(s *pschema.DataRef) *DataRef {
	return &DataRef{
		Name:   s.Name,
		Module: s.Module,
	}
}

func arrayToSchema(s *pschema.Array) *Array {
	return &Array{
		Element: typeToSchema(s.Element),
	}
}

func mapToSchema(s *pschema.Map) *Map {
	return &Map{
		Key:   typeToSchema(s.Key),
		Value: typeToSchema(s.Value),
	}
}

func metadataListToSchema(s []*pschema.Metadata) []Metadata {
	var out []Metadata
	for _, n := range s {
		out = append(out, metadataToSchema(n))
	}
	return out
}

func metadataToSchema(s *pschema.Metadata) Metadata {
	switch s := s.Value.(type) { //nolint:gocritic
	case *pschema.Metadata_Calls:
		return &MetadataCalls{
			Calls: verbRefListToSchema(s.Calls.Calls),
		}
	}
	panic("unreachable")
}

func verbRefListToSchema(s []*pschema.VerbRef) []*VerbRef {
	var out []*VerbRef
	for _, n := range s {
		out = append(out, verbRefToSchema(n))
	}
	return out
}
