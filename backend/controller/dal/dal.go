// Package dal provides a data abstraction layer for the Controller
package dal

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/maps"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/common/pubsub"
	"github.com/TBD54566975/ftl/backend/common/sha256"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
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
	Runner     model.RunnerKey
	Deployment model.DeploymentName
	Endpoint   string
	Path       string
	Module     string
	Verb       string
}

type IngressRouteEntry struct {
	Deployment model.DeploymentName
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
		return DeploymentArtefact{}, err
	}
	return DeploymentArtefact{
		Digest:     digest,
		Executable: in.Executable,
		Path:       in.Path,
	}, nil
}

func runnerFromDB(row sql.GetRunnerRow) Runner {
	var deployment optional.Option[model.DeploymentName]
	if name, ok := row.DeploymentName.Get(); ok {
		deployment = optional.Some(model.DeploymentName(name))
	}
	attrs := model.Labels{}
	if err := json.Unmarshal(row.Labels, &attrs); err != nil {
		return Runner{}
	}
	return Runner{
		Key:        model.RunnerKey(row.RunnerKey),
		Endpoint:   row.Endpoint,
		State:      RunnerState(row.State),
		Deployment: deployment,
		Labels:     attrs,
	}
}

type Runner struct {
	Key                model.RunnerKey
	Endpoint           string
	State              RunnerState
	ReservationTimeout optional.Option[time.Duration]
	Module             optional.Option[string]
	// Assigned deployment key, if any.
	Deployment optional.Option[model.DeploymentName]
	Labels     model.Labels
}

func (r Runner) notification() {}

type Reconciliation struct {
	Deployment model.DeploymentName
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

type RequestOrigin string

const (
	RequestOriginIngress = RequestOrigin(sql.OriginIngress)
	RequestOriginCron    = RequestOrigin(sql.OriginCron)
	RequestOriginPubsub  = RequestOrigin(sql.OriginPubsub)
)

type Deployment struct {
	Name        model.DeploymentName
	Language    string
	Module      string
	MinReplicas int
	Schema      *schema.Module
	CreatedAt   time.Time
	Labels      model.Labels
}

func (d Deployment) String() string { return d.Name.String() }

func (d Deployment) notification() {}

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
	Routes        []Route
}

// A Reservation of a Runner.
type Reservation interface {
	Runner() Runner
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Route struct {
	Module     string
	Runner     model.RunnerKey
	Deployment model.DeploymentName
	Endpoint   string
}

func (r Route) String() string {
	return fmt.Sprintf("%s -> %s (%s)", r.Deployment, r.Runner, r.Endpoint)
}

func (r Route) notification() {}

func WithReservation(ctx context.Context, reservation Reservation, fn func() error) error {
	if err := fn(); err != nil {
		if rerr := reservation.Rollback(ctx); rerr != nil {
			err = errors.Join(err, rerr)
		}
		return err
	}
	return reservation.Commit(ctx)
}

func New(ctx context.Context, pool *pgxpool.Pool) (*DAL, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "failed to acquire PG PubSub connection", err)
	}
	dal := &DAL{
		db:                sql.NewDB(pool),
		DeploymentChanges: pubsub.New[DeploymentNotification](),
	}
	go dal.runListener(ctx, conn.Hijack())
	return dal, nil
}

type DAL struct {
	db *sql.DB

	// DeploymentChanges is a Topic that receives changes to the deployments table.
	DeploymentChanges *pubsub.Topic[DeploymentNotification]
	// RouteChanges is a Topic that receives changes to the routing table.
}

func (d *DAL) GetControllers(ctx context.Context, allControllers bool) ([]Controller, error) {
	controllers, err := d.db.GetControllers(ctx, allControllers)
	if err != nil {
		return nil, translatePGError(err)
	}
	return slices.Map(controllers, func(in sql.Controller) Controller {
		return Controller{
			Key:      in.Key,
			Endpoint: in.Endpoint,
			State:    ControllerState(in.State),
		}
	}), nil
}

