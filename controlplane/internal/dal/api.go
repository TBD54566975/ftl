package dal

import (
	"context"
	"time"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/schema"
)

var (
	// ErrConflict is returned by select methods in the DAL when a resource already exists.
	//
	// Its use will be documented in the corresponding methods.
	ErrConflict = errors.New("conflict")
	// ErrNotFound is returned by select methods in the DAL when no results are found.
	ErrNotFound = errors.New("not found")
)

// DAL is a data access layer for the ControlPlane.
//
// This is currently a monolithic interface, but will be split into smaller
// interfaces as services are split out of the ControlPlane (eg. metrics,
// logging, etc.)
type DAL interface {
	UpsertModule(ctx context.Context, language, name string) (err error)
	// GetMissingArtefacts returns the digests of the artefacts that are missing from the database.
	GetMissingArtefacts(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error)
	// CreateArtefact inserts a new artefact into the database and returns its ID.
	CreateArtefact(ctx context.Context, content []byte) (digest sha256.SHA256, err error)
	// CreateDeployment (possibly) creates a new deployment and associates
	// previously created artefacts with it.
	//
	// If an existing deployment with identical artefacts exists, it is returned.
	CreateDeployment(ctx context.Context, language string, schema *schema.Module, artefacts []DeploymentArtefact) (key model.DeploymentKey, err error)
	GetDeployment(ctx context.Context, id model.DeploymentKey) (*model.Deployment, error)
	// UpsertRunner registers or updates a new runner.
	//
	// ErrConflict will be returned if a runner with the same endpoint and a
	// different key already exists.
	UpsertRunner(ctx context.Context, runner Runner) error
	// DeleteStaleRunners deletes runners that have not had heartbeats for the given duration.
	DeleteStaleRunners(ctx context.Context, age time.Duration) (int64, error)
	// DeregisterRunner deregisters the given runner.
	DeregisterRunner(ctx context.Context, key model.RunnerKey) error
	// ClaimRunnerForDeployment reserves a runner for the given deployment.
	//
	// Once a runner is reserved, it will be unavailable for other reservations
	// or deployments and will not be returned by GetIdleRunnersForLanguage.
	ClaimRunnerForDeployment(ctx context.Context, language string, deployment model.DeploymentKey, reservationTimeout time.Duration) (Runner, error)
	// SetDeploymentReplicas activates the given deployment.
	SetDeploymentReplicas(ctx context.Context, key model.DeploymentKey, minReplicas int) error
	// GetDeploymentsNeedingReconciliation returns deployments that have a
	// mismatch between the number of assigned and required replicas.
	GetDeploymentsNeedingReconciliation(ctx context.Context) ([]Reconciliation, error)
	// GetIdleRunnersForLanguage returns up to limit idle runners for the given language.
	//
	// If no runners are available, it will return an empty slice.
	GetIdleRunnersForLanguage(ctx context.Context, language string, limit int) ([]Runner, error)
	// GetRoutingTable returns the endpoints for all runners for the given module.
	GetRoutingTable(ctx context.Context, module string) ([]string, error)
	GetRunnerState(ctx context.Context, runnerKey model.RunnerKey) (RunnerState, error)
	// ExpireRunnerReservations and return the count.
	ExpireRunnerClaims(ctx context.Context) (int64, error)
	InsertDeploymentLogEntry(ctx context.Context, deployment model.DeploymentKey, logEntry log.Entry) error
	InsertMetricEntry(ctx context.Context, metric Metric) error
}
