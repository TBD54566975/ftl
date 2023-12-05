//nolint:forcetypeassert
package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

func nodeListToProto[T proto.Message, U Node](nodes []U) []T {
	out := []T{}
	for _, n := range nodes {
		out = append(out, n.ToProto().(T))
	}
	return out
}

func declListToProto(nodes []Decl) []*schemapb.Decl {
	out := []*schemapb.Decl{}
	for _, n := range nodes {
		var v schemapb.IsDeclValue
		switch n := n.(type) {
		case *Verb:
			v = &schemapb.Decl_Verb{Verb: n.ToProto().(*schemapb.Verb)}
		case *Data:
			v = &schemapb.Decl_Data{Data: n.ToProto().(*schemapb.Data)}
		}
		out = append(out, &schemapb.Decl{Value: v})
	}
	return out
}

func metadataListToProto(nodes []Metadata) []*schemapb.Metadata {
	var out []*schemapb.Metadata
	for _, n := range nodes {
		var v schemapb.IsMetadataValue
		switch n := n.(type) {
		case *MetadataCalls:
			v = &schemapb.Metadata_Calls{Calls: n.ToProto().(*schemapb.MetadataCalls)}

		case *MetadataIngress:
			v = &schemapb.Metadata_Ingress{Ingress: n.ToProto().(*schemapb.MetadataIngress)}

		default:
			panic(fmt.Sprintf("unhandled metadata type %T", n))
		}
		out = append(out, &schemapb.Metadata{Value: v})
	}
	return out
}

func (p Position) ToProto() proto.Message {
	return &schemapb.Position{
		Line:     int64(p.Line),
		Column:   int64(p.Column),
		Filename: p.Filename,
	}
}

func (s *Schema) ToProto() proto.Message {
	return &schemapb.Schema{
		Pos:     s.Pos.ToProto().(*schemapb.Position),
		Modules: nodeListToProto[*schemapb.Module](s.Modules),
	}
}

func (m *Module) ToProto() proto.Message {
	return &schemapb.Module{
		Pos:      m.Pos.ToProto().(*schemapb.Position),
		Name:     m.Name,
		Comments: m.Comments,
		Decls:    declListToProto(m.Decls),
	}
}

func (v *Verb) ToProto() proto.Message {
	return &schemapb.Verb{
		Pos:      v.Pos.ToProto().(*schemapb.Position),
		Name:     v.Name,
		Comments: v.Comments,
		Request:  v.Request.ToProto().(*schemapb.DataRef),
		Response: v.Response.ToProto().(*schemapb.DataRef),
		Metadata: metadataListToProto(v.Metadata),
	}
}

func (d *Data) ToProto() proto.Message {
	return &schemapb.Data{
		Pos:      d.Pos.ToProto().(*schemapb.Position),
		Name:     d.Name,
		Fields:   nodeListToProto[*schemapb.Field](d.Fields),
		Comments: d.Comments,
	}
}

func (f *Field) ToProto() proto.Message {
	return &schemapb.Field{
		Pos:      f.Pos.ToProto().(*schemapb.Position),
		Name:     f.Name,
		Type:     typeToProto(f.Type),
		Comments: f.Comments,
	}
}

func (v *VerbRef) ToProto() proto.Message {
	return &schemapb.VerbRef{
		Pos:    v.Pos.ToProto().(*schemapb.Position),
		Name:   v.Name,
		Module: v.Module,
	}
}

func (s *DataRef) ToProto() proto.Message {
	return &schemapb.DataRef{
		Pos:    s.Pos.ToProto().(*schemapb.Position),
		Name:   s.Name,
		Module: s.Module,
	}
}

func (m *MetadataCalls) ToProto() proto.Message {
	return &schemapb.MetadataCalls{
		Pos:   m.Pos.ToProto().(*schemapb.Position),
		Calls: nodeListToProto[*schemapb.VerbRef](m.Calls),
	}
}

func (m *MetadataIngress) ToProto() proto.Message {
	return &schemapb.MetadataIngress{
		Pos:    m.Pos.ToProto().(*schemapb.Position),
		Method: m.Method,
		Path:   m.Path,
	}
}

func (i *Int) ToProto() proto.Message {
	return &schemapb.Int{}
}

func (s *String) ToProto() proto.Message {
	return &schemapb.String{}
}

func (b *Bool) ToProto() proto.Message {
	return &schemapb.Bool{}
}

func (f *Float) ToProto() proto.Message {
	return &schemapb.Float{}
}

func (t *Time) ToProto() proto.Message {
	return &schemapb.Time{}
}

func (m *Map) ToProto() proto.Message {
	return &schemapb.Map{
		Key:   typeToProto(m.Key),
		Value: typeToProto(m.Value),
	}
}

func (a *Array) ToProto() proto.Message {
	return &schemapb.Array{Element: typeToProto(a.Element)}
}

func (o *Optional) ToProto() proto.Message {
	return &schemapb.Optional{Type: typeToProto(o.Type)}
}

func typeToProto(t Type) *schemapb.Type {
	switch t := t.(type) {
	case *VerbRef:
		panic("unreachable")

	case *DataRef:
		return &schemapb.Type{Value: &schemapb.Type_DataRef{DataRef: t.ToProto().(*schemapb.DataRef)}}

	case *Int:
		return &schemapb.Type{Value: &schemapb.Type_Int{Int: t.ToProto().(*schemapb.Int)}}

	case *Float:
		return &schemapb.Type{Value: &schemapb.Type_Float{Float: t.ToProto().(*schemapb.Float)}}

	case *String:
		return &schemapb.Type{Value: &schemapb.Type_String_{String_: t.ToProto().(*schemapb.String)}}

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
	panic("unreachable")
}
