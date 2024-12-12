package schema

import (
	"fmt"
	"time"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

// ModuleRuntime is runtime configuration for a module that can be dynamically updated.
type ModuleRuntime struct {
	Base       ModuleRuntimeBase        `protobuf:"1"` // Base is always present.
	Scaling    *ModuleRuntimeScaling    `protobuf:"2,optional"`
	Deployment *ModuleRuntimeDeployment `protobuf:"3,optional"`
}

// ApplyEvent applies a ModuleRuntimeEvent to the ModuleRuntime.
func (m *ModuleRuntime) ApplyEvent(event ModuleRuntimeEvent) {
	switch event := event.(type) {
	case *ModuleRuntimeBase:
		m.Base = *event
	case *ModuleRuntimeScaling:
		m.Scaling = event
	case *ModuleRuntimeDeployment:
		m.Deployment = event
	}
}

func ModuleRuntimeFromProto(s *schemapb.ModuleRuntime) *ModuleRuntime {
	if s == nil {
		return nil
	}
	return &ModuleRuntime{
		Base:       *ModuleRuntimeBaseFromProto(s.Base),
		Scaling:    ModuleRuntimeScalingFromProto(s.Scaling),
		Deployment: ModuleRuntimeDeploymentFromProto(s.Deployment),
	}
}

func ModuleRuntimeEventFromProto(s *schemapb.ModuleRuntimeEvent) ModuleRuntimeEvent {
	switch s.Value.(type) {
	case *schemapb.ModuleRuntimeEvent_ModuleRuntimeBase:
		return ModuleRuntimeBaseFromProto(s.GetModuleRuntimeBase())

	case *schemapb.ModuleRuntimeEvent_ModuleRuntimeScaling:
		return ModuleRuntimeScalingFromProto(s.GetModuleRuntimeScaling())

	case *schemapb.ModuleRuntimeEvent_ModuleRuntimeDeployment:
		return ModuleRuntimeDeploymentFromProto(s.GetModuleRuntimeDeployment())

	default:
		panic(fmt.Errorf("unknown ModuleRuntimeEvent variant %T", s.Value))
	}
}

//sumtype:decl
type ModuleRuntimeEvent interface {
	RuntimeEvent

	moduleRuntime()
}

//protobuf:1
//protobuf:1 RuntimeEvent
type ModuleRuntimeBase struct {
	CreateTime time.Time `protobuf:"1"`
	Language   string    `protobuf:"2"`
	OS         string    `protobuf:"3,optional"`
	Arch       string    `protobuf:"4,optional"`
	// Image is the name of the runner image. Defaults to "ftl0/ftl-runner".
	// Must not include a tag, as FTL's version will be used as the tag.
	Image string `protobuf:"5,optional"`
}

func (ModuleRuntimeBase) moduleRuntime() {}

func (m *ModuleRuntimeBase) runtimeEvent() {}
func ModuleRuntimeBaseFromProto(s *schemapb.ModuleRuntimeBase) *ModuleRuntimeBase {
	if s == nil {
		return &ModuleRuntimeBase{}
	}
	return &ModuleRuntimeBase{
		CreateTime: s.GetCreateTime().AsTime(),
		Language:   s.GetLanguage(),
		OS:         s.GetOs(),
		Arch:       s.GetArch(),
		Image:      s.GetImage(),
	}
}

//protobuf:2
//protobuf:2 RuntimeEvent
type ModuleRuntimeScaling struct {
	MinReplicas int32 `protobuf:"1"`
}

func (*ModuleRuntimeScaling) moduleRuntime() {}

func (m *ModuleRuntimeScaling) runtimeEvent() {}
func ModuleRuntimeScalingFromProto(s *schemapb.ModuleRuntimeScaling) *ModuleRuntimeScaling {
	if s == nil {
		return nil
	}
	return &ModuleRuntimeScaling{
		MinReplicas: s.MinReplicas,
	}
}

//protobuf:3
//protobuf:3 RuntimeEvent
type ModuleRuntimeDeployment struct {
	// Endpoint is the endpoint of the deployed module.
	Endpoint      string `protobuf:"1"`
	DeploymentKey string `protobuf:"2"`
}

func (m *ModuleRuntimeDeployment) moduleRuntime() {}

func (m *ModuleRuntimeDeployment) runtimeEvent() {}

func ModuleRuntimeDeploymentFromProto(s *schemapb.ModuleRuntimeDeployment) *ModuleRuntimeDeployment {
	if s == nil {
		return nil
	}
	return &ModuleRuntimeDeployment{
		Endpoint:      s.Endpoint,
		DeploymentKey: s.DeploymentKey,
	}
}
