package schema

import (
	"fmt"
	"time"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/slices"
)

type VerbRuntime struct {
	Base         VerbRuntimeBase          `protobuf:"1"`
	Subscription *VerbRuntimeSubscription `protobuf:"2,optional"`
}

type VerbRuntimeEvent struct {
	ID      string             `protobuf:"1"`
	Payload VerbRuntimePayload `protobuf:"2"`
}

func (v *VerbRuntimeEvent) ApplyTo(m *Module) {
	for verb := range slices.FilterVariants[*Verb](m.Decls) {
		if verb.Name == v.ID {
			if verb.Runtime == nil {
				verb.Runtime = &VerbRuntime{}
			}

			switch payload := v.Payload.(type) {
			case *VerbRuntimeBase:
				verb.Runtime.Base = *payload
			case *VerbRuntimeSubscription:
				verb.Runtime.Subscription = payload
			default:
				panic(fmt.Sprintf("unknown verb runtime payload type: %T", payload))
			}
		}
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
	CreateTime time.Time `protobuf:"1,optional"`
	StartTime  time.Time `protobuf:"2,optional"`
}

func (*VerbRuntimeBase) verbRuntime() {}

func VerbRuntimeBaseFromProto(s *schemapb.VerbRuntimeBase) *VerbRuntimeBase {
	if s == nil {
		return &VerbRuntimeBase{}
	}
	return &VerbRuntimeBase{
		CreateTime: s.CreateTime.AsTime(),
		StartTime:  s.StartTime.AsTime(),
	}
}

//protobuf:2
type VerbRuntimeSubscription struct {
	KafkaBrokers []string `protobuf:"1"`
}

func (*VerbRuntimeSubscription) verbRuntime() {}

func VerbRuntimeSubscriptionFromProto(s *schemapb.VerbRuntimeSubscription) *VerbRuntimeSubscription {
	if s == nil {
		return nil
	}
	return &VerbRuntimeSubscription{
		KafkaBrokers: s.KafkaBrokers,
	}
}
