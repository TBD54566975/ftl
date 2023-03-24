package schema

import pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"

func ProtoToSchema(s *pschema.Schema) *Schema {
	return &Schema{
		Modules: moduleListToSchema(s.Modules),
	}
}

func moduleListToSchema(s []*pschema.Module) []*Module {
	out := []*Module{}
	for _, n := range s {
		out = append(out, ProtoToModule(n))
	}
	return out
}

func ProtoToModule(s *pschema.Module) *Module {
	return &Module{
		Name:     s.Name,
		Comments: s.Comments,
		Decls:    declListToSchema(s.Decls),
	}
}

func declListToSchema(s []*pschema.Decl) []Decl {
	out := []Decl{}
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
		Name:   s.Name,
		Fields: fieldListToSchema(s.Fields),
	}
}

func fieldListToSchema(s []*pschema.Field) []*Field {
	out := []*Field{}
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
	out := []Metadata{}
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
	out := []*VerbRef{}
	for _, n := range s {
		out = append(out, verbRefToSchema(n))
	}
	return out
}
