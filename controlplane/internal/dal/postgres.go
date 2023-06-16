// Package dal provides a data abstraction layer for the ControlPlane
package dal

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	sets "github.com/deckarep/golang-set/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/controlplane/internal/sql"
	"github.com/TBD54566975/ftl/controlplane/internal/sqltypes"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/slices"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(conn sql.DBI) *Postgres {
	return &Postgres{db: sql.NewDB(conn)}
}

func (d *Postgres) UpsertModule(ctx context.Context, language, name string) (err error) {
	_, err = d.db.UpsertModule(ctx, language, name)
	return errors.WithStack(translatePGError(err))
}

// GetMissingArtefacts returns the digests of the artefacts that are missing from the database.
func (d *Postgres) GetMissingArtefacts(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error) {
	have, err := d.db.GetArtefactDigests(ctx, sha256esToBytes(digests))
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	haveStr := slices.Map(have, func(in sql.GetArtefactDigestsRow) sha256.SHA256 {
		return sha256.FromBytes(in.Digest)
	})
	return sets.NewSet(digests...).Difference(sets.NewSet(haveStr...)).ToSlice(), nil
}

// CreateArtefact inserts a new artefact into the database and returns its ID.
func (d *Postgres) CreateArtefact(ctx context.Context, content []byte) (digest sha256.SHA256, err error) {
	sha256digest := sha256.Sum(content)
	_, err = d.db.CreateArtefact(ctx, sha256digest[:], content)
	return sha256digest, errors.WithStack(translatePGError(err))
}

type DeploymentArtefact struct {
	Digest     sha256.SHA256
	Executable bool
	Path       string
}

func (d *DeploymentArtefact) ToProto() *ftlv1.DeploymentArtefact {
	return &ftlv1.DeploymentArtefact{
		Digest:     d.Digest.String(),
		Executable: d.Executable,
		Path:       d.Path,
	}
}

func DeploymentArtefactFromProto(in *ftlv1.DeploymentArtefact) (DeploymentArtefact, error) {
	digest, err := sha256.ParseSHA256(in.Digest)
	if err != nil {
		return DeploymentArtefact{}, errors.WithStack(err)
	}
	return DeploymentArtefact{
		Digest:     digest,
		Executable: in.Executable,
		Path:       in.Path,
	}, nil
}

// CreateDeployment (possibly) creates a new deployment and associates
// previously created artefacts with it.
//
// If an existing deployment with identical artefacts exists, it is returned.
func (d *Postgres) CreateDeployment(ctx context.Context, language string, schema *schema.Module, artefacts []DeploymentArtefact) (key model.DeploymentKey, err error) {
	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return model.DeploymentKey{}, errors.WithStack(err)
	}
	defer tx.CommitOrRollback(ctx, &err)()

	// Check if the deployment already exists and if so, return it.
	existing, err := tx.GetDeploymentsWithArtefacts(ctx,
		sha256esToBytes(slices.Map(artefacts, func(in DeploymentArtefact) sha256.SHA256 { return in.Digest })),
		len(artefacts),
	)
	if err != nil {
		return model.DeploymentKey{}, errors.WithStack(err)
	}
	if len(existing) > 0 {
		return model.DeploymentKey(existing[0].Key), nil
	}

	artefactsByDigest := maps.FromSlice(artefacts, func(in DeploymentArtefact) (sha256.SHA256, DeploymentArtefact) {
		return in.Digest, in
	})

	schemaBytes, err := proto.Marshal(schema.ToProto())
	if err != nil {
		return model.DeploymentKey{}, errors.WithStack(err)
	}

	// TODO(aat): "schema" containing language?
	_, err = tx.UpsertModule(ctx, language, schema.Name)

	deploymentKey := ulid.Make()
	// Create the deployment
	err = tx.CreateDeployment(ctx, sqltypes.Key(deploymentKey), schema.Name, schemaBytes)
	if err != nil {
		return model.DeploymentKey{}, errors.WithStack(translatePGError(err))
	}

	uploadedDigests := slices.Map(artefacts, func(in DeploymentArtefact) []byte { return in.Digest[:] })
	artefactDigests, err := tx.GetArtefactDigests(ctx, uploadedDigests)
	if err != nil {
		return model.DeploymentKey{}, errors.WithStack(err)
	}
	if len(artefactDigests) != len(artefacts) {
		missingDigests := strings.Join(slices.Map(artefacts, func(in DeploymentArtefact) string { return in.Digest.String() }), ", ")
		return model.DeploymentKey{}, errors.Errorf("missing %d artefacts: %s", len(artefacts)-len(artefactDigests), missingDigests)
	}

	// Associate the artefacts with the deployment
	for _, row := range artefactDigests {
		artefact := artefactsByDigest[sha256.FromBytes(row.Digest)]
		err = tx.AssociateArtefactWithDeployment(ctx, sql.AssociateArtefactWithDeploymentParams{
			Key:        sqltypes.Key(deploymentKey),
			ArtefactID: row.ID,
			Executable: artefact.Executable,
			Path:       artefact.Path,
		})
		if err != nil {
			return model.DeploymentKey{}, errors.WithStack(translatePGError(err))
		}
	}

	return model.DeploymentKey(deploymentKey), nil
}