func (d *DAL) GetStatus(
	ctx context.Context,
	allControllers, allRunners, allDeployments, allIngressRoutes bool,
) (Status, error) {
	controllers, err := d.GetControllers(ctx, allControllers)
	if err != nil {
		return Status{}, fmt.Errorf("%s: %w", "could not get control planes", translatePGError(err))
	}
	runners, err := d.db.GetActiveRunners(ctx, allRunners)
	if err != nil {
		return Status{}, fmt.Errorf("%s: %w", "could not get active runners", translatePGError(err))
	}
	deployments, err := d.db.GetActiveDeployments(ctx, allDeployments)
	if err != nil {
		return Status{}, fmt.Errorf("%s: %w", "could not get active deployments", translatePGError(err))
	}
	ingressRoutes, err := d.db.GetAllIngressRoutes(ctx, allIngressRoutes)
	if err != nil {
		return Status{}, fmt.Errorf("%s: %w", "could not get ingress routes", translatePGError(err))
	}
	routes, err := d.db.GetRoutingTable(ctx, nil)
	if err != nil {
		return Status{}, fmt.Errorf("%s: %w", "could not get routing table", translatePGError(err))
	}
	statusDeployments, err := slices.MapErr(deployments, func(in sql.GetActiveDeploymentsRow) (Deployment, error) {
		labels := model.Labels{}
		err = json.Unmarshal(in.Deployment.Labels, &labels)
		if err != nil {
			return Deployment{}, fmt.Errorf("%q: invalid labels in database: %w", in.ModuleName, err)
		}
		return Deployment{
			Name:        in.Deployment.Name,
			Module:      in.ModuleName,
			Language:    in.Language,
			MinReplicas: int(in.Deployment.MinReplicas),
			Schema:      in.Deployment.Schema,
			Labels:      labels,
		}, nil
	})
	if err != nil {
		return Status{}, err
	}
	domainRunners, err := slices.MapErr(runners, func(in sql.GetActiveRunnersRow) (Runner, error) {
		var deployment optional.Option[model.DeploymentName]
		if name, ok := in.DeploymentName.Get(); ok {
			deployment = optional.Some(model.DeploymentName(name))
		}
		attrs := model.Labels{}
		if err := json.Unmarshal(in.Labels, &attrs); err != nil {
			return Runner{}, fmt.Errorf("invalid attributes JSON for runner %s: %w", in.RunnerKey, err)
		}
		return Runner{
			Key:        model.RunnerKey(in.RunnerKey),
			Endpoint:   in.Endpoint,
			State:      RunnerState(in.State),
			Deployment: deployment,
			Labels:     attrs,
		}, nil
	})
	if err != nil {
		return Status{}, err
	}
	return Status{
		Controllers: controllers,
		Deployments: statusDeployments,
		Runners:     domainRunners,
		IngressRoutes: slices.Map(ingressRoutes, func(in sql.GetAllIngressRoutesRow) IngressRouteEntry {
			return IngressRouteEntry{
				Deployment: in.DeploymentName,
				Module:     in.Module,
				Verb:       in.Verb,
				Method:     in.Method,
				Path:       in.Path,
			}
		}),
		Routes: slices.Map(routes, func(row sql.GetRoutingTableRow) Route {
			return Route{
				Module:     row.ModuleName.MustGet(),
				Runner:     model.RunnerKey(row.RunnerKey),
				Deployment: row.DeploymentName,
				Endpoint:   row.Endpoint,
			}
		}),
	}, nil
}

func (d *DAL) GetRunnersForDeployment(ctx context.Context, deployment model.DeploymentName) ([]Runner, error) {
	runners := []Runner{}
	rows, err := d.db.GetRunnersForDeployment(ctx, deployment)
	if err != nil {
		return nil, translatePGError(err)
	}
	for _, row := range rows {
		attrs := model.Labels{}
		if err := json.Unmarshal(row.Labels, &attrs); err != nil {
			return nil, fmt.Errorf("invalid attributes JSON for runner %d: %w", row.ID, err)
		}
		runners = append(runners, Runner{
			Key:        model.RunnerKey(row.Key),
			Endpoint:   row.Endpoint,
			State:      RunnerState(row.State),
			Deployment: optional.Some(deployment),
			Labels:     attrs,
		})
	}
	return runners, nil
}

func (d *DAL) UpsertModule(ctx context.Context, language, name string) (err error) {
	_, err = d.db.UpsertModule(ctx, language, name)
	return translatePGError(err)
}

