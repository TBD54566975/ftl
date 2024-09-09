package infra

import (
	"context"
	"errors"
)

type ResourceID struct {
	Kind       string
	Module     string
	Deployment string
}

type Resource interface {
	Id() ResourceID
	Plan(ctx context.Context) (*OperationGraph, error)
}

// FTL is the root of the global resource graph
type FTL struct{}

func (r *FTL) Id() ResourceID {
	return ResourceID{Kind: "FTL"}
}

func (r *FTL) Plan(ctx context.Context) (*OperationGraph, error) {
	return nil, errors.New("can not add new FTL instances")
}

// Deployment is a temporary resource as a root for resources created during deployments
type Deployment struct {
	id string
}

func (r *Deployment) Id() ResourceID {
	return ResourceID{Kind: "Deployment", Deployment: r.id}
}

func (r *Deployment) Plan(ctx context.Context) (*OperationGraph, error) {
	panic("Implement")
}

// Module is where the user code lives
type Module struct {
	name         string
	deploymentId string
}

func NewModule(name, deploymentId string) *Module {
	return &Module{name: name, deploymentId: deploymentId}
}

func (r *Module) Id() ResourceID {
	return ResourceID{Kind: "Module", Deployment: r.deploymentId, Module: r.name}
}

func (r *Module) Plan(ctx context.Context) (*OperationGraph, error) {
	panic("Implement")
}
