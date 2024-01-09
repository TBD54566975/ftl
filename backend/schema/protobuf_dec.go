package schema

import (
	"fmt"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

func posFromProto(pos *schemapb.Position) Position {
	if pos == nil {
		return Position{}
	}
	return Position{
		Line:     int(pos.Line),
		Column:   int(pos.Column),
		Filename: pos.Filename,
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

func typeToSchema(s *schemapb.Type) Type {
	switch s := s.Value.(type) {
	// case *schemapb.Type_VerbRef:
	// 	return verbRefToSchema(s.VerbRef)
	case *schemapb.Type_DataRef:
		return dataRefToSchema(s.DataRef)
	case *schemapb.Type_Int:
		return &Int{Pos: posFromProto(s.Int.Pos)}
	case *schemapb.Type_Float:
		return &Float{Pos: posFromProto(s.Float.Pos)}
	case *schemapb.Type_String_:
		return &String{Pos: posFromProto(s.String_.Pos)}
	case *schemapb.Type_Bytes:
		return &Bytes{Pos: posFromProto(s.Bytes.Pos)}
	case *schemapb.Type_Time:
		return &Time{Pos: posFromProto(s.Time.Pos)}
	case *schemapb.Type_Bool:
		return &Bool{Pos: posFromProto(s.Bool.Pos)}
	case *schemapb.Type_Array:
		return arrayToSchema(s.Array)
	case *schemapb.Type_Map:
		return mapToSchema(s.Map)
	case *schemapb.Type_Optional:
		return &Optional{Pos: posFromProto(s.Optional.Pos), Type: typeToSchema(s.Optional.Type)}
	}
	panic("unreachable")
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
			Pos:   posFromProto(s.Calls.Pos),
			Calls: verbRefListToSchema(s.Calls.Calls),
		}

	case *schemapb.Metadata_Ingress:
		return &MetadataIngress{
			Pos:    posFromProto(s.Ingress.Pos),
			Method: s.Ingress.Method,
			Path:   ingressPathComponentListToSchema(s.Ingress.Path),
		}

	default:
		panic(fmt.Sprintf("unhandled metadata type: %T", s))
	}
}
