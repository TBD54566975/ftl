//nolint:forcetypeassert
package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
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

		case *Enum:
			v = &schemapb.Decl_Enum{Enum: n.ToProto().(*schemapb.Enum)}

		case *Config:
			v = &schemapb.Decl_Config{Config: n.ToProto().(*schemapb.Config)}

		case *Secret:
			v = &schemapb.Decl_Secret{Secret: n.ToProto().(*schemapb.Secret)}

		case *TypeAlias:
			v = &schemapb.Decl_TypeAlias{TypeAlias: n.ToProto().(*schemapb.TypeAlias)}

		case *Topic:
			v = &schemapb.Decl_Topic{Topic: n.ToProto().(*schemapb.Topic)}

		case *Subscription:
			v = &schemapb.Decl_Subscription{Subscription: n.ToProto().(*schemapb.Subscription)}
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

		case *MetadataConfig:
			v = &schemapb.Metadata_Config{Config: n.ToProto().(*schemapb.MetadataConfig)}

		case *MetadataDatabases:
			v = &schemapb.Metadata_Databases{Databases: n.ToProto().(*schemapb.MetadataDatabases)}

		case *MetadataIngress:
			v = &schemapb.Metadata_Ingress{Ingress: n.ToProto().(*schemapb.MetadataIngress)}

		case *MetadataCronJob:
			v = &schemapb.Metadata_CronJob{CronJob: n.ToProto().(*schemapb.MetadataCronJob)}

		case *MetadataAlias:
			v = &schemapb.Metadata_Alias{Alias: n.ToProto().(*schemapb.MetadataAlias)}

		case *MetadataRetry:
			v = &schemapb.Metadata_Retry{Retry: n.ToProto().(*schemapb.MetadataRetry)}

		case *MetadataSecrets:
			v = &schemapb.Metadata_Secrets{Secrets: n.ToProto().(*schemapb.MetadataSecrets)}

		case *MetadataSubscriber:
			v = &schemapb.Metadata_Subscriber{Subscriber: n.ToProto().(*schemapb.MetadataSubscriber)}

		case *MetadataTypeMap:
			v = &schemapb.Metadata_TypeMap{TypeMap: n.ToProto().(*schemapb.MetadataTypeMap)}

		case *MetadataEncoding:
			v = &schemapb.Metadata_Encoding{Encoding: n.ToProto().(*schemapb.MetadataEncoding)}

		case *MetadataPublisher:
			v = &schemapb.Metadata_Publisher{Publisher: n.ToProto().(*schemapb.MetadataPublisher)}

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

// TypeToProto creates a schemapb.Type "sum type" from a concreate Type.
func TypeToProto(t Type) *schemapb.Type {
	switch t := t.(type) {
	case *Any:
		return &schemapb.Type{Value: &schemapb.Type_Any{Any: t.ToProto().(*schemapb.Any)}}

	case *Unit:
		return &schemapb.Type{Value: &schemapb.Type_Unit{Unit: t.ToProto().(*schemapb.Unit)}}

	case *Ref:
		return &schemapb.Type{Value: &schemapb.Type_Ref{Ref: t.ToProto().(*schemapb.Ref)}}

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

func valueToProto(v Value) *schemapb.Value {
	switch t := v.(type) {
	case *StringValue:
		return &schemapb.Value{Value: &schemapb.Value_StringValue{StringValue: t.ToProto().(*schemapb.StringValue)}}

	case *IntValue:
		return &schemapb.Value{Value: &schemapb.Value_IntValue{IntValue: t.ToProto().(*schemapb.IntValue)}}

	case *TypeValue:
		return &schemapb.Value{Value: &schemapb.Value_TypeValue{TypeValue: t.ToProto().(*schemapb.TypeValue)}}
	}
	panic(fmt.Sprintf("unhandled value type: %T", v))
}
