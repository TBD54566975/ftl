package deployment

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
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
	Handler  Provisioner
	Module   string
	State    TaskState
	Desired  []*provisioner.Resource
	Existing []*provisioner.Resource
	// populated only when the task is done
	Output []*provisioner.Resource

	// set if the task is currently running
	RunningToken string
}

func (t *Task) Start(ctx context.Context) error {
	if t.State != TaskStatePending {
		return fmt.Errorf("task state is not pending: %s", t.State)
	}
	t.State = TaskStateRunning
	token, err := t.Handler.Provision(ctx, t.Module, t.constructResourceContext(t.Desired), t.Existing)
	if err != nil {
		t.State = TaskStateFailed
		return fmt.Errorf("error provisioning resources: %w", err)
	}
	if token == "" {
		// no changes
		t.State = TaskStateDone
		t.Output = t.Desired
	}
	t.RunningToken = token
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
	if t.State != TaskStateRunning {
		return fmt.Errorf("task state is not running: %s", t.State)
	}
	state, output, err := t.Handler.State(ctx, t.RunningToken, t.Desired)
	if err != nil {
		return fmt.Errorf("error getting state: %w", err)
	}
	if state == TaskStateDone {
		t.State = TaskStateDone
		t.Output = output
	}
	return nil
}

// Deployment is a single deployment of resources for a single module
type Deployment struct {
	Module string
	Tasks  []*Task
}

// next running or pending task. Nil if all tasks are done.
func (d *Deployment) next() *Task {
	for _, t := range d.Tasks {
		if t.State == TaskStatePending || t.State == TaskStateRunning || t.State == TaskStateFailed {
			return t
		}
	}
	return nil
}

// Progress the deployment. Returns true if there are still tasks running or pending.
func (d *Deployment) Progress(ctx context.Context) (bool, error) {
	next := d.next()
	if next == nil {
		return false, nil
	}

	if next.State == TaskStatePending {
		err := next.Start(ctx)
		if err != nil {
			return true, err
		}
	}
	err := next.Progress(ctx)
	return d.next() != nil, err
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
		switch t.State {
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
