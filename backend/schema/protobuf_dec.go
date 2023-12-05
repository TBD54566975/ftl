package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// FromProto converts a protobuf Schema to a Schema and validates it.
func FromProto(s *schemapb.Schema) (*Schema, error) {
	modules, err := moduleListToSchema(s.Modules)
	if err != nil {
		return nil, err
	}
	schema := &Schema{
		Modules: modules,
	}
	return schema, Validate(schema)
}

func moduleListToSchema(s []*schemapb.Module) ([]*Module, error) {
	var out []*Module
	for _, n := range s {
		module, err := ModuleFromProto(n)
		if err != nil {
			return nil, err
		}
		out = append(out, module)
	}
	return out, nil
}

func PosFromProto(pos *schemapb.Position) Position {
	if pos == nil {
		return Position{}
	}
	return Position{
		Line:     int(pos.Line),
		Column:   int(pos.Column),
		Filename: pos.Filename,
	}
}

// ModuleFromProto converts a protobuf Module to a Module and validates it.
func ModuleFromProto(s *schemapb.Module) (*Module, error) {
	module := &Module{
		Pos:      PosFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Decls:    declListToSchema(s.Decls),
	}
	return module, ValidateModule(module)
}

func ModuleFromBytes(b []byte) (*Module, error) {
	s := &schemapb.Module{}
	if err := proto.Unmarshal(b, s); err != nil {
		return nil, err
	}
	return ModuleFromProto(s)
}

// VerbRefFromProto converts a protobuf VerbRef to a VerbRef.
func VerbRefFromProto(s *schemapb.VerbRef) *VerbRef {
	return &VerbRef{
		Module: s.Module,
		Name:   s.Name,
	}
}

func declListToSchema(s []*schemapb.Decl) []Decl {
	var out []Decl
	for _, n := range s {
		switch n := n.Value.(type) {
		case *schemapb.Decl_Verb:
			out = append(out, VerbToSchema(n.Verb))
		case *schemapb.Decl_Data:
			out = append(out, DataToSchema(n.Data))
		}
	}
	return out
}

func VerbToSchema(s *schemapb.Verb) *Verb {
	return &Verb{
		Pos:      PosFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Request:  dataRefToSchema(s.Request),
		Response: dataRefToSchema(s.Response),
		Metadata: metadataListToSchema(s.Metadata),
	}
}

func DataToSchema(s *schemapb.Data) *Data {
	return &Data{
		Pos:      PosFromProto(s.Pos),
		Name:     s.Name,
		Fields:   fieldListToSchema(s.Fields),
		Comments: s.Comments,
	}
}

func fieldListToSchema(s []*schemapb.Field) []*Field {
	var out []*Field
	for _, n := range s {
		out = append(out, fieldToSchema(n))
	}
	return out
}

func fieldToSchema(s *schemapb.Field) *Field {
	return &Field{
		Pos:      PosFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Type:     typeToSchema(s.Type),
	}
}

func typeToSchema(s *schemapb.Type) Type {
	switch s := s.Value.(type) {
	// case *schemapb.Type_VerbRef:
	// 	return verbRefToSchema(s.VerbRef)
	case *schemapb.Type_DataRef:
		return dataRefToSchema(s.DataRef)
	case *schemapb.Type_Int:
		return &Int{Pos: PosFromProto(s.Int.Pos)}
	case *schemapb.Type_Float:
		return &Float{Pos: PosFromProto(s.Float.Pos)}
	case *schemapb.Type_String_:
		return &String{Pos: PosFromProto(s.String_.Pos)}
	case *schemapb.Type_Time:
		return &Time{Pos: PosFromProto(s.Time.Pos)}
	case *schemapb.Type_Bool:
		return &Bool{Pos: PosFromProto(s.Bool.Pos)}
	case *schemapb.Type_Array:
		return arrayToSchema(s.Array)
	case *schemapb.Type_Map:
		return mapToSchema(s.Map)
	case *schemapb.Type_Optional:
		return &Optional{Pos: PosFromProto(s.Optional.Pos), Type: typeToSchema(s.Optional.Type)}
	}
	panic("unreachable")
}

func verbRefToSchema(s *schemapb.VerbRef) *VerbRef {
	return &VerbRef{
		Pos:    PosFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func dataRefToSchema(s *schemapb.DataRef) *DataRef {
	return &DataRef{
		Pos:    PosFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func arrayToSchema(s *schemapb.Array) *Array {
	return &Array{
		Pos:     PosFromProto(s.Pos),
		Element: typeToSchema(s.Element),
	}
}

func mapToSchema(s *schemapb.Map) *Map {
	return &Map{
		Pos:   PosFromProto(s.Pos),
		Key:   typeToSchema(s.Key),
		Value: typeToSchema(s.Value),
	}
}

func metadataListToSchema(s []*schemapb.Metadata) []Metadata {
	var out []Metadata
	for _, n := range s {
		out = append(out, metadataToSchema(n))
	}
	return out
}

func metadataToSchema(s *schemapb.Metadata) Metadata {
	switch s := s.Value.(type) {
	case *schemapb.Metadata_Calls:
		return &MetadataCalls{
			Pos:   PosFromProto(s.Calls.Pos),
			Calls: verbRefListToSchema(s.Calls.Calls),
		}

	case *schemapb.Metadata_Ingress:
		return &MetadataIngress{
			Pos:    PosFromProto(s.Ingress.Pos),
			Method: s.Ingress.Method,
			Path:   s.Ingress.Path,
		}

	default:
		panic(fmt.Sprintf("unhandled metadata type: %T", s))
	}
}

func verbRefListToSchema(s []*schemapb.VerbRef) []*VerbRef {
	var out []*VerbRef
	for _, n := range s {
		out = append(out, verbRefToSchema(n))
	}
	return out
}
