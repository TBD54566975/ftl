package provisioner

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/internal/log"
)

type inMemProvisioningTask struct {
	steps []*inMemProvisioningStep
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
	Resource *provisioner.Resource
	Err      error
	Done     atomic.Bool
}

// InMemResourceProvisionerFn is a function that provisions a resource
type InMemResourceProvisionerFn func(context.Context, *provisioner.ResourceContext, string, string) (*provisioner.Resource, error)

// InMemProvisioner for running an in memory provisioner, constructing all resources concurrently
//
// It spawns a separate goroutine for each resource to be provisioned, and
// finishes the task when all resources are provisioned or an error occurs.
type InMemProvisioner struct {
	running  *xsync.MapOf[string, *inMemProvisioningTask]
	handlers map[ResourceType]InMemResourceProvisionerFn
}

func NewEmbeddedProvisioner(handlers map[ResourceType]InMemResourceProvisionerFn) *InMemProvisioner {
	return &InMemProvisioner{
		running:  xsync.NewMapOf[string, *inMemProvisioningTask](),
		handlers: handlers,
	}
}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*InMemProvisioner)(nil)

func (d *InMemProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (d *InMemProvisioner) Plan(context.Context, *connect.Request[provisioner.PlanRequest]) (*connect.Response[provisioner.PlanResponse], error) {
	panic("unimplemented")
}

func (d *InMemProvisioner) Provision(ctx context.Context, req *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	logger := log.FromContext(ctx)

	previous := map[string]*provisioner.Resource{}
	for _, r := range req.Msg.ExistingResources {
		previous[r.ResourceId] = r
	}

	task := &inMemProvisioningTask{}
	for _, r := range req.Msg.DesiredResources {
		if handler, ok := d.handlers[TypeOf(r.Resource)]; ok {
			step := &inMemProvisioningStep{Resource: r.Resource}
			task.steps = append(task.steps, step)
			go func() {
				defer step.Done.Store(true)
				output, err := handler(ctx, r, req.Msg.Module, r.Resource.ResourceId)
				if err != nil {
					step.Err = err
					logger.Errorf(err, "failed to provision resource %s", r.Resource.ResourceId)
					return
				}
				step.Resource = output
			}()
		} else {
			err := fmt.Errorf("unsupported resource type: %T", r.Resource.Resource)
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}

	token := uuid.New().String()
	logger.Debugf("started a task with token %s", token)
	d.running.Store(token, task)

	return connect.NewResponse(&provisioner.ProvisionResponse{
		ProvisioningToken: token,
		Status:            provisioner.ProvisionResponse_SUBMITTED,
	}), nil
}

func (d *InMemProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	logger := log.FromContext(ctx)

	token := req.Msg.ProvisioningToken
	task, ok := d.running.Load(token)
	if !ok {
		return statusFailure(fmt.Sprintf("unknown token: %s", token), connect.CodeNotFound)
	}
	done, err := task.Done()
	if err != nil {
		logger.Debugf("task with token %s failed with error: %s", token, err.Error())
		return statusFailure(err.Error(), connect.CodeInternal)
	}

	if !done {
		return connect.NewResponse(&provisioner.StatusResponse{
			Status: &provisioner.StatusResponse_Running{},
		}), nil
	}
	logger.Debugf("task with token %s is done", token)

	var resources []*provisioner.Resource
	for _, step := range task.steps {
		if step.Err != nil {
			return statusFailure(step.Err.Error(), connect.CodeInternal)
		}
		resources = append(resources, step.Resource)
	}
	d.running.Delete(token)

	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Success{
			Success: &provisioner.StatusResponse_ProvisioningSuccess{
				UpdatedResources: resources,
			},
		},
	}), nil
}

func statusFailure(message string, code connect.Code) (*connect.Response[provisioner.StatusResponse], error) {
	return nil, connect.NewError(code, errors.New(message))
}