func (d *Postgres) GetDeployment(ctx context.Context, id model.DeploymentKey) (*model.Deployment, error) {
	deployment, err := d.db.GetDeployment(ctx, sqltypes.Key(id))
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	return d.loadDeployment(ctx, deployment)
}

// UpsertRunner registers or updates a new runner.
//
// ErrConflict will be returned if a runner with the same endpoint and a
// different key already exists.
func (d *Postgres) UpsertRunner(ctx context.Context, runner Runner) error {
	var pgDeploymentKey pgtype.UUID
	if dkey, ok := runner.Deployment.Get(); ok {
		pgDeploymentKey.Valid = true
		pgDeploymentKey.Bytes = dkey
	}
	deploymentID, err := d.db.UpsertRunner(ctx, sql.UpsertRunnerParams{
		Key:           sqltypes.Key(runner.Key),
		Language:      runner.Language,
		Endpoint:      runner.Endpoint,
		State:         sql.RunnerState(runner.State),
		DeploymentKey: pgDeploymentKey,
	})
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	if runner.Deployment.Ok() && !deploymentID.Valid {
		return errors.Errorf("deployment %s not found", runner.Deployment)
	}
	return nil
}

// DeleteStaleRunners deletes runners that have not had heartbeats for the given duration.
func (d *Postgres) DeleteStaleRunners(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.DeleteStaleRunners(ctx, pgtype.Interval{
		Microseconds: int64(age / time.Microsecond),
		Valid:        true,
	})
	return count, errors.WithStack(err)
}