// GetMissingArtefacts returns the digests of the artefacts that are missing from the database.
func (d *DAL) GetMissingArtefacts(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error) {
	have, err := d.db.GetArtefactDigests(ctx, sha256esToBytes(digests))
	if err != nil {
		return nil, translatePGError(err)
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
	return sha256digest, translatePGError(err)
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
func (d *DAL) CreateDeployment(ctx context.Context, language string, moduleSchema *schema.Module, artefacts []DeploymentArtefact, ingressRoutes []IngressRoutingEntry) (key model.DeploymentName, err error) {
	logger := log.FromContext(ctx)

	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: %w", "could not start transaction", err)
	}

	defer tx.CommitOrRollback(ctx, &err)

	existingDeployment, err := d.checkForExistingDeployments(ctx, tx, moduleSchema, artefacts)
	if err != nil {
		return "", err
	} else if existingDeployment != "" {
		logger.Tracef("Returning existing deployment %s", existingDeployment)
		return existingDeployment, nil
	}

	artefactsByDigest := maps.FromSlice(artefacts, func(in DeploymentArtefact) (sha256.SHA256, DeploymentArtefact) {
		return in.Digest, in
	})

	schemaBytes, err := proto.Marshal(moduleSchema.ToProto())
	if err != nil {
		return "", fmt.Errorf("%s: %w", "failed to marshal schema", err)
	}

	// TODO(aat): "schema" containing language?
	_, err = tx.UpsertModule(ctx, language, moduleSchema.Name)
	if err != nil {
		return "", fmt.Errorf("%s: %w", "failed to upsert module", translatePGError(err))
	}

	deploymentName := model.NewDeploymentName(moduleSchema.Name)
	// Create the deployment
	err = tx.CreateDeployment(ctx, deploymentName, moduleSchema.Name, schemaBytes)
	if err != nil {
		return "", fmt.Errorf("%s: %w", "failed to create deployment", translatePGError(err))
	}

	uploadedDigests := slices.Map(artefacts, func(in DeploymentArtefact) []byte { return in.Digest[:] })
	artefactDigests, err := tx.GetArtefactDigests(ctx, uploadedDigests)
	if err != nil {
		return "", fmt.Errorf("%s: %w", "failed to get artefact digests", err)
	}
	if len(artefactDigests) != len(artefacts) {
		missingDigests := strings.Join(slices.Map(artefacts, func(in DeploymentArtefact) string { return in.Digest.String() }), ", ")
		return "", fmt.Errorf("missing %d artefacts: %s", len(artefacts)-len(artefactDigests), missingDigests)
	}

	// Associate the artefacts with the deployment
	for _, row := range artefactDigests {
		artefact := artefactsByDigest[sha256.FromBytes(row.Digest)]
		err = tx.AssociateArtefactWithDeployment(ctx, sql.AssociateArtefactWithDeploymentParams{
			Name:       deploymentName,
			ArtefactID: row.ID,
			Executable: artefact.Executable,
			Path:       artefact.Path,
		})
		if err != nil {
			return "", fmt.Errorf("%s: %w", "failed to associate artefact with deployment", translatePGError(err))
		}
	}

	for _, ingressRoute := range ingressRoutes {
		err = tx.CreateIngressRoute(ctx, sql.CreateIngressRouteParams{
			Name:   deploymentName,
			Method: ingressRoute.Method,
			Path:   ingressRoute.Path,
			Module: moduleSchema.Name,
			Verb:   ingressRoute.Verb,
		})
		if err != nil {
			return "", fmt.Errorf("%s: %w", "failed to create ingress route", translatePGError(err))
		}
	}

	return deploymentName, nil
}

func (d *DAL) GetDeployment(ctx context.Context, name model.DeploymentName) (*model.Deployment, error) {
	deployment, err := d.db.GetDeployment(ctx, name)
	if err != nil {
		return nil, translatePGError(err)
	}
	return d.loadDeployment(ctx, deployment)
}

// UpsertRunner registers or updates a new runner.
//
// ErrConflict will be returned if a runner with the same endpoint and a
// different key already exists.
func (d *DAL) UpsertRunner(ctx context.Context, runner Runner) error {
	var pgDeploymentName optional.Option[string]
	if dkey, ok := runner.Deployment.Get(); ok {
		pgDeploymentName = optional.Some(dkey.String())
	}
	attrBytes, err := json.Marshal(runner.Labels)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to JSON encode runner labels", err)
	}
	deploymentID, err := d.db.UpsertRunner(ctx, sql.UpsertRunnerParams{
		Key:            sql.Key(runner.Key),
		Endpoint:       runner.Endpoint,
		State:          sql.RunnerState(runner.State),
		DeploymentName: pgDeploymentName,
		Labels:         attrBytes,
	})
	if err != nil {
		return translatePGError(err)
	}
	if err != nil {
		return translatePGError(err)
	}
	if runner.Deployment.Ok() && !deploymentID.Ok() {
		return fmt.Errorf("deployment %s not found", runner.Deployment)
	}
	return nil
}

