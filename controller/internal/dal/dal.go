// Package dal provides a data abstraction layer for the Controller
package dal

import (
	"context"
	stdsql "database/sql"
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
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/controller/internal/sql"
	"github.com/TBD54566975/ftl/controller/internal/sqltypes"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/slices"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
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

type IngressRoute struct {
	Runner   model.RunnerKey
	Endpoint string
	Module   string
	Verb     string
}

type IngressRouteEntry struct {
	Deployment model.DeploymentKey
	Module     string
	Verb       string
	Method     string
	Path       string
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

type Runner struct {
	Key      model.RunnerKey
	Language string
	Endpoint string
	State    RunnerState
	// Assigned deployment key, if any.
	Deployment types.Option[model.DeploymentKey]
}

type Reconciliation struct {
	Deployment model.DeploymentKey
	Module     string
	Language   string

	AssignedReplicas int
	RequiredReplicas int
}

type RunnerState string

// Runner states.
const (
	RunnerStateIdle     = RunnerState(sql.RunnerStateIdle)
	RunnerStateReserved = RunnerState(sql.RunnerStateReserved)
	RunnerStateAssigned = RunnerState(sql.RunnerStateAssigned)
	RunnerStateDead     = RunnerState(sql.RunnerStateDead)
)

func RunnerStateFromProto(state ftlv1.RunnerState) RunnerState {
	return RunnerState(strings.ToLower(strings.TrimPrefix(state.String(), "RUNNER_")))
}

func (s RunnerState) ToProto() ftlv1.RunnerState {
	return ftlv1.RunnerState(ftlv1.RunnerState_value["RUNNER_"+strings.ToUpper(string(s))])
}

type ControllerState string

// Controller states.
const (
	ControllerStateLive = ControllerState(sql.ControllerStateLive)
	ControllerStateDead = ControllerState(sql.ControllerStateDead)
)

func ControllerStateFromProto(state ftlv1.ControllerState) ControllerState {
	return ControllerState(strings.ToLower(strings.TrimPrefix(state.String(), "CONTROLLER_")))
}

func (s ControllerState) ToProto() ftlv1.ControllerState {
	return ftlv1.ControllerState(ftlv1.ControllerState_value["CONTROLLER_"+strings.ToUpper(string(s))])
}

type CallEntry struct {
	ID            int64
	RequestKey    model.IngressRequestKey
	RunnerKey     model.RunnerKey
	ControllerKey model.ControllerKey
	Time          time.Time
	SourceVerb    schema.VerbRef
	DestVerb      schema.VerbRef
	Duration      time.Duration
	Request       []byte
	Response      []byte
	Error         error
}

type ModuleCallKey struct {
	Module string
	Verb   string
}

type Deployment struct {
	Key         model.DeploymentKey
	Language    string
	Module      string
	MinReplicas int
	Schema      *schema.Module
}

type Controller struct {
	Key      model.ControllerKey
	Endpoint string
	State    ControllerState
}

type Status struct {
	Controllers   []Controller
	Runners       []Runner
	Deployments   []Deployment
	IngressRoutes []IngressRouteEntry
}

// A Reservation of a Runner.
type Reservation interface {
	Runner() Runner
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Route struct {
	Runner   model.RunnerKey
	Endpoint string
}

func WithReservation(ctx context.Context, reservation Reservation, fn func() error) error {
	if err := fn(); err != nil {
		if rerr := reservation.Rollback(ctx); rerr != nil {
			err = errors.Join(err, rerr)
		}
		return errors.WithStack(err)
	}
	return errors.WithStack(reservation.Commit(ctx))
}

func New(conn sql.DBI) *DAL {
	return &DAL{db: sql.NewDB(conn)}
}

type DAL struct {
	db *sql.DB
}

func (d *DAL) GetStatus(
	ctx context.Context,
	allControllers, allRunners, allDeployments, allIngressRoutes bool,
) (Status, error) {
	controllers, err := d.db.GetControllers(ctx, allControllers)
	if err != nil {
		return Status{}, errors.Wrap(translatePGError(err), "could not get control planes")
	}
	runners, err := d.db.GetActiveRunners(ctx, allRunners)
	if err != nil {
		return Status{}, errors.Wrap(translatePGError(err), "could not get active runners")
	}
	deployments, err := d.db.GetDeployments(ctx, allDeployments)
	if err != nil {
		return Status{}, errors.Wrap(translatePGError(err), "could not get active deployments")
	}
	ingressRoutes, err := d.db.GetAllIngressRoutes(ctx, allIngressRoutes)
	if err != nil {
		return Status{}, errors.Wrap(translatePGError(err), "could not get ingress routes")
	}
	statusDeployments, err := slices.MapErr(deployments, func(in sql.GetDeploymentsRow) (Deployment, error) {
		protoSchema := &pschema.Module{}
		if err := proto.Unmarshal(in.Schema, protoSchema); err != nil {
			return Deployment{}, errors.Wrapf(err, "%q: could not unmarshal schema", in.ModuleName)
		}
		modelSchema, err := schema.ModuleFromProto(protoSchema)
		if err != nil {
			return Deployment{}, errors.Wrapf(err, "%q: invalid schema in database", in.ModuleName)
		}
		return Deployment{
			Key:         model.DeploymentKey(in.Key),
			Module:      in.ModuleName,
			Language:    in.Language,
			MinReplicas: int(in.MinReplicas),
			Schema:      modelSchema,
		}, nil
	})
	if err != nil {
		return Status{}, errors.WithStack(err)
	}
	return Status{
		Controllers: slices.Map(controllers, func(in sql.Controller) Controller {
			return Controller{
				Key:      model.ControllerKey(in.Key),
				Endpoint: in.Endpoint,
				State:    ControllerState(in.State),
			}
		}),
		Deployments: statusDeployments,
		Runners: slices.Map(runners, func(in sql.GetActiveRunnersRow) Runner {
			var deployment types.Option[model.DeploymentKey]
			// Need some hackery here because sqlc doesn't correctly handle the null column in this query.
			if in.DeploymentKey != nil {
				deployment = types.Some[model.DeploymentKey](in.DeploymentKey.([16]byte)) //nolint:forcetypeassert
			}
			return Runner{
				Key:        model.RunnerKey(in.RunnerKey),
				Language:   in.Language,
				Endpoint:   in.Endpoint,
				State:      RunnerState(in.State),
				Deployment: deployment,
			}
		}),
		IngressRoutes: slices.Map(ingressRoutes, func(in sql.GetAllIngressRoutesRow) IngressRouteEntry {
			return IngressRouteEntry{
				Deployment: model.DeploymentKey(in.DeploymentKey),
				Module:     in.Module,
				Verb:       in.Verb,
				Method:     in.Method,
				Path:       in.Path,
			}
		}),
	}, nil
}

func (d *DAL) GetRunnersForDeployment(ctx context.Context, deployment model.DeploymentKey) ([]Runner, error) {
	runners := []Runner{}
	rows, err := d.db.GetRunnersForDeployment(ctx, sqltypes.Key(deployment))
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	for _, row := range rows {
		runners = append(runners, Runner{
			Key:        model.RunnerKey(row.Key),
			Language:   row.Language,
			Endpoint:   row.Endpoint,
			State:      RunnerState(row.State),
			Deployment: types.Some(deployment),
		})
	}
	return runners, nil
}

func (d *DAL) UpsertModule(ctx context.Context, language, name string) (err error) {
	_, err = d.db.UpsertModule(ctx, language, name)
	return errors.WithStack(translatePGError(err))
}

// GetMissingArtefacts returns the digests of the artefacts that are missing from the database.
func (d *DAL) GetMissingArtefacts(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error) {
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
func (d *DAL) CreateArtefact(ctx context.Context, content []byte) (digest sha256.SHA256, err error) {
	sha256digest := sha256.Sum(content)
	_, err = d.db.CreateArtefact(ctx, sha256digest[:], content)
	return sha256digest, errors.WithStack(translatePGError(err))
}

type IngressRoutingEntry struct {
	Verb   string
	Method string
	Path   string
}

// CreateDeployment (possibly) creates a new deployment and associates
// previously created artefacts with it.
//
// If an existing deployment with identical artefacts exists, it is returned.
func (d *DAL) CreateDeployment(
	ctx context.Context,
	language string,
	schema *schema.Module,
	artefacts []DeploymentArtefact,
	ingressRoutes []IngressRoutingEntry,
) (key model.DeploymentKey, err error) {
	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return model.DeploymentKey{}, errors.WithStack(err)
	}
	defer tx.CommitOrRollback(ctx, &err)

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

	deploymentKey := model.NewDeploymentKey()
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

	for _, ingressRoute := range ingressRoutes {
		err = tx.CreateIngressRoute(ctx, sql.CreateIngressRouteParams{
			Key:    sqltypes.Key(deploymentKey),
			Method: ingressRoute.Method,
			Path:   ingressRoute.Path,
			Module: schema.Name,
			Verb:   ingressRoute.Verb,
		})
		if err != nil {
			return model.DeploymentKey{}, errors.WithStack(translatePGError(err))
		}
	}

	return deploymentKey, nil
}

func (d *DAL) GetDeployment(ctx context.Context, id model.DeploymentKey) (*model.Deployment, error) {
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
func (d *DAL) UpsertRunner(ctx context.Context, runner Runner) error {
	var pgDeploymentKey types.Option[sqltypes.Key]
	if dkey, ok := runner.Deployment.Get(); ok {
		pgDeploymentKey = types.Some(sqltypes.Key(dkey))
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

// KillStaleRunners deletes runners that have not had heartbeats for the given duration.
func (d *DAL) KillStaleRunners(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.KillStaleRunners(ctx, pgtype.Interval{
		Microseconds: int64(age / time.Microsecond),
		Valid:        true,
	})
	return count, errors.WithStack(err)
}

// KillStaleControllers deletes controllers that have not had heartbeats for the given duration.
func (d *DAL) KillStaleControllers(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.KillStaleControllers(ctx, pgtype.Interval{
		Microseconds: int64(age / time.Microsecond),
		Valid:        true,
	})
	return count, errors.WithStack(err)
}

// DeregisterRunner deregisters the given runner.
func (d *DAL) DeregisterRunner(ctx context.Context, key model.RunnerKey) error {
	count, err := d.db.DeregisterRunner(ctx, sqltypes.Key(key))
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *DAL) ReserveRunnerForDeployment(ctx context.Context, language string, deployment model.DeploymentKey, reservationTimeout time.Duration) (Reservation, error) {
	ctx, cancel := context.WithTimeout(ctx, reservationTimeout)
	tx, err := d.db.Begin(ctx)
	if err != nil {
		cancel()
		return nil, errors.WithStack(translatePGError(err))
	}
	runner, err := tx.ReserveRunner(ctx, language, pgtype.Timestamptz{Time: time.Now().Add(reservationTimeout), Valid: true}, sqltypes.Key(deployment))
	if err != nil {
		if rerr := tx.Rollback(context.Background()); rerr != nil {
			err = errors.Join(err, errors.WithStack(translatePGError(rerr)))
		}
		cancel()
		if isNotFound(err) {
			return nil, errors.Wrapf(ErrNotFound, "no idle runners for language %q", language)
		}
		return nil, errors.WithStack(translatePGError(err))
	}
	return &postgresClaim{
		cancel: cancel,
		tx:     tx,
		runner: Runner{
			Key:        model.RunnerKey(runner.Key),
			Language:   runner.Language,
			Endpoint:   runner.Endpoint,
			State:      RunnerState(runner.State),
			Deployment: types.Some(deployment),
		},
	}, nil
}

var _ Reservation = (*postgresClaim)(nil)

type postgresClaim struct {
	tx     *sql.Tx
	runner Runner
	cancel context.CancelFunc
}

func (p *postgresClaim) Commit(ctx context.Context) error {
	defer p.cancel()
	return errors.WithStack(translatePGError(p.tx.Commit(ctx)))
}

func (p *postgresClaim) Rollback(ctx context.Context) error {
	defer p.cancel()
	return errors.WithStack(translatePGError(p.tx.Rollback(ctx)))
}

func (p *postgresClaim) Runner() Runner { return p.runner }

// SetDeploymentReplicas activates the given deployment.
func (d *DAL) SetDeploymentReplicas(ctx context.Context, key model.DeploymentKey, minReplicas int) error {
	err := d.db.SetDeploymentDesiredReplicas(ctx, sqltypes.Key(key), int32(minReplicas))
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	return nil
}

// ReplaceDeployment replaces an old deployment of a module with a new deployment.
func (d *DAL) ReplaceDeployment(ctx context.Context, newDeploymentKey model.DeploymentKey, minReplicas int) (err error) {
	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	defer tx.CommitOrRollback(ctx, &err)
	newDeployment, err := tx.GetDeployment(ctx, sqltypes.Key(newDeploymentKey))
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	// If there's an existing deployment, set its desired replicas to 0
	oldDeployment, err := tx.GetExistingDeploymentForModule(ctx, newDeployment.ModuleName)
	if err == nil {
		count, err := tx.ReplaceDeployment(ctx, oldDeployment.Key, sqltypes.Key(newDeploymentKey), int32(minReplicas))
		if err != nil {
			return errors.WithStack(translatePGError(err))
		}
		if count == 1 {
			return errors.Wrap(ErrConflict, "deployment already exists")
		}
	} else if !isNotFound(err) {
		return errors.WithStack(translatePGError(err))
	} else {
		// Set the desired replicas for the new deployment
		err = tx.SetDeploymentDesiredReplicas(ctx, sqltypes.Key(newDeploymentKey), int32(minReplicas))
		if err != nil {
			return errors.WithStack(translatePGError(err))
		}
	}
	return nil
}

// GetDeploymentsNeedingReconciliation returns deployments that have a
// mismatch between the number of assigned and required replicas.
func (d *DAL) GetDeploymentsNeedingReconciliation(ctx context.Context) ([]Reconciliation, error) {
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

// GetActiveDeployments returns all active deployments.
func (d *DAL) GetActiveDeployments(ctx context.Context) ([]Deployment, error) {
	rows, err := d.db.GetDeployments(ctx, false)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, errors.WithStack(translatePGError(err))
	}
	deployments, err := slices.MapErr(rows, func(in sql.GetDeploymentsRow) (Deployment, error) {
		protoSchema := &pschema.Module{}
		if err := proto.Unmarshal(in.Schema, protoSchema); err != nil {
			return Deployment{}, errors.Wrapf(err, "%q: could not unmarshal schema", in.ModuleName)
		}
		modelSchema, err := schema.ModuleFromProto(protoSchema)
		if err != nil {
			return Deployment{}, errors.Wrapf(err, "%q: invalid schema in database", in.ModuleName)
		}
		return Deployment{
			Key:         model.DeploymentKey(in.Key),
			Module:      in.ModuleName,
			Language:    in.Language,
			MinReplicas: int(in.MinReplicas),
			Schema:      modelSchema,
		}, nil
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return deployments, nil
}

// GetIdleRunnersForLanguage returns up to limit idle runners for the given language.
//
// If no runners are available, it will return an empty slice.
func (d *DAL) GetIdleRunnersForLanguage(ctx context.Context, language string, limit int) ([]Runner, error) {
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

// GetRoutingTable returns the endpoints for all runners for the given module.
func (d *DAL) GetRoutingTable(ctx context.Context, module string) ([]Route, error) {
	routes, err := d.db.GetRoutingTable(ctx, module)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	if len(routes) == 0 {
		return nil, errors.WithStack(ErrNotFound)
	}
	return slices.Map(routes, func(row sql.GetRoutingTableRow) Route {
		return Route{
			Runner:   model.RunnerKey(row.Key),
			Endpoint: row.Endpoint,
		}
	}), nil
}

func (d *DAL) GetRunnerState(ctx context.Context, runnerKey model.RunnerKey) (RunnerState, error) {
	state, err := d.db.GetRunnerState(ctx, sqltypes.Key(runnerKey))
	if err != nil {
		return "", errors.WithStack(translatePGError(err))
	}
	return RunnerState(state), nil
}

func (d *DAL) ExpireRunnerClaims(ctx context.Context) (int64, error) {
	count, err := d.db.ExpireRunnerReservations(ctx)
	return count, errors.WithStack(translatePGError(err))
}

func (d *DAL) InsertDeploymentLogEntry(ctx context.Context, deployment model.DeploymentKey, logEntry log.Entry) error {
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

func (d *DAL) loadDeployment(ctx context.Context, deployment sql.GetDeploymentRow) (*model.Deployment, error) {
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

func (d *DAL) CreateIngressRequest(ctx context.Context, addr string) (model.IngressRequestKey, error) {
	key := model.NewIngressRequestKey()
	err := d.db.CreateIngressRequest(ctx, sqltypes.Key(key), addr)
	return key, errors.WithStack(err)
}

func (d *DAL) GetIngressRoutes(ctx context.Context, method string, path string) ([]IngressRoute, error) {
	routes, err := d.db.GetIngressRoutes(ctx, method, path)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	if len(routes) == 0 {
		return nil, errors.WithStack(ErrNotFound)
	}
	return slices.Map(routes, func(row sql.GetIngressRoutesRow) IngressRoute {
		return IngressRoute{
			Runner:   model.RunnerKey(row.RunnerKey),
			Endpoint: row.Endpoint,
			Module:   row.Module,
			Verb:     row.Verb,
		}
	}), nil
}

func (d *DAL) UpsertController(ctx context.Context, key model.ControllerKey, addr string) (int64, error) {
	id, err := d.db.UpsertController(ctx, sqltypes.Key(key), addr)
	return id, errors.WithStack(translatePGError(err))
}

func (d *DAL) InsertCallEntry(ctx context.Context, call *CallEntry) error {
	callError := pgtype.Text{}
	if call.Error != nil {
		callError.String = call.Error.Error()
		callError.Valid = true
	}
	return errors.WithStack(translatePGError(d.db.InsertCallEntry(ctx, sql.InsertCallEntryParams{
		Key:          sqltypes.Key(call.RunnerKey),
		Key_2:        sqltypes.Key(call.RequestKey),
		Key_3:        sqltypes.Key(call.ControllerKey),
		SourceModule: call.SourceVerb.Module,
		SourceVerb:   call.SourceVerb.Name,
		DestModule:   call.DestVerb.Module,
		DestVerb:     call.DestVerb.Name,
		DurationMs:   call.Duration.Milliseconds(),
		Request:      call.Request,
		Response:     call.Response,
		Error:        callError,
	})))
}

func (d *DAL) GetModuleCalls(ctx context.Context, modules []string) (map[ModuleCallKey][]CallEntry, error) {
	calls, err := d.db.GetModuleCalls(ctx, modules)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	out := map[ModuleCallKey][]CallEntry{}
	for _, call := range calls {
		key := ModuleCallKey{
			Module: call.DestModule,
			Verb:   call.DestVerb,
		}
		var callError error
		if call.Error.Valid {
			callError = errors.New(call.Error.String)
		}
		out[key] = append(out[key], CallEntry{
			ID:            call.ID,
			RequestKey:    model.IngressRequestKey(call.IngressRequestKey),
			RunnerKey:     model.RunnerKey(call.RunnerKey),
			ControllerKey: model.ControllerKey(call.ControllerKey),
			Time:          call.Time.Time,
			SourceVerb: schema.VerbRef{
				Module: call.SourceModule,
				Name:   call.SourceVerb,
			},
			DestVerb: schema.VerbRef{
				Module: call.DestModule,
				Name:   call.DestVerb,
			},
			Duration: time.Duration(call.DurationMs) * time.Millisecond,
			Request:  call.Request,
			Response: call.Response,
			Error:    callError,
		})
	}
	return out, nil
}

func (d *DAL) GetRequestCalls(ctx context.Context, requestKey model.IngressRequestKey) ([]CallEntry, error) {
	calls, err := d.db.GetRequestCalls(ctx, sqltypes.Key(requestKey))
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	var out []CallEntry
	for _, call := range calls {
		var callError error
		if call.Error.Valid {
			callError = errors.New(call.Error.String)
		}
		out = append(out, CallEntry{
			ID:            call.ID,
			RequestKey:    requestKey,
			RunnerKey:     model.RunnerKey(call.RunnerKey),
			ControllerKey: model.ControllerKey(call.ControllerKey),
			Time:          call.Time.Time,
			SourceVerb: schema.VerbRef{
				Module: call.SourceModule,
				Name:   call.SourceVerb,
			},
			DestVerb: schema.VerbRef{
				Module: call.DestModule,
				Name:   call.DestVerb,
			},
			Duration: time.Duration(call.DurationMs) * time.Millisecond,
			Request:  call.Request,
			Response: call.Response,
			Error:    callError,
		})
	}
	return out, nil
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
	if err == nil {
		return nil
	}
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
