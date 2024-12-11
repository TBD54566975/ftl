package provisioner

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"

	provisioner "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/common/slices"
	"github.com/TBD54566975/ftl/internal/log"
)

type inMemProvisioningTask struct {
	steps []*inMemProvisioningStep

	events []*RuntimeEvent
}

func (t *inMemProvisioningTask) Done() (bool, error) {
	done := true
	for _, step := range t.steps {
		if !step.Done.Load() {
			done = false
		}
		if step.Err != nil {
			return false, step.Err
		}
	}
	return done, nil
}

type inMemProvisioningStep struct {
	Err  error
	Done *atomic.Value[bool]
}

// RuntimeEvent is a union type of all runtime events
// TODO: Remove once we have fully typed provisioners
type RuntimeEvent struct {
	Module   schema.ModuleRuntimeEvent
	Database *schema.DatabaseRuntimeEvent
	Topic    *schema.TopicRuntimeEvent
	Verb     *schema.VerbRuntimeEvent
}

type InMemResourceProvisionerFn func(ctx context.Context, module string, resource schema.Provisioned) (*RuntimeEvent, error)

// InMemProvisioner for running an in memory provisioner, constructing all resources concurrently
//
// It spawns a separate goroutine for each resource to be provisioned, and
// finishes the task when all resources are provisioned or an error occurs.
type InMemProvisioner struct {
	running  *xsync.MapOf[string, *inMemProvisioningTask]
	handlers map[schema.ResourceType]InMemResourceProvisionerFn
}

func NewEmbeddedProvisioner(handlers map[schema.ResourceType]InMemResourceProvisionerFn) *InMemProvisioner {
	return &InMemProvisioner{
		running:  xsync.NewMapOf[string, *inMemProvisioningTask](),
		handlers: handlers,
	}
}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*InMemProvisioner)(nil)

func (d *InMemProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (d *InMemProvisioner) Provision(ctx context.Context, req *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	logger := log.FromContext(ctx)

	var previousModule *schema.Module
	if req.Msg.PreviousModule != nil {
		pm, err := schema.ModuleFromProto(req.Msg.PreviousModule)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		previousModule = pm
	}
	desiredModule, err := schema.ModuleFromProto(req.Msg.DesiredModule)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	kinds := slices.Map(req.Msg.Kinds, func(k string) schema.ResourceType { return schema.ResourceType(k) })
	previousNodes := schema.GetProvisioned(previousModule)
	desiredNodes := schema.GetProvisioned(desiredModule)

	task := &inMemProvisioningTask{}
	for id, desired := range desiredNodes {
		previous, ok := previousNodes[id]

		for _, resource := range desired.GetProvisioned() {
			if !ok || !resource.IsEqual(previous.GetProvisioned().Get(resource.Kind)) {
				if slices.Contains(kinds, resource.Kind) {
					if handler, ok := d.handlers[resource.Kind]; ok {
						step := &inMemProvisioningStep{Done: atomic.New(false)}
						task.steps = append(task.steps, step)
						go func() {
							defer step.Done.Store(true)
							event, err := handler(ctx, desiredModule.Name, desired)
							if err != nil {
								step.Err = err
								logger.Errorf(err, "failed to provision resource %s:%s", resource.Kind, desired.ResourceID())
								return
							}
							if event != nil {
								task.events = append(task.events, event)
							}
						}()
					} else {
						err := fmt.Errorf("unsupported resource type: %s", resource.Kind)
						return nil, connect.NewError(connect.CodeInvalidArgument, err)
					}
				}
			}
		}
	}

	token := uuid.New().String()
	logger.Debugf("started a task with token %s", token)
	d.running.Store(token, task)

	return connect.NewResponse(&provisioner.ProvisionResponse{
		ProvisioningToken: token,
		Status:            provisioner.ProvisionResponse_PROVISION_RESPONSE_STATUS_SUBMITTED,
	}), nil
}

func (d *InMemProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	logger := log.FromContext(ctx)

	token := req.Msg.ProvisioningToken
	task, ok := d.running.Load(token)
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("unknown token: %s", token))
	}
	done, err := task.Done()
	if err != nil {
		logger.Debugf("task with token %s failed with error: %s", token, err.Error())
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if !done {
		return connect.NewResponse(&provisioner.StatusResponse{
			Status: &provisioner.StatusResponse_Running{},
		}), nil
	}
	logger.Debugf("task with token %s is done", token)

	d.running.Delete(token)

	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Success{
			Success: &provisioner.StatusResponse_ProvisioningSuccess{
				Events: eventsToProto(task.events),
			},
		},
	}), nil
}

func eventsToProto(events []*RuntimeEvent) []*provisioner.ProvisioningEvent {
	return slices.Map(events, func(e *RuntimeEvent) *provisioner.ProvisioningEvent {
		switch {
		case e.Database != nil:
			return &provisioner.ProvisioningEvent{Value: &provisioner.ProvisioningEvent_DatabaseRuntimeEvent{DatabaseRuntimeEvent: e.Database.ToProto()}}
		case e.Module != nil:
			switch event := e.Module.(type) {
			case *schema.ModuleRuntimeDeployment:
				return &provisioner.ProvisioningEvent{Value: &provisioner.ProvisioningEvent_ModuleRuntimeEvent{ModuleRuntimeEvent: &schemapb.ModuleRuntimeEvent{
					Value: &schemapb.ModuleRuntimeEvent_ModuleRuntimeDeployment{ModuleRuntimeDeployment: event.ToProto()},
				}}}
			case *schema.ModuleRuntimeScaling:
				return &provisioner.ProvisioningEvent{Value: &provisioner.ProvisioningEvent_ModuleRuntimeEvent{ModuleRuntimeEvent: &schemapb.ModuleRuntimeEvent{
					Value: &schemapb.ModuleRuntimeEvent_ModuleRuntimeScaling{ModuleRuntimeScaling: event.ToProto()},
				}}}
			case *schema.ModuleRuntimeBase:
				return &provisioner.ProvisioningEvent{Value: &provisioner.ProvisioningEvent_ModuleRuntimeEvent{ModuleRuntimeEvent: &schemapb.ModuleRuntimeEvent{
					Value: &schemapb.ModuleRuntimeEvent_ModuleRuntimeBase{ModuleRuntimeBase: event.ToProto()},
				}}}
			default:
				panic("unknown module event type")
			}
		case e.Topic != nil:
			return &provisioner.ProvisioningEvent{Value: &provisioner.ProvisioningEvent_TopicRuntimeEvent{TopicRuntimeEvent: e.Topic.ToProto()}}
		case e.Verb != nil:
			return &provisioner.ProvisioningEvent{Value: &provisioner.ProvisioningEvent_VerbRuntimeEvent{VerbRuntimeEvent: e.Verb.ToProto()}}
		default:
			panic("unknown event type")
		}
	})
}