// KillStaleRunners deletes runners that have not had heartbeats for the given duration.
func (d *DAL) KillStaleRunners(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.KillStaleRunners(ctx, age)
	return count, err
}

// KillStaleControllers deletes controllers that have not had heartbeats for the given duration.
func (d *DAL) KillStaleControllers(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.KillStaleControllers(ctx, age)
	return count, err
}

// DeregisterRunner deregisters the given runner.
func (d *DAL) DeregisterRunner(ctx context.Context, key model.RunnerKey) error {
	count, err := d.db.DeregisterRunner(ctx, sql.Key(key))
	if err != nil {
		return translatePGError(err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// ReserveRunnerForDeployment reserves a runner for the given deployment.
//
// It returns a Reservation that must be committed or rolled back.
func (d *DAL) ReserveRunnerForDeployment(ctx context.Context, deployment model.DeploymentName, reservationTimeout time.Duration, labels model.Labels) (Reservation, error) {
	jsonLabels, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "failed to JSON encode labels", err)
	}
	ctx, cancel := context.WithTimeout(ctx, reservationTimeout)
	tx, err := d.db.Begin(ctx)
	if err != nil {
		cancel()
		return nil, translatePGError(err)
	}
	runner, err := tx.ReserveRunner(ctx, time.Now().Add(reservationTimeout), deployment, jsonLabels)
	if err != nil {
		if rerr := tx.Rollback(context.Background()); rerr != nil {
			err = errors.Join(err, translatePGError(rerr))
		}
		cancel()
		if isNotFound(err) {
			return nil, fmt.Errorf("no idle runners found matching labels %s: %w", jsonLabels, ErrNotFound)
		}
		return nil, translatePGError(err)
	}
	runnerLabels := model.Labels{}
	if err := json.Unmarshal(runner.Labels, &runnerLabels); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to JSON decode labels for runner %d: %w", runner.ID, err)
	}
	return &postgresClaim{
		cancel: cancel,
		tx:     tx,
		runner: Runner{
			Key:        model.RunnerKey(runner.Key),
			Endpoint:   runner.Endpoint,
			State:      RunnerState(runner.State),
			Deployment: optional.Some(deployment),
			Labels:     runnerLabels,
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
	return translatePGError(p.tx.Commit(ctx))
}

func (p *postgresClaim) Rollback(ctx context.Context) error {
	defer p.cancel()
	return translatePGError(p.tx.Rollback(ctx))
}

func (p *postgresClaim) Runner() Runner { return p.runner }

// SetDeploymentReplicas activates the given deployment.
func (d *DAL) SetDeploymentReplicas(ctx context.Context, key model.DeploymentName, minReplicas int) error {
	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return translatePGError(err)
	}

	defer tx.CommitOrRollback(ctx, &err)

	deployment, err := d.db.GetDeployment(ctx, key)
	if err != nil {
		return translatePGError(err)
	}

	err = d.db.SetDeploymentDesiredReplicas(ctx, key, int32(minReplicas))
	if err != nil {
		return translatePGError(err)
	}

	err = tx.InsertDeploymentUpdatedEvent(ctx, sql.InsertDeploymentUpdatedEventParams{
		DeploymentName:  key.String(),
		MinReplicas:     int32(minReplicas),
		PrevMinReplicas: deployment.MinReplicas,
	})
	if err != nil {
		return translatePGError(err)
	}

	return nil
}

// ReplaceDeployment replaces an old deployment of a module with a new deployment.
func (d *DAL) ReplaceDeployment(ctx context.Context, newDeploymentName model.DeploymentName, minReplicas int) (err error) {
	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return translatePGError(err)
	}

	defer tx.CommitOrRollback(ctx, &err)
	newDeployment, err := tx.GetDeployment(ctx, newDeploymentName)
	if err != nil {
		return translatePGError(err)
	}

	var replacedDeployment optional.Option[string]

	// If there's an existing deployment, set its desired replicas to 0
	oldDeployment, err := tx.GetExistingDeploymentForModule(ctx, newDeployment.ModuleName)
	if err == nil {
		count, err := tx.ReplaceDeployment(ctx, oldDeployment.Name.String(), newDeploymentName.String(), int32(minReplicas))
		if err != nil {
			return translatePGError(err)
		}
		if count == 1 {
			return fmt.Errorf("%s: %w", "deployment already exists", ErrConflict)
		}
		replacedDeployment = optional.Some(oldDeployment.Name.String())
	} else if !isNotFound(err) {
		return translatePGError(err)
	} else {
		// Set the desired replicas for the new deployment
		err = tx.SetDeploymentDesiredReplicas(ctx, newDeploymentName, int32(minReplicas))
		if err != nil {
			return translatePGError(err)
		}
	}

	err = tx.InsertDeploymentCreatedEvent(ctx, sql.InsertDeploymentCreatedEventParams{
		DeploymentName: newDeploymentName.String(),
		Language:       newDeployment.Language,
		ModuleName:     newDeployment.ModuleName,
		MinReplicas:    int32(minReplicas),
		Replaced:       replacedDeployment,
	})
	if err != nil {
		return translatePGError(err)
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
		return nil, translatePGError(err)
	}
	return slices.Map(counts, func(t sql.GetDeploymentsNeedingReconciliationRow) Reconciliation {
		return Reconciliation{
			Deployment:       t.DeploymentName,
			Module:           t.ModuleName,
			Language:         t.Language,
			AssignedReplicas: int(t.AssignedRunnersCount),
			RequiredReplicas: int(t.RequiredRunnersCount),
		}
	}), nil
}

