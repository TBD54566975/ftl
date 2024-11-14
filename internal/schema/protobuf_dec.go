package schema

import (
	"fmt"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

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

func declListToSchema(s []*schemapb.Decl) []Decl {
	var out []Decl
	for _, n := range s {
		switch n := n.Value.(type) {
		case *schemapb.Decl_Verb:
			out = append(out, VerbFromProto(n.Verb))
		case *schemapb.Decl_Data:
			out = append(out, DataFromProto(n.Data))
		case *schemapb.Decl_Database:
			out = append(out, DatabaseFromProto(n.Database))
		case *schemapb.Decl_Enum:
			out = append(out, EnumFromProto(n.Enum))
		case *schemapb.Decl_TypeAlias:
			out = append(out, TypeAliasFromProto(n.TypeAlias))
		case *schemapb.Decl_Config:
			out = append(out, ConfigFromProto(n.Config))
		case *schemapb.Decl_Secret:
			out = append(out, SecretFromProto(n.Secret))
		case *schemapb.Decl_Topic:
			out = append(out, TopicFromProto(n.Topic))
		case *schemapb.Decl_Subscription:
			out = append(out, SubscriptionFromProto(n.Subscription))
		}
	}
	return out
}

func TypeFromProto(s *schemapb.Type) Type {
	switch s := s.Value.(type) {
	case *schemapb.Type_Ref:
		return RefFromProto(s.Ref)
	case *schemapb.Type_Int:
		return &Int{Pos: PosFromProto(s.Int.Pos)}
	case *schemapb.Type_Float:
		return &Float{Pos: PosFromProto(s.Float.Pos)}
	case *schemapb.Type_String_:
		return &String{Pos: PosFromProto(s.String_.Pos)}
	case *schemapb.Type_Bytes:
		return &Bytes{Pos: PosFromProto(s.Bytes.Pos)}
	case *schemapb.Type_Time:
		return &Time{Pos: PosFromProto(s.Time.Pos)}
	case *schemapb.Type_Bool:
		return &Bool{Pos: PosFromProto(s.Bool.Pos)}
	case *schemapb.Type_Array:
		return arrayToSchema(s.Array)
	case *schemapb.Type_Map:
		return mapToSchema(s.Map)
	case *schemapb.Type_Optional:
		return &Optional{Pos: PosFromProto(s.Optional.Pos), Type: TypeFromProto(s.Optional.Type)}
	case *schemapb.Type_Unit:
		return &Unit{Pos: PosFromProto(s.Unit.Pos)}
	case *schemapb.Type_Any:
		return &Any{Pos: PosFromProto(s.Any.Pos)}
	}
	panic(fmt.Sprintf("unhandled type: %T", s.Value))
}

func valueToSchema(v *schemapb.Value) Value {
	switch s := v.Value.(type) {
	case *schemapb.Value_IntValue:
		return &IntValue{
			Pos:   PosFromProto(s.IntValue.Pos),
			Value: int(s.IntValue.Value),
		}
	case *schemapb.Value_StringValue:
		return &StringValue{
			Pos:   PosFromProto(s.StringValue.Pos),
			Value: s.StringValue.GetValue(),
		}
	case *schemapb.Value_TypeValue:
		return &TypeValue{
			Pos:   PosFromProto(s.TypeValue.Pos),
			Value: TypeFromProto(s.TypeValue.Value),
		}
	}
	panic(fmt.Sprintf("unhandled schema value: %T", v.Value))
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
			Calls: refListToSchema(s.Calls.Calls),
		}
	case *schemapb.Metadata_Config:
		return &MetadataConfig{
			Pos:    PosFromProto(s.Config.Pos),
			Config: refListToSchema(s.Config.Config),
		}
	case *schemapb.Metadata_Databases:
		return &MetadataDatabases{
			Pos:   PosFromProto(s.Databases.Pos),
			Calls: refListToSchema(s.Databases.Calls),
		}

	case *schemapb.Metadata_Ingress:
		return &MetadataIngress{
			Pos:    PosFromProto(s.Ingress.Pos),
			Type:   s.Ingress.Type,
			Method: s.Ingress.Method,
			Path:   ingressPathComponentListToSchema(s.Ingress.Path),
		}

	case *schemapb.Metadata_CronJob:
		return &MetadataCronJob{
			Pos:  PosFromProto(s.CronJob.Pos),
			Cron: s.CronJob.Cron,
		}

	case *schemapb.Metadata_Alias:
		return &MetadataAlias{
			Pos:   PosFromProto(s.Alias.Pos),
			Kind:  AliasKind(s.Alias.Kind),
			Alias: s.Alias.Alias,
		}

	case *schemapb.Metadata_Retry:
		var count *int
		if s.Retry.Count != nil {
			countValue := int(*s.Retry.Count)
			count = &countValue
		}
		var catch *Ref
		if s.Retry.Catch != nil {
			catch = RefFromProto(s.Retry.Catch)
		}
		return &MetadataRetry{
			Pos:        PosFromProto(s.Retry.Pos),
			Count:      count,
			MinBackoff: s.Retry.MinBackoff,
			MaxBackoff: s.Retry.MaxBackoff,
			Catch:      catch,
		}

	case *schemapb.Metadata_Secrets:
		return &MetadataSecrets{
			Pos:     PosFromProto(s.Secrets.Pos),
			Secrets: refListToSchema(s.Secrets.Secrets),
		}

	case *schemapb.Metadata_Subscriber:
		return &MetadataSubscriber{
			Pos:  PosFromProto(s.Subscriber.Pos),
			Name: s.Subscriber.Name,
		}

	case *schemapb.Metadata_TypeMap:
		return &MetadataTypeMap{
			Pos:        PosFromProto(s.TypeMap.Pos),
			Runtime:    s.TypeMap.Runtime,
			NativeName: s.TypeMap.NativeName,
		}

	case *schemapb.Metadata_Encoding:
		return &MetadataEncoding{
			Pos:     PosFromProto(s.Encoding.Pos),
			Lenient: s.Encoding.Lenient,
		}

	case *schemapb.Metadata_Publisher:
		return &MetadataPublisher{
			Pos:    PosFromProto(s.Publisher.Pos),
			Topics: refListToSchema(s.Publisher.Topics),
		}

	default:
		panic(fmt.Sprintf("unhandled metadata type: %T", s))
	}
}
