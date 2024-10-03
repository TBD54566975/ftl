package dev

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"sync/atomic"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/internal/dev"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
	"github.com/XSAM/otelsql"
	"github.com/puzpuzpuz/xsync/v3"
)

type task struct {
	steps []*step
}

func (t *task) Done() bool {
	for _, step := range t.steps {
		if !step.done.Load() {
			return false
		}
	}
	return true
}

type step struct {
	resource *provisioner.Resource
	err      error
	done     atomic.Bool
}

// Provisioner for running FTL locally
type Provisioner struct {
	running     *xsync.MapOf[string, *task]
	postgresDSN string

	postgresPort int
}

func NewProvisioner(postgresPort int) *Provisioner {
	return &Provisioner{
		postgresPort: postgresPort,
		running:      xsync.NewMapOf[string, *task](),
	}
}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*Provisioner)(nil)

func (d *Provisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (d *Provisioner) Plan(context.Context, *connect.Request[provisioner.PlanRequest]) (*connect.Response[provisioner.PlanResponse], error) {
	panic("unimplemented")
}

func (d *Provisioner) Provision(ctx context.Context, req *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	previous := map[string]*provisioner.Resource{}
	for _, r := range req.Msg.ExistingResources {
		previous[r.ResourceId] = r
	}

	task := &task{}
	for _, r := range req.Msg.DesiredResources {
		if _, ok := previous[r.Resource.ResourceId]; !ok {
			switch tr := r.Resource.Resource.(type) {
			case *provisioner.Resource_Postgres:
				step := &step{resource: r.Resource}
				task.steps = append(task.steps, step)

				d.provisionPostgres(ctx, tr, req.Msg.Module, r.Resource.ResourceId, step)
			default:
				err := fmt.Errorf("unsupported resource type: %T", r.Resource.Resource)
				return nil, connect.NewError(connect.CodeInvalidArgument, err)
			}
		}
	}

	if len(task.steps) == 0 {
		return connect.NewResponse(&provisioner.ProvisionResponse{
			Status: provisioner.ProvisionResponse_NO_CHANGES,
		}), nil
	}

	token := strconv.Itoa(rand.Int()) //nolint:gosec
	d.running.Store(token, task)

	return connect.NewResponse(&provisioner.ProvisionResponse{
		ProvisioningToken: token,
		Status:            provisioner.ProvisionResponse_SUBMITTED,
	}), nil
}

func (d *Provisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	token := req.Msg.ProvisioningToken
	task, ok := d.running.Load(token)
	if !ok {
		return statusFailure("unknown token")
	}
	if !task.Done() {
		return connect.NewResponse(&provisioner.StatusResponse{
			Status: &provisioner.StatusResponse_Running{},
		}), nil
	}

	var resources []*provisioner.Resource
	for _, step := range task.steps {
		if step.err != nil {
			return statusFailure(step.err.Error())
		}
		resources = append(resources, step.resource)
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

func statusFailure(message string) (*connect.Response[provisioner.StatusResponse], error) {
	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Failed{
			Failed: &provisioner.StatusResponse_ProvisioningFailed{
				ErrorMessage: message,
			},
		},
	}), nil
}

func (d *Provisioner) provisionPostgres(ctx context.Context, tr *provisioner.Resource_Postgres, module, id string, step *step) {
	logger := log.FromContext(ctx)
	logger.Infof("provisioning postgres database: %s_%s", module, id)

	go func() {
		defer step.done.Store(true)
		if d.postgresDSN == "" {
			// We assume that the DB hsas already been started when running in dev mode
			dsn, err := dev.WaitForDBReady(ctx, d.postgresPort)
			if err != nil {
				step.err = err
				return
			}
			d.postgresDSN = dsn
		}
		dbName := strcase.ToLowerSnake(module) + "_" + strcase.ToLowerSnake(id)
		conn, err := otelsql.Open("pgx", d.postgresDSN)
		if err != nil {
			step.err = err
			return
		}
		defer conn.Close()

		res, err := conn.Query("SELECT * FROM pg_catalog.pg_database WHERE datname=$1", dbName)
		if err != nil {
			step.err = err
			return
		}
		defer res.Close()
		if !res.Next() {
			_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
			if err != nil {
				step.err = err
				return
			}
		}

		if tr.Postgres == nil {
			tr.Postgres = &provisioner.PostgresResource{}
		}
		tr.Postgres.Output = &provisioner.PostgresResource_PostgresResourceOutput{
			ReadEndpoint:  d.postgresDSN,
			WriteEndpoint: d.postgresDSN,
		}
	}()
}
