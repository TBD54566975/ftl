package deployment

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
)

type TaskState string

const (
	TaskStatePending TaskState = ""
	TaskStateRunning TaskState = "running"
	TaskStateDone    TaskState = "done"
	TaskStateFailed  TaskState = "failed"
)

// Task is a unit of work for a deployment
type Task struct {
	handler  provisionerconnect.ProvisionerPluginServiceClient
	module   string
	state    TaskState
	desired  []*provisioner.Resource
	existing []*provisioner.Resource
	// populated only when the task is done
	output []*provisioner.Resource

	// set if the task is currently running
	runningToken string
}

func (t *Task) Start(ctx context.Context) error {
	if t.state != TaskStatePending {
		return fmt.Errorf("task state is not pending: %s", t.state)
	}
	t.state = TaskStateRunning

	resp, err := t.handler.Provision(ctx, connect.NewRequest(&provisioner.ProvisionRequest{
		Module: t.module,
		// TODO: We need a proper cluster speicific ID here
		FtlClusterId:      "ftl",
		ExistingResources: t.existing,
		DesiredResources:  t.constructResourceContext(t.desired),
	}))
	if err != nil {
		t.state = TaskStateFailed
		return fmt.Errorf("error provisioning resources: %w", err)
	}
	if resp.Msg.Status == provisioner.ProvisionResponse_NO_CHANGES {
		t.state = TaskStateDone
		t.output = t.desired
	}
	t.runningToken = resp.Msg.ProvisioningToken
	return nil
}

func (t *Task) constructResourceContext(r []*provisioner.Resource) []*provisioner.ResourceContext {
	result := make([]*provisioner.ResourceContext, len(r))
	for i, res := range r {
		result[i] = &provisioner.ResourceContext{
			Resource: res,
			// TODO: Collect previously constructed resources from a dependency graph here
		}
	}
	return result
}

func (t *Task) Progress(ctx context.Context) error {
	if t.state != TaskStateRunning {
		return fmt.Errorf("task state is not running: %s", t.state)
	}

	resp, err := t.handler.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
		ProvisioningToken: t.runningToken,
		DesiredResources:  t.desired,
	}))
	if err != nil {
		return fmt.Errorf("error getting state: %w", err)
	}
	if succ, ok := resp.Msg.Status.(*provisioner.StatusResponse_Success); ok {
		t.state = TaskStateDone
		t.output = succ.Success.UpdatedResources
	}
	if fail, ok := resp.Msg.Status.(*provisioner.StatusResponse_Failed); ok {
		t.state = TaskStateFailed
		return errors.New(fail.Failed.ErrorMessage)
	}
	return nil
}

// Deployment is a single deployment of resources for a single module
type Deployment struct {
	Module string
	Tasks  []*Task
}

// next running or pending task. Nil if all tasks are done.
func (d *Deployment) next() optional.Option[*Task] {
	for _, t := range d.Tasks {
		if t.state == TaskStatePending || t.state == TaskStateRunning || t.state == TaskStateFailed {
			return optional.Some(t)
		}
	}
	return optional.None[*Task]()
}

// Progress the deployment. Returns true if there are still tasks running or pending.
func (d *Deployment) Progress(ctx context.Context) (bool, error) {
	next, ok := d.next().Get()
	if !ok {
		return false, nil
	}

	if next.state == TaskStatePending {
		err := next.Start(ctx)
		if err != nil {
			return true, err
		}
	}
	if next.state != TaskStateDone {
		err := next.Progress(ctx)
		if err != nil {
			return true, err
		}
	}
	return d.next().Ok(), nil
}

type DeploymentState struct {
	Pending []*Task
	Running *Task
	Failed  *Task
	Done    []*Task
}

func (d *Deployment) State() *DeploymentState {
	result := &DeploymentState{}
	for _, t := range d.Tasks {
		switch t.state {
		case TaskStatePending:
			result.Pending = append(result.Pending, t)
		case TaskStateRunning:
			result.Running = t
		case TaskStateFailed:
			result.Failed = t
		case TaskStateDone:
			result.Done = append(result.Done, t)
		}
	}
	return result
}