// GetActiveDeployments returns all active deployments.
func (d *DAL) GetActiveDeployments(ctx context.Context) ([]Deployment, error) {
	rows, err := d.db.GetActiveDeployments(ctx, false)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, translatePGError(err)
	}
	return slices.MapErr(rows, func(in sql.GetActiveDeploymentsRow) (Deployment, error) {
		return Deployment{
			Name:        in.Deployment.Name,
			Module:      in.ModuleName,
			Language:    in.Language,
			MinReplicas: int(in.Deployment.MinReplicas),
			Schema:      in.Deployment.Schema,
			CreatedAt:   in.Deployment.CreatedAt,
		}, nil
	})
}

func (d *DAL) GetActiveDeploymentSchemas(ctx context.Context) ([]*schema.Module, error) {
	rows, err := d.db.GetActiveDeploymentSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "could not get active deployments", translatePGError(err))
	}
	return slices.MapErr(rows, func(in sql.GetActiveDeploymentSchemasRow) (*schema.Module, error) { return in.Schema, nil })
}

type ProcessRunner struct {
	Key      model.RunnerKey
	Endpoint string
	Labels   model.Labels
}

type Process struct {
	Deployment  model.DeploymentName
	MinReplicas int
	Labels      model.Labels
	Runner      optional.Option[ProcessRunner]
}

// GetProcessList returns a list of all "processes".
func (d *DAL) GetProcessList(ctx context.Context) ([]Process, error) {
	rows, err := d.db.GetProcessList(ctx)
	if err != nil {
		return nil, translatePGError(err)
	}
	return slices.MapErr(rows, func(row sql.GetProcessListRow) (Process, error) {
		var runner optional.Option[ProcessRunner]
		if endpoint, ok := row.Endpoint.Get(); ok {
			var labels model.Labels
			if err := json.Unmarshal(row.RunnerLabels, &labels); err != nil {
				return Process{}, fmt.Errorf("invalid labels JSON for runner %s: %w", row.RunnerKey, err)
			}
			runner = optional.Some(ProcessRunner{
				Key:      model.RunnerKey(row.RunnerKey.MustGet()),
				Endpoint: endpoint,
				Labels:   labels,
			})
		}
		var labels model.Labels
		if err := json.Unmarshal(row.DeploymentLabels, &labels); err != nil {
			return Process{}, fmt.Errorf("invalid labels JSON for deployment %s: %w", row.DeploymentName, err)
		}
		return Process{
			Deployment:  row.DeploymentName,
			Labels:      labels,
			MinReplicas: int(row.MinReplicas),
			Runner:      runner,
		}, nil
	})
}

