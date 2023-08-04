//nolint:forcetypeassert
package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

func nodeListToProto[T proto.Message, U Node](nodes []U) []T {
	out := []T{}
	for _, n := range nodes {
		out = append(out, n.ToProto().(T))
	}
	return out
}

func declListToProto(nodes []Decl) []*pschema.Decl {
	out := []*pschema.Decl{}
	for _, n := range nodes {
		var v pschema.IsDeclValue
		switch n := n.(type) {
		case *Verb:
			v = &pschema.Decl_Verb{Verb: n.ToProto().(*pschema.Verb)}
		case *Data:
			v = &pschema.Decl_Data{Data: n.ToProto().(*pschema.Data)}
		}
		out = append(out, &pschema.Decl{Value: v})
	}
	return out
}

func metadataListToProto(nodes []Metadata) []*pschema.Metadata {
	var out []*pschema.Metadata
	for _, n := range nodes {
		var v pschema.IsMetadataValue
		switch n := n.(type) {
		case *MetadataCalls:
			v = &pschema.Metadata_Calls{Calls: n.ToProto().(*pschema.MetadataCalls)}

		case *MetadataIngress:
			v = &pschema.Metadata_Ingress{Ingress: n.ToProto().(*pschema.MetadataIngress)}

		default:
			panic(fmt.Sprintf("unhandled metadata type %T", n))
		}
		out = append(out, &pschema.Metadata{Value: v})
	}
	return out
}

func typeToProto(t Type) *pschema.Type {
	switch t.(type) {
	case *VerbRef:
		return &pschema.Type{Value: &pschema.Type_VerbRef{VerbRef: t.ToProto().(*pschema.VerbRef)}}

	case *DataRef:
		return &pschema.Type{Value: &pschema.Type_DataRef{DataRef: t.ToProto().(*pschema.DataRef)}}

	case *Int:
		return &pschema.Type{Value: &pschema.Type_Int{Int: t.ToProto().(*pschema.Int)}}

	case *Float:
		return &pschema.Type{Value: &pschema.Type_Float{Float: t.ToProto().(*pschema.Float)}}

	case *String:
		return &pschema.Type{Value: &pschema.Type_String_{String_: t.ToProto().(*pschema.String)}}

	case *Time:
		return &pschema.Type{Value: &pschema.Type_Time{Time: t.ToProto().(*pschema.Time)}}

	case *Bool:
		return &pschema.Type{Value: &pschema.Type_Bool{Bool: t.ToProto().(*pschema.Bool)}}

	case *Array:
		return &pschema.Type{Value: &pschema.Type_Array{Array: t.ToProto().(*pschema.Array)}}

	case *Map:
		return &pschema.Type{Value: &pschema.Type_Map{Map: t.ToProto().(*pschema.Map)}}
	}
	panic("unreachable")
}

func (p Position) ToProto() proto.Message {
	return &pschema.Position{
		Line:     int64(p.Line),
		Column:   int64(p.Column),
		Filename: p.Filename,
	}
}

func (s *Schema) ToProto() proto.Message {
	return &pschema.Schema{
		Pos:     s.Pos.ToProto().(*pschema.Position),
		Modules: nodeListToProto[*pschema.Module](s.Modules),
	}
}

func (m *Module) ToProto() proto.Message {
	return &pschema.Module{
		Pos:      m.Pos.ToProto().(*pschema.Position),
		Name:     m.Name,
		Comments: m.Comments,
		Decls:    declListToProto(m.Decls),
	}
}

func (v *Verb) ToProto() proto.Message {
	return &pschema.Verb{
		Pos:      v.Pos.ToProto().(*pschema.Position),
		Name:     v.Name,
		Comments: v.Comments,
		Request:  v.Request.ToProto().(*pschema.DataRef),
		Response: v.Response.ToProto().(*pschema.DataRef),
		Metadata: metadataListToProto(v.Metadata),
	}
}

func (d *Data) ToProto() proto.Message {
	return &pschema.Data{
		Pos:      d.Pos.ToProto().(*pschema.Position),
		Name:     d.Name,
		Fields:   nodeListToProto[*pschema.Field](d.Fields),
		Comments: d.Comments,
	}
}

func (f *Field) ToProto() proto.Message {
	return &pschema.Field{
		Pos:      f.Pos.ToProto().(*pschema.Position),
		Name:     f.Name,
		Type:     typeToProto(f.Type),
		Comments: f.Comments,
	}
}

func (v *VerbRef) ToProto() proto.Message {
	return &pschema.VerbRef{
		Pos:    v.Pos.ToProto().(*pschema.Position),
		Name:   v.Name,
		Module: v.Module,
	}
}

func (s *DataRef) ToProto() proto.Message {
	return &pschema.DataRef{
		Pos:    s.Pos.ToProto().(*pschema.Position),
		Name:   s.Name,
		Module: s.Module,
	}
}

func (m *MetadataCalls) ToProto() proto.Message {
	return &pschema.MetadataCalls{
		Pos:   m.Pos.ToProto().(*pschema.Position),
		Calls: nodeListToProto[*pschema.VerbRef](m.Calls),
	}
}

func (m *MetadataIngress) ToProto() proto.Message {
	return &pschema.MetadataIngress{
		Pos:    m.Pos.ToProto().(*pschema.Position),
		Method: m.Method,
		Path:   m.Path,
	}
}

func (i *Int) ToProto() proto.Message {
	return &pschema.Int{}
}

func (s *String) ToProto() proto.Message {
	return &pschema.String{}
}

func (b *Bool) ToProto() proto.Message {
	return &pschema.Bool{}
}

func (f *Float) ToProto() proto.Message {
	return &pschema.Float{}
}

func (t *Time) ToProto() proto.Message {
	return &pschema.Time{}
}

func (m *Map) ToProto() proto.Message {
	return &pschema.Map{
		Key:   typeToProto(m.Key),
		Value: typeToProto(m.Value),
	}
}

func (a *Array) ToProto() proto.Message {
	return &pschema.Array{
		Element: typeToProto(a.Element),
	}
}
