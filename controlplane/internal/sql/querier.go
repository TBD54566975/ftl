// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package sql

import (
	"context"

	"github.com/TBD54566975/ftl/controlplane/internal/sqltypes"
	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	AssociateArtefactWithDeployment(ctx context.Context, arg AssociateArtefactWithDeploymentParams) error
	// Find an idle runner and claim it for the given deployment.
	ClaimRunner(ctx context.Context, language string, reservationTimeout pgtype.Timestamptz, deploymentKey sqltypes.Key) (Runner, error)
	// Create a new artefact and return the artefact ID.
	CreateArtefact(ctx context.Context, digest []byte, content []byte) (int64, error)
	CreateDeployment(ctx context.Context, key sqltypes.Key, moduleName string, schema []byte) error
	DeleteStaleRunners(ctx context.Context, dollar_1 pgtype.Interval) (int64, error)
	DeregisterRunner(ctx context.Context, key sqltypes.Key) (int64, error)
	ExpireRunnerReservations(ctx context.Context) (int64, error)
	GetArtefactContentRange(ctx context.Context, start int32, count int32, iD int64) ([]byte, error)
	// Return the digests that exist in the database.
	GetArtefactDigests(ctx context.Context, digests [][]byte) ([]GetArtefactDigestsRow, error)
	GetDeployment(ctx context.Context, key sqltypes.Key) (GetDeploymentRow, error)
	// Get all artefacts matching the given digests.
	GetDeploymentArtefacts(ctx context.Context, deploymentID int64) ([]GetDeploymentArtefactsRow, error)
	GetDeploymentsByID(ctx context.Context, ids []int64) ([]Deployment, error)
	// Get deployments that have a mismatch between the number of assigned and required replicas.
	GetDeploymentsNeedingReconciliation(ctx context.Context) ([]GetDeploymentsNeedingReconciliationRow, error)
	// Get all deployments that have artefacts matching the given digests.
	GetDeploymentsWithArtefacts(ctx context.Context, digests [][]byte, count interface{}) ([]GetDeploymentsWithArtefactsRow, error)
	GetIdleRunnerCountsByLanguage(ctx context.Context) ([]GetIdleRunnerCountsByLanguageRow, error)
	GetIdleRunnersForLanguage(ctx context.Context, language string, limit int32) ([]Runner, error)
	GetModulesByID(ctx context.Context, ids []int64) ([]Module, error)
	GetRoutingTable(ctx context.Context, name string) ([]string, error)
	GetRunnerState(ctx context.Context, key sqltypes.Key) (RunnerState, error)
	GetRunnersForDeployment(ctx context.Context, key sqltypes.Key) ([]Runner, error)
	InsertDeploymentLogEntry(ctx context.Context, arg InsertDeploymentLogEntryParams) error
	InsertMetricEntry(ctx context.Context, arg InsertMetricEntryParams) error
	SetDeploymentDesiredReplicas(ctx context.Context, key sqltypes.Key, minReplicas int32) error
	UpsertModule(ctx context.Context, language string, name string) (int64, error)
	// Upsert a runner and return the deployment ID that it is assigned to, if any.
	// If the deployment key is null, then deployment_rel.id will be null,
	// otherwise we try to retrieve the deployments.id using the key. If
	// there is no corresponding deployment, then the deployment ID is -1
	// and the parent statement will fail due to a foreign key constraint.
	UpsertRunner(ctx context.Context, arg UpsertRunnerParams) (pgtype.Int8, error)
}

var _ Querier = (*Queries)(nil)