// GetIdleRunners returns up to limit idle runners matching the given labels.
//
// "labels" is a single level map of key-value pairs. Values may be scalar or
// lists of scalars. If a value is a list, it will match the labels if
// all the values in the list match.
//
// e.g. {"languages": ["kotlin"], "arch": "arm64"}' will match a runner with the labels
// '{"languages": ["go", "kotlin"], "os": "linux", "arch": "amd64", "pid": 1234}'
//
// If no runners are available, it will return an empty slice.
func (d *DAL) GetIdleRunners(ctx context.Context, limit int, labels model.Labels) ([]Runner, error) {
	jsonb, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "could not marshal labels", err)
	}
	runners, err := d.db.GetIdleRunners(ctx, jsonb, int64(limit))
	if isNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, translatePGError(err)
	}
	return slices.MapErr(runners, func(row sql.Runner) (Runner, error) {
		rowLabels := model.Labels{}
		err := json.Unmarshal(row.Labels, &rowLabels)
		if err != nil {
			return Runner{}, fmt.Errorf("%s: %w", "could not unmarshal labels", err)
		}
		return Runner{
			Key:      model.RunnerKey(row.Key),
			Endpoint: row.Endpoint,
			State:    RunnerState(row.State),
			Labels:   labels,
		}, nil
	})
}

// GetRoutingTable returns the endpoints for all runners for the given modules,
// or all routes if modules is empty.
//
// Returns route map keyed by module.
func (d *DAL) GetRoutingTable(ctx context.Context, modules []string) (map[string][]Route, error) {
	routes, err := d.db.GetRoutingTable(ctx, modules)
	if err != nil {
		return nil, translatePGError(err)
	}
	if len(routes) == 0 {
		return nil, fmt.Errorf("%s: %w", "no routes found", ErrNotFound)
	}
	out := make(map[string][]Route, len(routes))
	for _, route := range routes {
		// This is guaranteed to be non-nil by the query, but sqlc doesn't quite understand that.
		moduleName := route.ModuleName.MustGet()
		out[moduleName] = append(out[moduleName], Route{
			Module:     moduleName,
			Deployment: route.DeploymentName,
			Runner:     model.RunnerKey(route.RunnerKey),
			Endpoint:   route.Endpoint,
		})
	}
	return out, nil
}

func (d *DAL) GetRunnerState(ctx context.Context, runnerKey model.RunnerKey) (RunnerState, error) {
	state, err := d.db.GetRunnerState(ctx, sql.Key(runnerKey))
	if err != nil {
		return "", translatePGError(err)
	}
	return RunnerState(state), nil
}

func (d *DAL) GetRunner(ctx context.Context, runnerKey model.RunnerKey) (Runner, error) {
	row, err := d.db.GetRunner(ctx, sql.Key(runnerKey))
	if err != nil {
		return Runner{}, translatePGError(err)
	}
	return runnerFromDB(row), nil
}

func (d *DAL) ExpireRunnerClaims(ctx context.Context) (int64, error) {
	count, err := d.db.ExpireRunnerReservations(ctx)
	return count, translatePGError(err)
}

func (d *DAL) InsertLogEvent(ctx context.Context, log *LogEvent) error {
	attributes, err := json.Marshal(log.Attributes)
	if err != nil {
		return err
	}
	var requestName optional.Option[string]
	if name, ok := log.RequestName.Get(); ok {
		requestName = optional.Some(string(name))
	}
	return translatePGError(d.db.InsertLogEvent(ctx, sql.InsertLogEventParams{
		DeploymentName: log.DeploymentName,
		RequestName:    requestName,
		TimeStamp:      log.Time,
		Level:          log.Level,
		Attributes:     attributes,
		Message:        log.Message,
		Error:          log.Error,
		Stack:          log.Stack,
	}))
}

