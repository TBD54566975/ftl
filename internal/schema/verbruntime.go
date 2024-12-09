package schema

import (
	"time"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/slices"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type VerbRuntime struct {
	Base         VerbRuntimeBase          `protobuf:"1"`
	Subscription *VerbRuntimeSubscription `protobuf:"2,optional"`
}

//sumtype:decl
type VerbRuntimeEvent struct {
	ID      string             `protobuf:"1"`
	Payload VerbRuntimePayload `protobuf:"2"`
}

func (v *VerbRuntimeEvent) ApplyTo(m *Module) {
	for verb := range slices.FilterVariants[*Verb](m.Decls) {
		if verb.Name == v.ID {
			switch payload := v.Payload.(type) {
			case *VerbRuntimeBase:
				verb.Runtime.Base = *payload
			case *VerbRuntimeSubscription:
				verb.Runtime.Subscription = payload
			}
		}
	}
}

func (v *VerbRuntimeEvent) ToProto() protoreflect.ProtoMessage {
	return &schemapb.VerbRuntimeEvent{
		Id:      v.ID,
		Payload: v.Payload.ToProto().(*schemapb.VerbRuntimePayload),
	}
}

func VerbRuntimeEventFromProto(p *schemapb.VerbRuntimeEvent) *VerbRuntimeEvent {
	return &VerbRuntimeEvent{
		ID:      p.Id,
		Payload: VerbRuntimePayloadFromProto(p.Payload),
	}
}

//sumtype:decl
type VerbRuntimePayload interface {
	verbRuntime()
	ToProto() protoreflect.ProtoMessage
}

func VerbRuntimePayloadFromProto(p *schemapb.VerbRuntimePayload) VerbRuntimePayload {
	switch p.Value.(type) {
	case *schemapb.VerbRuntimePayload_VerbRuntimeBase:
		return VerbRuntimeBaseFromProto(p.GetVerbRuntimeBase())
	case *schemapb.VerbRuntimePayload_VerbRuntimeSubscription:
		return VerbRuntimeSubscriptionFromProto(p.GetVerbRuntimeSubscription())
	default:
		panic("unknown verb runtime payload type")
	}
}

//protobuf:1
type VerbRuntimeBase struct {
	CreateTime *time.Time `protobuf:"1,optional"`
	StartTime  *time.Time `protobuf:"2,optional"`
}

func (*VerbRuntimeBase) verbRuntime() {}

func (v *VerbRuntimeBase) ToProto() protoreflect.ProtoMessage {
	return &schemapb.VerbRuntimeBase{
		CreateTime: timestampToProto(v.CreateTime),
		StartTime:  timestampToProto(v.StartTime),
	}
}

func VerbRuntimeBaseFromProto(s *schemapb.VerbRuntimeBase) *VerbRuntimeBase {
	if s == nil {
		return nil
	}
	return &VerbRuntimeBase{
		CreateTime: timestampFromProto(s.CreateTime),
		StartTime:  timestampFromProto(s.StartTime),
	}
}

//protobuf:2
type VerbRuntimeSubscription struct {
	KafkaBrokers []string `protobuf:"1,optional"`
}

func (*VerbRuntimeSubscription) verbRuntime() {}

func (v *VerbRuntimeSubscription) ToProto() protoreflect.ProtoMessage {
	if v == nil {
		return nil
	}
	return &schemapb.VerbRuntimeSubscription{
		KafkaBrokers: v.KafkaBrokers,
	}
}

func VerbRuntimeSubscriptionFromProto(s *schemapb.VerbRuntimeSubscription) *VerbRuntimeSubscription {
	if s == nil {
		return nil
	}
	return &VerbRuntimeSubscription{
		KafkaBrokers: s.KafkaBrokers,
	}
}

func timestampToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func timestampFromProto(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	time := t.AsTime()
	return &time
}