// DeregisterRunner deregisters the given runner.
func (d *Postgres) DeregisterRunner(ctx context.Context, key model.RunnerKey) error {
	count, err := d.db.DeregisterRunner(ctx, sqltypes.Key(key))
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

type Runner struct {
	Key      model.RunnerKey
	Language string
	Endpoint string
	State    RunnerState
	// Assigned deployment key, if any.
	Deployment types.Option[model.DeploymentKey]
}

// ClaimRunnerForDeployment reserves a runner for the given deployment.
//
// Once a runner is reserved, it will be unavailable for other reservations
// or deployments and will not be returned by GetIdleRunnersForLanguage.
func (d *Postgres) ClaimRunnerForDeployment(ctx context.Context, language string, deployment model.DeploymentKey) (Runner, error) {
	runner, err := d.db.ClaimRunner(ctx, language, sqltypes.Key(deployment))
	if err != nil {
		if isNotFound(err) {
			counts, err := d.db.GetIdleRunnerCountsByLanguage(ctx)
			if err != nil {
				return Runner{}, errors.WithStack(translatePGError(err))
			}
			msg := fmt.Sprintf("no idle runners for language %q", language)
			if len(counts) > 0 {
				msg += ": available: " + strings.Join(slices.Map(counts, func(in sql.GetIdleRunnerCountsByLanguageRow) string {
					return fmt.Sprintf("%s:%d", in.Language, in.Count)
				}), ", ")
			}
			return Runner{}, errors.Wrap(ErrNotFound, msg)
		}
		return Runner{}, errors.WithStack(translatePGError(err))
	}
	return Runner{
		Key:      model.RunnerKey(runner.Key),
		Language: runner.Language,
		Endpoint: runner.Endpoint,
		State:    RunnerState(runner.State),
	}, nil
}

// SetDeploymentReplicas activates the given deployment.
func (d *Postgres) SetDeploymentReplicas(ctx context.Context, key model.DeploymentKey, minReplicas int) error {
	err := d.db.SetDeploymentDesiredReplicas(ctx, sqltypes.Key(key), int32(minReplicas))
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	return nil
}

type Reconciliation struct {
	Deployment model.DeploymentKey
	Module     string
	Language   string

	AssignedReplicas int
	RequiredReplicas int
}

// GetDeploymentsNeedingReconciliation returns deployments that have a
// mismatch between the number of assigned and required replicas.
func (d *Postgres) GetDeploymentsNeedingReconciliation(ctx context.Context) ([]Reconciliation, error) {
	counts, err := d.db.GetDeploymentsNeedingReconciliation(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, errors.WithStack(translatePGError(err))
	}
	return slices.Map(counts, func(t sql.GetDeploymentsNeedingReconciliationRow) Reconciliation {
		return Reconciliation{
			Deployment:       model.DeploymentKey(t.Key),
			Module:           t.ModuleName,
			Language:         t.Language,
			AssignedReplicas: int(t.AssignedRunnersCount),
			RequiredReplicas: int(t.RequiredRunnersCount),
		}
	}), nil
}

// GetIdleRunnersForLanguage returns up to limit idle runners for the given language.
//
// If no runners are available, it will return an empty slice.
func (d *Postgres) GetIdleRunnersForLanguage(ctx context.Context, language string, limit int) ([]Runner, error) {
	runners, err := d.db.GetIdleRunnersForLanguage(ctx, language, int32(limit))
	if isNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	return slices.Map(runners, func(row sql.Runner) Runner {
		return Runner{
			Key:      model.RunnerKey(row.Key),
			Language: row.Language,
			Endpoint: row.Endpoint,
			State:    RunnerState(row.State),
		}
	}), nil
}

// GetRunnersForModule returns all runners for the given module.
//
// If no runners are available, it will return an empty slice.
func (d *Postgres) GetRunnersForModule(ctx context.Context, module string) ([]Runner, error) {
	runners, err := d.db.GetRunnersForModule(ctx, module)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	if len(runners) == 0 {
		return nil, errors.WithStack(ErrNotFound)
	}
	return slices.Map(runners, func(row sql.GetRunnersForModuleRow) Runner {
		return Runner{
			Key:        model.RunnerKey(row.Key),
			Language:   row.Language,
			Endpoint:   row.Endpoint,
			State:      RunnerState(row.State),
			Deployment: types.Some(model.DeploymentKey(row.DeploymentKey)),
		}
	}), nil
}

// GetRoutingTable returns the endpoints for all runners for the given module.
func (d *Postgres) GetRoutingTable(ctx context.Context, module string) ([]string, error) {
	routes, err := d.db.GetRoutingTable(ctx, module)
	if len(routes) == 0 {
		return nil, errors.WithStack(ErrNotFound)
	}
	return routes, errors.WithStack(translatePGError(err))
}

type RunnerState string

// Runner states.
const (
	RunnerStateIdle     = RunnerState(sql.RunnerStateIdle)
	RunnerStateClaimed  = RunnerState(sql.RunnerStateClaimed)
	RunnerStateReserved = RunnerState(sql.RunnerStateReserved)
	RunnerStateAssigned = RunnerState(sql.RunnerStateAssigned)
)

func RunnerStateFromProto(state ftlv1.RunnerState) RunnerState {
	return RunnerState(strings.ToLower(state.String()))
}

func (d *Postgres) GetRunnerState(ctx context.Context, runnerKey model.RunnerKey) (RunnerState, error) {
	state, err := d.db.GetRunnerState(ctx, sqltypes.Key(runnerKey))
	if err != nil {
		return "", errors.WithStack(translatePGError(err))
	}
	return RunnerState(state), nil
}

// ExpireRunnerReservations and return the count.
func (d *Postgres) ExpireRunnerReservations(ctx context.Context) (int64, error) {
	count, err := d.db.ExpireRunnerReservations(ctx)
	return count, errors.WithStack(translatePGError(err))
}

func (d *Postgres) InsertDeploymentLogEntry(ctx context.Context, deployment model.DeploymentKey, logEntry log.Entry) error {
	logError := pgtype.Text{}
	if logEntry.Error != nil {
		logError.String = logEntry.Error.Error()
		logError.Valid = true
	}
	return errors.WithStack(translatePGError(d.db.InsertDeploymentLogEntry(ctx, sql.InsertDeploymentLogEntryParams{
		Key:       sqltypes.Key(deployment),
		TimeStamp: pgtype.Timestamptz{Time: logEntry.Time, Valid: true},
		Level:     int32(logEntry.Level.Severity()),
		Scope:     strings.Join(logEntry.Scope, ":"),
		Message:   logEntry.Message,
		Error:     logError,
	})))
}

func (d *Postgres) loadDeployment(ctx context.Context, deployment sql.GetDeploymentRow) (*model.Deployment, error) {
	pm := &pschema.Module{}
	err := proto.Unmarshal(deployment.Schema, pm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	module, err := schema.ModuleFromProto(pm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	out := &model.Deployment{
		Module:   deployment.ModuleName,
		Language: deployment.Language,
		Key:      model.DeploymentKey(deployment.Key),
		Schema:   module,
	}
	artefacts, err := d.db.GetDeploymentArtefacts(ctx, deployment.ID)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	out.Artefacts = slices.Map(artefacts, func(row sql.GetDeploymentArtefactsRow) *model.Artefact {
		return &model.Artefact{
			Path:       row.Path,
			Executable: row.Executable,
			Content:    &artefactReader{id: row.ID, db: d.db},
			Digest:     sha256.FromBytes(row.Digest),
		}
	})
	return out, nil
}

type DataPoint interface {
	isDataPoint()
}

type MetricHistogram struct {
	Count  int64
	Sum    int64
	Bucket []int64
}

func (MetricHistogram) isDataPoint() {}

type MetricCounter struct {
	Value int64
}

func (MetricCounter) isDataPoint() {}

type Metric struct {
	RunnerKey    model.RunnerKey
	StartTime    time.Time
	EndTime      time.Time
	SourceModule string
	SourceVerb   string
	DestModule   string
	DestVerb     string
	Name         string
	DataPoint    DataPoint
}

func (d *Postgres) InsertMetricEntry(ctx context.Context, metric Metric) error {
	var metricType sql.MetricType
	switch metric.DataPoint.(type) {
	case MetricHistogram:
		metricType = sql.MetricTypeHistogram
	case MetricCounter:
		metricType = sql.MetricTypeCounter
	default:
		panic(fmt.Sprintf("unknown metric type %T", metric.DataPoint))
	}

	datapointJSON, err := json.Marshal(metric.DataPoint)
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(translatePGError(d.db.InsertMetricEntry(ctx, sql.InsertMetricEntryParams{
		RunnerKey:    sqltypes.Key(metric.RunnerKey),
		StartTime:    pgtype.Timestamptz{Time: metric.StartTime, Valid: true},
		EndTime:      pgtype.Timestamptz{Time: metric.EndTime, Valid: true},
		SourceModule: metric.SourceModule,
		SourceVerb:   metric.SourceVerb,
		DestModule:   metric.DestModule,
		DestVerb:     metric.DestVerb,
		Name:         metric.Name,
		Type:         metricType,
		Value:        datapointJSON,
	})))
}

func sha256esToBytes(digests []sha256.SHA256) [][]byte {
	return slices.Map(digests, func(digest sha256.SHA256) []byte { return digest[:] })
}

type artefactReader struct {
	id     int64
	db     *sql.DB
	offset int32
}

func (r *artefactReader) Close() error { return nil }

func (r *artefactReader) Read(p []byte) (n int, err error) {
	content, err := r.db.GetArtefactContentRange(context.Background(), r.offset+1, int32(len(p)), r.id)
	if err != nil {
		return 0, errors.WithStack(translatePGError(err))
	}
	copy(p, content)
	clen := len(content)
	r.offset += int32(clen)
	if clen == 0 {
		err = io.EOF
	}
	return clen, err
}

func isNotFound(err error) bool {
	return errors.Is(err, stdsql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows)
}

func translatePGError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.ForeignKeyViolation {
			return errors.Wrap(ErrNotFound, strings.TrimSuffix(strings.TrimPrefix(pgErr.ConstraintName, pgErr.TableName+"_"), "_id_fkey"))
		} else if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrConflict
		}
	} else if isNotFound(err) {
		return ErrNotFound
	}
	return err
}
