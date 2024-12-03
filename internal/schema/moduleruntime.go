package schema

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

var _ Runtime = (*ModuleRuntime)(nil)

// ModuleRuntime is the base runtime for a module, plus any additional dynamic runtime configuration.
type ModuleRuntime struct {
	CreateTime time.Time `protobuf:"1"`
	Language   string    `protobuf:"2"`
	OS         string    `protobuf:"3,optional"`
	Arch       string    `protobuf:"4,optional"`
	// Image is the name of the runner image. Defaults to "ftl0/ftl-runner".
	// Must not include a tag, as FTL's version will be used as the tag.
	Image string `protobuf:"5,optional"`

	// Configuration that can be dynamically added to a module and updated after creation.
	Scaling    *ModuleRuntimeScaling    `protobuf:"7,optional"`
	Deployment *ModuleRuntimeDeployment `protobuf:"8,optional"`
}

func (*ModuleRuntime) runtime() {}

// ApplyEvent applies a ModuleRuntimeEvent to the ModuleRuntime.
func (m *ModuleRuntime) ApplyEvent(event ModuleRuntimeEvent) {
	switch event := event.(type) {
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
		CreateTime: s.GetCreateTime().AsTime(),
		Language:   s.Language,
		OS:         s.GetOs(),
		Arch:       s.GetArch(),
		Image:      s.GetImage(),
		Scaling:    ModuleRuntimeScalingFromProto(s.Scaling),
		Deployment: ModuleRuntimeDeploymentFromProto(s.Deployment),
	}
}

func (m *ModuleRuntime) ToProto() protoreflect.ProtoMessage {
	out := &schemapb.ModuleRuntime{
		CreateTime: timestamppb.New(m.CreateTime),
		Language:   m.Language,
	}
	if m.OS != "" {
		out.Os = &m.OS
	}
	if m.Arch != "" {
		out.Arch = &m.Arch
	}
	if m.Image != "" {
		out.Image = &m.Image
	}
	return out
}

func ModuleRuntimeEventFromProto(s *schemapb.ModuleRuntimeEvent) ModuleRuntimeEvent {
	switch s.Value.(type) {
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
	moduleRuntime()
}

//protobuf:1
type ModuleRuntimeScaling struct {
	MinReplicas int32 `protobuf:"1"`
}

func (*ModuleRuntimeScaling) moduleRuntime() {}

func ModuleRuntimeScalingFromProto(s *schemapb.ModuleRuntimeScaling) *ModuleRuntimeScaling {
	if s == nil {
		return nil
	}
	return &ModuleRuntimeScaling{
		MinReplicas: s.MinReplicas,
	}
}

func (m *ModuleRuntimeScaling) ToProto() *schemapb.ModuleRuntimeScaling {
	if m == nil {
		return nil
	}
	return &schemapb.ModuleRuntimeScaling{
		MinReplicas: m.MinReplicas,
	}
}

//protobuf:2
type ModuleRuntimeDeployment struct {
	// Endpoint is the endpoint of the deployed module.
	Endpoint      string `protobuf:"1"`
	DeploymentKey string `protobuf:"2"`
}

func (m *ModuleRuntimeDeployment) moduleRuntime() {}

func ModuleRuntimeDeploymentFromProto(s *schemapb.ModuleRuntimeDeployment) *ModuleRuntimeDeployment {
	if s == nil {
		return nil
	}
	return &ModuleRuntimeDeployment{
		Endpoint:      s.Endpoint,
		DeploymentKey: s.DeploymentKey,
	}
}

func (m *ModuleRuntimeDeployment) ToProto() *schemapb.ModuleRuntimeDeployment {
	if m == nil {
		return nil
	}
	return &schemapb.ModuleRuntimeDeployment{
		Endpoint:      m.Endpoint,
		DeploymentKey: m.DeploymentKey,
	}
}
