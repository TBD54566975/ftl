//nolint:forcetypeassert
package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

func posToProto(pos Position) *schemapb.Position {
	return &schemapb.Position{Line: int64(pos.Line), Column: int64(pos.Column), Filename: pos.Filename}
}

func nodeListToProto[T proto.Message, U Node](nodes []U) []T {
	out := make([]T, len(nodes))
	for i, n := range nodes {
		out[i] = n.ToProto().(T)
	}
	return out
}

func declListToProto(nodes []Decl) []*schemapb.Decl {
	out := make([]*schemapb.Decl, len(nodes))
	for i, n := range nodes {
		var v schemapb.IsDeclValue
		switch n := n.(type) {
		case *Verb:
			v = &schemapb.Decl_Verb{Verb: n.ToProto().(*schemapb.Verb)}

		case *Data:
			v = &schemapb.Decl_Data{Data: n.ToProto().(*schemapb.Data)}

		case *Database:
			v = &schemapb.Decl_Database{Database: n.ToProto().(*schemapb.Database)}

		case *Bool, *Bytes, *Float, *Int, *Module, *String, *Time, *Unit, *Any:
		}
		out[i] = &schemapb.Decl{Value: v}
	}
	return out
}

func metadataListToProto(nodes []Metadata) []*schemapb.Metadata {
	out := make([]*schemapb.Metadata, len(nodes))
	for i, n := range nodes {
		var v schemapb.IsMetadataValue
		switch n := n.(type) {
		case *MetadataCalls:
			v = &schemapb.Metadata_Calls{Calls: n.ToProto().(*schemapb.MetadataCalls)}

		case *MetadataDatabases:
			v = &schemapb.Metadata_Databases{Databases: n.ToProto().(*schemapb.MetadataDatabases)}

		case *MetadataIngress:
			v = &schemapb.Metadata_Ingress{Ingress: n.ToProto().(*schemapb.MetadataIngress)}

		default:
			panic(fmt.Sprintf("unhandled metadata type %T", n))
		}
		out[i] = &schemapb.Metadata{Value: v}
	}
	return out
}

func ingressListToProto(nodes []IngressPathComponent) []*schemapb.IngressPathComponent {
	out := make([]*schemapb.IngressPathComponent, len(nodes))
	for i, n := range nodes {
		switch n := n.(type) {
		case *IngressPathLiteral:
			out[i] = &schemapb.IngressPathComponent{Value: &schemapb.IngressPathComponent_IngressPathLiteral{IngressPathLiteral: n.ToProto().(*schemapb.IngressPathLiteral)}}
		case *IngressPathParameter:
			out[i] = &schemapb.IngressPathComponent{Value: &schemapb.IngressPathComponent_IngressPathParameter{IngressPathParameter: n.ToProto().(*schemapb.IngressPathParameter)}}

		default:
			panic(fmt.Sprintf("unhandled ingress path component type %T", n))
		}
	}
	return out
}

func typeToProto(t Type) *schemapb.Type {
	switch t := t.(type) {
	case *Any:
		return &schemapb.Type{Value: &schemapb.Type_Any{Any: t.ToProto().(*schemapb.Any)}}

	case *Unit:
		return &schemapb.Type{Value: &schemapb.Type_Unit{Unit: t.ToProto().(*schemapb.Unit)}}

	case *VerbRef, *SourceRef, *SinkRef:
		panic("unreachable")

	case *DataRef:
		return &schemapb.Type{Value: &schemapb.Type_DataRef{DataRef: t.ToProto().(*schemapb.DataRef)}}

	case *Int:
		return &schemapb.Type{Value: &schemapb.Type_Int{Int: t.ToProto().(*schemapb.Int)}}

	case *Float:
		return &schemapb.Type{Value: &schemapb.Type_Float{Float: t.ToProto().(*schemapb.Float)}}

	case *String:
		return &schemapb.Type{Value: &schemapb.Type_String_{String_: t.ToProto().(*schemapb.String)}}

	case *Bytes:
		return &schemapb.Type{Value: &schemapb.Type_Bytes{Bytes: t.ToProto().(*schemapb.Bytes)}}

	case *Time:
		return &schemapb.Type{Value: &schemapb.Type_Time{Time: t.ToProto().(*schemapb.Time)}}

	case *Bool:
		return &schemapb.Type{Value: &schemapb.Type_Bool{Bool: t.ToProto().(*schemapb.Bool)}}

	case *Array:
		return &schemapb.Type{Value: &schemapb.Type_Array{Array: t.ToProto().(*schemapb.Array)}}

	case *Map:
		return &schemapb.Type{Value: &schemapb.Type_Map{Map: t.ToProto().(*schemapb.Map)}}

	case *Optional:
		return &schemapb.Type{Value: &schemapb.Type_Optional{Optional: t.ToProto().(*schemapb.Optional)}}
	}
	panic(fmt.Sprintf("unhandled type: %T", t))
}