func (d *DAL) loadDeployment(ctx context.Context, deployment sql.GetDeploymentRow) (*model.Deployment, error) {
	out := &model.Deployment{
		Module:   deployment.ModuleName,
		Language: deployment.Language,
		Name:     deployment.Deployment.Name,
		Schema:   deployment.Deployment.Schema,
	}
	artefacts, err := d.db.GetDeploymentArtefacts(ctx, deployment.Deployment.ID)
	if err != nil {
		return nil, translatePGError(err)
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

func (d *DAL) CreateIngressRequest(ctx context.Context, route, addr string) (model.RequestName, error) {
	name := model.NewRequestName(model.OriginIngress, route)
	err := d.db.CreateIngressRequest(ctx, sql.OriginIngress, string(name), addr)
	return name, err
}

func (d *DAL) GetIngressRoutes(ctx context.Context, method string) ([]IngressRoute, error) {
	routes, err := d.db.GetIngressRoutes(ctx, method)
	if err != nil {
		return nil, translatePGError(err)
	}
	if len(routes) == 0 {
		return nil, ErrNotFound
	}
	return slices.Map(routes, func(row sql.GetIngressRoutesRow) IngressRoute {
		return IngressRoute{
			Runner:     model.RunnerKey(row.RunnerKey),
			Deployment: row.DeploymentName,
			Endpoint:   row.Endpoint,
			Path:       row.Path,
			Module:     row.Module,
			Verb:       row.Verb,
		}
	}), nil
}

func (d *DAL) UpsertController(ctx context.Context, key model.ControllerKey, addr string) (int64, error) {
	id, err := d.db.UpsertController(ctx, key, addr)
	return id, translatePGError(err)
}

func (d *DAL) InsertCallEvent(ctx context.Context, call *CallEvent) error {
	var sourceModule, sourceVerb optional.Option[string]
	if sr, ok := call.SourceVerb.Get(); ok {
		sourceModule, sourceVerb = optional.Some(sr.Module), optional.Some(sr.Name)
	}
	var requestName optional.Option[string]
	if rn, ok := call.RequestName.Get(); ok {
		requestName = optional.Some(string(rn))
	}
	return translatePGError(d.db.InsertCallEvent(ctx, sql.InsertCallEventParams{
		DeploymentName: call.DeploymentName.String(),
		RequestName:    requestName,
		TimeStamp:      call.Time,
		SourceModule:   sourceModule,
		SourceVerb:     sourceVerb,
		DestModule:     call.DestVerb.Module,
		DestVerb:       call.DestVerb.Name,
		DurationMs:     call.Duration.Milliseconds(),
		Request:        call.Request,
		Response:       call.Response,
		Error:          call.Error,
		Stack:          call.Stack,
	}))
}

func (d *DAL) GetActiveRunners(ctx context.Context) ([]Runner, error) {
	rows, err := d.db.GetActiveRunners(ctx, false)
	if err != nil {
		return nil, translatePGError(err)
	}
	return slices.Map(rows, func(row sql.GetActiveRunnersRow) Runner {
		return runnerFromDB(sql.GetRunnerRow(row))
	}), nil
}

// Check if a deployment exists that exactly matches the given artefacts and schema.
func (*DAL) checkForExistingDeployments(ctx context.Context, tx *sql.Tx, moduleSchema *schema.Module, artefacts []DeploymentArtefact) (model.DeploymentName, error) {
	schemaBytes, err := schema.ModuleToBytes(moduleSchema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}
	existing, err := tx.GetDeploymentsWithArtefacts(ctx,
		sha256esToBytes(slices.Map(artefacts, func(in DeploymentArtefact) sha256.SHA256 { return in.Digest })),
		schemaBytes,
		int64(len(artefacts)),
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", "couldn't check for existing deployment", err)
	}
	if len(existing) > 0 {
		return existing[0].DeploymentName, nil
	}
	return "", nil
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
		return 0, translatePGError(err)
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
			return fmt.Errorf("%s: %w", strings.TrimSuffix(strings.TrimPrefix(pgErr.ConstraintName, pgErr.TableName+"_"), "_id_fkey"), ErrNotFound)
		} else if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrConflict
		}
	} else if isNotFound(err) {
		return ErrNotFound
	}
	return err
}
