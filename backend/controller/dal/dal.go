// Package dal provides a data abstraction layer for the Controller
package dal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	sets "github.com/deckarep/golang-set/v2"
	"google.golang.org/protobuf/proto"

	dalsql "github.com/TBD54566975/ftl/backend/controller/dal/internal/sql"
	"github.com/TBD54566975/ftl/backend/controller/encryption"
	leasedal "github.com/TBD54566975/ftl/backend/controller/leases/dal"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/backend/libdal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlencryption "github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
)

type IngressRoute struct {
	Runner     model.RunnerKey
	Deployment model.DeploymentKey
	Endpoint   string
	Path       string
	Module     string
	Verb       string
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
		return DeploymentArtefact{}, err
	}
	return DeploymentArtefact{
		Digest:     digest,
		Executable: in.Executable,
		Path:       in.Path,
	}, nil
}

func runnerFromDB(row dalsql.GetRunnerRow) Runner {

	attrs := model.Labels{}
	if err := json.Unmarshal(row.Labels, &attrs); err != nil {
		return Runner{}
	}

	return Runner{
		Key:        row.RunnerKey,
		Endpoint:   row.Endpoint,
		State:      RunnerState(row.State),
		Deployment: row.DeploymentKey,
		Labels:     attrs,
	}
}

type Runner struct {
	Key                model.RunnerKey
	Endpoint           string
	State              RunnerState
	ReservationTimeout optional.Option[time.Duration]
	Module             optional.Option[string]
	Deployment         model.DeploymentKey
	Labels             model.Labels
}

func (r Runner) notification() {}

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
	RunnerStateNew      = RunnerState(dalsql.RunnerStateNew)
	RunnerStateAssigned = RunnerState(dalsql.RunnerStateAssigned)
	RunnerStateDead     = RunnerState(dalsql.RunnerStateDead)
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
	ControllerStateLive = ControllerState(dalsql.ControllerStateLive)
	ControllerStateDead = ControllerState(dalsql.ControllerStateDead)
)

type RequestOrigin string

const (
	RequestOriginIngress = RequestOrigin(dalsql.OriginIngress)
	RequestOriginCron    = RequestOrigin(dalsql.OriginCron)
	RequestOriginPubsub  = RequestOrigin(dalsql.OriginPubsub)
)

type Deployment struct {
	Key         model.DeploymentKey
	Language    string
	Module      string
	MinReplicas int
	Replicas    optional.Option[int] // Depending on the query this may or may not be populated.
	Schema      *schema.Module
	CreatedAt   time.Time
	Labels      model.Labels
}

func (d Deployment) String() string { return d.Key.String() }

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
	Deployment model.DeploymentKey
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

func New(ctx context.Context, conn libdal.Connection, encryption *encryption.Service) *DAL {
	var d *DAL
	d = &DAL{
		leaseDAL:   leasedal.New(conn),
		db:         dalsql.New(conn),
		encryption: encryption,
		Handle: libdal.New(conn, func(h *libdal.Handle[DAL]) *DAL {
			return &DAL{
				Handle:            h,
				db:                dalsql.New(h.Connection),
				leaseDAL:          leasedal.New(h.Connection),
				encryption:        d.encryption,
				DeploymentChanges: d.DeploymentChanges,
			}
		}),
		DeploymentChanges: pubsub.New[DeploymentNotification](),
	}

	return d
}

type DAL struct {
	*libdal.Handle[DAL]
	db dalsql.Querier

	leaseDAL   *leasedal.DAL
	encryption *encryption.Service

	// DeploymentChanges is a Topic that receives changes to the deployments table.
	DeploymentChanges *pubsub.Topic[DeploymentNotification]
}

func (d *DAL) GetActiveControllers(ctx context.Context) ([]Controller, error) {
	controllers, err := d.db.GetActiveControllers(ctx)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	return slices.Map(controllers, func(in dalsql.Controller) Controller {
		return Controller{
			Key:      in.Key,
			Endpoint: in.Endpoint,
		}
	}), nil
}

func (d *DAL) GetStatus(ctx context.Context) (Status, error) {
	controllers, err := d.GetActiveControllers(ctx)
	if err != nil {
		return Status{}, fmt.Errorf("could not get control planes: %w", libdal.TranslatePGError(err))
	}
	runners, err := d.db.GetActiveRunners(ctx)
	if err != nil {
		return Status{}, fmt.Errorf("could not get active runners: %w", libdal.TranslatePGError(err))
	}
	deployments, err := d.db.GetActiveDeployments(ctx)
	if err != nil {
		return Status{}, fmt.Errorf("could not get active deployments: %w", libdal.TranslatePGError(err))
	}
	ingressRoutes, err := d.db.GetActiveIngressRoutes(ctx)
	if err != nil {
		return Status{}, fmt.Errorf("could not get ingress routes: %w", libdal.TranslatePGError(err))
	}
	routes, err := d.db.GetRoutingTable(ctx, nil)
	if err != nil {
		return Status{}, fmt.Errorf("could not get routing table: %w", libdal.TranslatePGError(err))
	}
	statusDeployments, err := slices.MapErr(deployments, func(in dalsql.GetActiveDeploymentsRow) (Deployment, error) {
		labels := model.Labels{}
		err = json.Unmarshal(in.Deployment.Labels, &labels)
		if err != nil {
			return Deployment{}, fmt.Errorf("%q: invalid labels in database: %w", in.ModuleName, err)
		}
		return Deployment{
			Key:         in.Deployment.Key,
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
	domainRunners, err := slices.MapErr(runners, func(in dalsql.GetActiveRunnersRow) (Runner, error) {
		attrs := model.Labels{}
		if err := json.Unmarshal(in.Labels, &attrs); err != nil {
			return Runner{}, fmt.Errorf("invalid attributes JSON for runner %s: %w", in.RunnerKey, err)
		}

		return Runner{
			Key:        in.RunnerKey,
			Endpoint:   in.Endpoint,
			State:      RunnerState(in.State),
			Deployment: in.DeploymentKey,
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
		IngressRoutes: slices.Map(ingressRoutes, func(in dalsql.GetActiveIngressRoutesRow) IngressRouteEntry {
			return IngressRouteEntry{
				Deployment: in.DeploymentKey,
				Module:     in.Module,
				Verb:       in.Verb,
				Method:     in.Method,
				Path:       in.Path,
			}
		}),
		Routes: slices.Map(routes, func(row dalsql.GetRoutingTableRow) Route {
			return Route{
				Module:     row.ModuleName.MustGet(),
				Runner:     row.RunnerKey,
				Deployment: row.DeploymentKey.MustGet(),
				Endpoint:   row.Endpoint,
			}
		}),
	}, nil
}

func (d *DAL) GetRunnersForDeployment(ctx context.Context, deployment model.DeploymentKey) ([]Runner, error) {
	runners := []Runner{}
	rows, err := d.db.GetRunnersForDeployment(ctx, deployment)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	for _, row := range rows {
		attrs := model.Labels{}
		if err := json.Unmarshal(row.Labels, &attrs); err != nil {
			return nil, fmt.Errorf("invalid attributes JSON for runner %d: %w", row.ID, err)
		}

		runners = append(runners, Runner{
			Key:        row.Key,
			Endpoint:   row.Endpoint,
			State:      RunnerState(row.State),
			Deployment: deployment,
			Labels:     attrs,
		})
	}
	return runners, nil
}

func (d *DAL) UpsertModule(ctx context.Context, language, name string) (err error) {
	_, err = d.db.UpsertModule(ctx, language, name)
	return libdal.TranslatePGError(err)
}

// GetMissingArtefacts returns the digests of the artefacts that are missing from the database.
func (d *DAL) GetMissingArtefacts(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error) {
	have, err := d.db.GetArtefactDigests(ctx, sha256esToBytes(digests))
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	haveStr := slices.Map(have, func(in dalsql.GetArtefactDigestsRow) sha256.SHA256 {
		return sha256.FromBytes(in.Digest)
	})
	return sets.NewSet(digests...).Difference(sets.NewSet(haveStr...)).ToSlice(), nil
}

// CreateArtefact inserts a new artefact into the database and returns its ID.
func (d *DAL) CreateArtefact(ctx context.Context, content []byte) (digest sha256.SHA256, err error) {
	sha256digest := sha256.Sum(content)
	_, err = d.db.CreateArtefact(ctx, sha256digest[:], content)
	return sha256digest, libdal.TranslatePGError(err)
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
func (d *DAL) CreateDeployment(ctx context.Context, language string, moduleSchema *schema.Module, artefacts []DeploymentArtefact, ingressRoutes []IngressRoutingEntry, cronJobs []model.CronJob) (key model.DeploymentKey, err error) {
	logger := log.FromContext(ctx)

	// Start the parent transaction
	tx, err := d.Begin(ctx)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	existingDeployment, err := d.checkForExistingDeployments(ctx, tx, moduleSchema, artefacts)
	if err != nil {
		return model.DeploymentKey{}, err
	} else if !existingDeployment.IsZero() {
		logger.Tracef("Returning existing deployment %s", existingDeployment)
		return existingDeployment, nil
	}

	artefactsByDigest := maps.FromSlice(artefacts, func(in DeploymentArtefact) (sha256.SHA256, DeploymentArtefact) {
		return in.Digest, in
	})

	schemaBytes, err := proto.Marshal(moduleSchema.ToProto())
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("failed to marshal schema: %w", err)
	}

	// TODO(aat): "schema" containing language?
	_, err = tx.db.UpsertModule(ctx, language, moduleSchema.Name)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("failed to upsert module: %w", libdal.TranslatePGError(err))
	}

	// upsert topics
	for _, decl := range moduleSchema.Decls {
		t, ok := decl.(*schema.Topic)
		if !ok {
			continue
		}
		err := tx.db.UpsertTopic(ctx, dalsql.UpsertTopicParams{
			Topic:     model.NewTopicKey(moduleSchema.Name, t.Name),
			Module:    moduleSchema.Name,
			Name:      t.Name,
			EventType: t.Event.String(),
		})
		if err != nil {
			return model.DeploymentKey{}, fmt.Errorf("could not insert topic: %w", libdal.TranslatePGError(err))
		}
	}

	deploymentKey := model.NewDeploymentKey(moduleSchema.Name)

	// Create the deployment
	err = tx.db.CreateDeployment(ctx, moduleSchema.Name, schemaBytes, deploymentKey)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("failed to create deployment: %w", libdal.TranslatePGError(err))
	}

	uploadedDigests := slices.Map(artefacts, func(in DeploymentArtefact) []byte { return in.Digest[:] })
	artefactDigests, err := tx.db.GetArtefactDigests(ctx, uploadedDigests)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("failed to get artefact digests: %w", err)
	}
	if len(artefactDigests) != len(artefacts) {
		missingDigests := strings.Join(slices.Map(artefacts, func(in DeploymentArtefact) string { return in.Digest.String() }), ", ")
		return model.DeploymentKey{}, fmt.Errorf("missing %d artefacts: %s", len(artefacts)-len(artefactDigests), missingDigests)
	}

	// Associate the artefacts with the deployment
	for _, row := range artefactDigests {
		artefact := artefactsByDigest[sha256.FromBytes(row.Digest)]
		err = tx.db.AssociateArtefactWithDeployment(ctx, dalsql.AssociateArtefactWithDeploymentParams{
			Key:        deploymentKey,
			ArtefactID: row.ID,
			Executable: artefact.Executable,
			Path:       artefact.Path,
		})
		if err != nil {
			return model.DeploymentKey{}, fmt.Errorf("failed to associate artefact with deployment: %w", libdal.TranslatePGError(err))
		}
	}

	for _, ingressRoute := range ingressRoutes {
		err = tx.db.CreateIngressRoute(ctx, dalsql.CreateIngressRouteParams{
			Key:    deploymentKey,
			Method: ingressRoute.Method,
			Path:   ingressRoute.Path,
			Module: moduleSchema.Name,
			Verb:   ingressRoute.Verb,
		})
		if err != nil {
			return model.DeploymentKey{}, fmt.Errorf("failed to create ingress route: %w", libdal.TranslatePGError(err))
		}
	}

	for _, job := range cronJobs {
		// Start time must be calculated by the caller rather than generated by db
		// This ensures that nextExecution is after start time, otherwise the job will never be triggered
		err := tx.db.CreateCronJob(ctx, dalsql.CreateCronJobParams{
			Key:           job.Key,
			DeploymentKey: deploymentKey,
			ModuleName:    job.Verb.Module,
			Verb:          job.Verb.Name,
			StartTime:     job.StartTime,
			Schedule:      job.Schedule,
			NextExecution: job.NextExecution,
		})
		if err != nil {
			return model.DeploymentKey{}, fmt.Errorf("failed to create cron job: %w", libdal.TranslatePGError(err))
		}
	}

	return deploymentKey, nil
}

func (d *DAL) GetDeployment(ctx context.Context, key model.DeploymentKey) (*model.Deployment, error) {
	deployment, err := d.db.GetDeployment(ctx, key)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	return d.loadDeployment(ctx, deployment)
}

// UpsertRunner registers or updates a new runner.
//
// ErrConflict will be returned if a runner with the same endpoint and a
// different key already exists.
func (d *DAL) UpsertRunner(ctx context.Context, runner Runner) error {
	attrBytes, err := json.Marshal(runner.Labels)
	if err != nil {
		return fmt.Errorf("failed to JSON encode runner labels: %w", err)
	}
	deploymentID, err := d.db.UpsertRunner(ctx, dalsql.UpsertRunnerParams{
		Key:           runner.Key,
		Endpoint:      runner.Endpoint,
		State:         dalsql.RunnerState(runner.State),
		DeploymentKey: runner.Deployment,
		Labels:        attrBytes,
	})
	if err != nil {
		return libdal.TranslatePGError(err)
	}
	if deploymentID < 0 {
		return fmt.Errorf("deployment %s not found", runner.Deployment)
	}
	return nil
}

// KillStaleRunners deletes runners that have not had heartbeats for the given duration.
func (d *DAL) KillStaleRunners(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.KillStaleRunners(ctx, sqltypes.Duration(age))
	return count, err
}

// KillStaleControllers deletes controllers that have not had heartbeats for the given duration.
func (d *DAL) KillStaleControllers(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.KillStaleControllers(ctx, sqltypes.Duration(age))
	return count, err
}

// DeregisterRunner deregisters the given runner.
func (d *DAL) DeregisterRunner(ctx context.Context, key model.RunnerKey) error {
	count, err := d.db.DeregisterRunner(ctx, key)
	if err != nil {
		return libdal.TranslatePGError(err)
	}
	if count == 0 {
		return libdal.ErrNotFound
	}
	return nil
}

var _ Reservation = (*postgresClaim)(nil)

type postgresClaim struct {
	tx     *DAL
	runner Runner
	cancel context.CancelFunc
}

func (p *postgresClaim) Commit(ctx context.Context) error {
	defer p.cancel()
	return libdal.TranslatePGError(p.tx.Commit(ctx))
}

func (p *postgresClaim) Rollback(ctx context.Context) error {
	defer p.cancel()
	return libdal.TranslatePGError(p.tx.Rollback(ctx))
}

func (p *postgresClaim) Runner() Runner { return p.runner }

// SetDeploymentReplicas activates the given deployment.
func (d *DAL) SetDeploymentReplicas(ctx context.Context, key model.DeploymentKey, minReplicas int) (err error) {
	// Start the transaction
	tx, err := d.Begin(ctx)
	if err != nil {
		return libdal.TranslatePGError(err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	deployment, err := tx.db.GetDeployment(ctx, key)
	if err != nil {
		return libdal.TranslatePGError(err)
	}

	err = tx.db.SetDeploymentDesiredReplicas(ctx, key, int32(minReplicas))
	if err != nil {
		return libdal.TranslatePGError(err)
	}
	if minReplicas == 0 {
		err = tx.deploymentWillDeactivate(ctx, key)
		if err != nil {
			return libdal.TranslatePGError(err)
		}
	} else if deployment.MinReplicas == 0 {
		err = tx.deploymentWillActivate(ctx, key)
		if err != nil {
			return libdal.TranslatePGError(err)
		}
	}
	var payload ftlencryption.EncryptedTimelineColumn
	err = d.encryption.EncryptJSON(map[string]interface{}{
		"prev_min_replicas": deployment.MinReplicas,
		"min_replicas":      minReplicas,
	}, &payload)
	if err != nil {
		return fmt.Errorf("failed to encrypt payload for InsertDeploymentUpdatedEvent: %w", err)
	}
	err = tx.db.InsertTimelineDeploymentUpdatedEvent(ctx, dalsql.InsertTimelineDeploymentUpdatedEventParams{
		DeploymentKey: key,
		Payload:       payload,
	})
	if err != nil {
		return libdal.TranslatePGError(err)
	}

	return nil
}

var ErrReplaceDeploymentAlreadyActive = errors.New("deployment already active")

// ReplaceDeployment replaces an old deployment of a module with a new deployment.
//
// returns ErrReplaceDeploymentAlreadyActive if the new deployment is already active.
func (d *DAL) ReplaceDeployment(ctx context.Context, newDeploymentKey model.DeploymentKey, minReplicas int) (err error) {
	// Start the transaction
	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("replace deployment failed to begin transaction for %v: %w", newDeploymentKey, libdal.TranslatePGError(err))
	}

	defer tx.CommitOrRollback(ctx, &err)
	newDeployment, err := tx.db.GetDeployment(ctx, newDeploymentKey)
	if err != nil {
		return fmt.Errorf("replace deployment failed to get deployment for %v: %w", newDeploymentKey, libdal.TranslatePGError(err))
	}

	// must be called before deploymentWillDeactivate for the old deployment
	err = tx.deploymentWillActivate(ctx, newDeploymentKey)
	if err != nil {
		return fmt.Errorf("replace deployment failed willActivate trigger for %v: %w", newDeploymentKey, libdal.TranslatePGError(err))
	}

	// If there's an existing deployment, set its desired replicas to 0
	var replacedDeploymentKey optional.Option[model.DeploymentKey]
	oldDeployment, err := tx.db.GetExistingDeploymentForModule(ctx, newDeployment.ModuleName)
	if err == nil {
		if oldDeployment.Key.String() == newDeploymentKey.String() {
			return fmt.Errorf("replace deployment failed: deployment already exists from %v to %v: %w", oldDeployment.Key, newDeploymentKey, ErrReplaceDeploymentAlreadyActive)
		}
		err = tx.db.SetDeploymentDesiredReplicas(ctx, oldDeployment.Key, 0)
		if err != nil {
			return fmt.Errorf("replace deployment failed to set old deployment replicas from %v to %v: %w", oldDeployment.Key, newDeploymentKey, libdal.TranslatePGError(err))
		}
		err = tx.db.SetDeploymentDesiredReplicas(ctx, newDeploymentKey, int32(minReplicas))
		if err != nil {
			return fmt.Errorf("replace deployment failed to set new deployment replicas from %v to %v: %w", oldDeployment.Key, newDeploymentKey, libdal.TranslatePGError(err))
		}
		err = d.deploymentWillDeactivate(ctx, oldDeployment.Key)
		if err != nil {
			return fmt.Errorf("replace deployment failed willDeactivate trigger from %v to %v: %w", oldDeployment.Key, newDeploymentKey, libdal.TranslatePGError(err))
		}
		replacedDeploymentKey = optional.Some(oldDeployment.Key)
	} else if !libdal.IsNotFound(err) {
		// any error other than not found
		return fmt.Errorf("replace deployment failed to get existing deployment for %v: %w", newDeploymentKey, libdal.TranslatePGError(err))
	} else {
		// Set the desired replicas for the new deployment
		err = tx.db.SetDeploymentDesiredReplicas(ctx, newDeploymentKey, int32(minReplicas))
		if err != nil {
			return fmt.Errorf("replace deployment failed to set replicas for %v: %w", newDeploymentKey, libdal.TranslatePGError(err))
		}
	}

	var payload ftlencryption.EncryptedTimelineColumn
	err = d.encryption.EncryptJSON(map[string]any{
		"min_replicas": int32(minReplicas),
		"replaced":     replacedDeploymentKey,
	}, &payload)
	if err != nil {
		return fmt.Errorf("replace deployment failed to encrypt payload: %w", err)
	}

	err = tx.db.InsertTimelineDeploymentCreatedEvent(ctx, dalsql.InsertTimelineDeploymentCreatedEventParams{
		DeploymentKey: newDeploymentKey,
		Language:      newDeployment.Language,
		ModuleName:    newDeployment.ModuleName,
		Payload:       payload,
	})
	if err != nil {
		return fmt.Errorf("replace deployment failed to create event: %w", libdal.TranslatePGError(err))
	}

	return nil
}

// deploymentWillActivate is called whenever a deployment goes from min_replicas=0 to min_replicas>0.
//
// when replacing a deployment, this should be called first before calling deploymentWillDeactivate on the old deployment.
// This allows the new deployment to migrate from the old deployment (such as subscriptions).
func (d *DAL) deploymentWillActivate(ctx context.Context, key model.DeploymentKey) error {
	module, err := d.db.GetSchemaForDeployment(ctx, key)
	if err != nil {
		return fmt.Errorf("could not get schema: %w", libdal.TranslatePGError(err))
	}
	err = d.createSubscriptions(ctx, key, module)
	if err != nil {
		return err
	}
	return d.createSubscribers(ctx, key, module)
}

// deploymentWillDeactivate is called whenever a deployment goes to min_replicas=0.
//
// it may be called when min_replicas was already 0
func (d *DAL) deploymentWillDeactivate(ctx context.Context, key model.DeploymentKey) error {
	return d.removeSubscriptionsAndSubscribers(ctx, key)
}

// GetActiveDeployments returns all active deployments.
func (d *DAL) GetActiveDeployments(ctx context.Context) ([]Deployment, error) {
	rows, err := d.db.GetActiveDeployments(ctx)
	if err != nil {
		if libdal.IsNotFound(err) {
			return nil, nil
		}
		return nil, libdal.TranslatePGError(err)
	}
	return slices.Map(rows, func(in dalsql.GetActiveDeploymentsRow) Deployment {
		return Deployment{
			Key:         in.Deployment.Key,
			Module:      in.ModuleName,
			Language:    in.Language,
			MinReplicas: int(in.Deployment.MinReplicas),
			Replicas:    optional.Some(int(in.Replicas)),
			Schema:      in.Deployment.Schema,
			CreatedAt:   in.Deployment.CreatedAt,
		}
	}), nil
}

// GetActiveSchema returns the schema for all active deployments.
func (d *DAL) GetActiveSchema(ctx context.Context) (*schema.Schema, error) {
	deployments, err := d.GetActiveDeployments(ctx)
	if err != nil {
		return nil, err
	}
	sch, err := schema.ValidateSchema(&schema.Schema{
		Modules: slices.Map(deployments, func(d Deployment) *schema.Module {
			return d.Schema
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("could not validate schema: %w", err)
	}
	return sch, nil
}

func (d *DAL) GetDeploymentsWithMinReplicas(ctx context.Context) ([]Deployment, error) {
	rows, err := d.db.GetDeploymentsWithMinReplicas(ctx)
	if err != nil {
		if libdal.IsNotFound(err) {
			return nil, nil
		}
		return nil, libdal.TranslatePGError(err)
	}
	return slices.Map(rows, func(in dalsql.GetDeploymentsWithMinReplicasRow) Deployment {
		return Deployment{
			Key:         in.Deployment.Key,
			Module:      in.ModuleName,
			Language:    in.Language,
			MinReplicas: int(in.Deployment.MinReplicas),
			Schema:      in.Deployment.Schema,
			CreatedAt:   in.Deployment.CreatedAt,
		}
	}), nil
}

func (d *DAL) GetActiveDeploymentSchemas(ctx context.Context) ([]*schema.Module, error) {
	rows, err := d.db.GetActiveDeploymentSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get active deployments: %w", libdal.TranslatePGError(err))
	}
	return slices.Map(rows, func(in dalsql.GetActiveDeploymentSchemasRow) *schema.Module { return in.Schema }), nil
}

type ProcessRunner struct {
	Key      model.RunnerKey
	Endpoint string
	Labels   model.Labels
}

type Process struct {
	Deployment  model.DeploymentKey
	MinReplicas int
	Labels      model.Labels
	Runner      optional.Option[ProcessRunner]
}

// GetProcessList returns a list of all "processes".
func (d *DAL) GetProcessList(ctx context.Context) ([]Process, error) {
	rows, err := d.db.GetProcessList(ctx)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	return slices.MapErr(rows, func(row dalsql.GetProcessListRow) (Process, error) { //nolint:wrapcheck
		var runner optional.Option[ProcessRunner]
		if endpoint, ok := row.Endpoint.Get(); ok {
			var labels model.Labels
			if err := json.Unmarshal(row.RunnerLabels.RawMessage, &labels); err != nil {
				return Process{}, fmt.Errorf("invalid labels JSON for runner %s: %w", row.RunnerKey, err)
			}

			runner = optional.Some(ProcessRunner{
				Key:      row.RunnerKey.MustGet(),
				Endpoint: endpoint,
				Labels:   labels,
			})
		}
		var labels model.Labels
		if err := json.Unmarshal(row.DeploymentLabels, &labels); err != nil {
			return Process{}, fmt.Errorf("invalid labels JSON for deployment %s: %w", row.DeploymentKey, err)
		}
		return Process{
			Deployment:  row.DeploymentKey,
			Labels:      labels,
			MinReplicas: int(row.MinReplicas),
			Runner:      runner,
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
		return nil, libdal.TranslatePGError(err)
	}
	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found: %w", libdal.ErrNotFound)
	}
	out := make(map[string][]Route, len(routes))
	for _, route := range routes {
		// This is guaranteed to be non-nil by the query, but sqlc doesn't quite understand that.
		moduleName := route.ModuleName.MustGet()
		deploymentKey := route.DeploymentKey.MustGet()
		out[moduleName] = append(out[moduleName], Route{
			Module:     moduleName,
			Deployment: deploymentKey,
			Runner:     route.RunnerKey,
			Endpoint:   route.Endpoint,
		})
	}
	return out, nil
}

func (d *DAL) GetRunnerState(ctx context.Context, runnerKey model.RunnerKey) (RunnerState, error) {
	state, err := d.db.GetRunnerState(ctx, runnerKey)
	if err != nil {
		return "", libdal.TranslatePGError(err)
	}
	return RunnerState(state), nil
}

func (d *DAL) GetRunner(ctx context.Context, runnerKey model.RunnerKey) (Runner, error) {
	row, err := d.db.GetRunner(ctx, runnerKey)
	if err != nil {
		return Runner{}, libdal.TranslatePGError(err)
	}
	return runnerFromDB(row), nil
}

func (d *DAL) InsertLogEvent(ctx context.Context, log *LogEvent) error {
	var requestKey optional.Option[string]
	if name, ok := log.RequestKey.Get(); ok {
		requestKey = optional.Some(name.String())
	}

	payload := map[string]any{
		"message":    log.Message,
		"attributes": log.Attributes,
		"error":      log.Error,
		"stack":      log.Stack,
	}
	var encryptedPayload ftlencryption.EncryptedTimelineColumn
	err := d.encryption.EncryptJSON(payload, &encryptedPayload)
	if err != nil {
		return fmt.Errorf("failed to encrypt log payload: %w", err)
	}
	return libdal.TranslatePGError(d.db.InsertTimelineLogEvent(ctx, dalsql.InsertTimelineLogEventParams{
		DeploymentKey: log.DeploymentKey,
		RequestKey:    requestKey,
		TimeStamp:     log.Time,
		Level:         log.Level,
		Payload:       encryptedPayload,
	}))
}

func (d *DAL) loadDeployment(ctx context.Context, deployment dalsql.GetDeploymentRow) (*model.Deployment, error) {
	out := &model.Deployment{
		Module:   deployment.ModuleName,
		Language: deployment.Language,
		Key:      deployment.Deployment.Key,
		Schema:   deployment.Deployment.Schema,
	}
	artefacts, err := d.db.GetDeploymentArtefacts(ctx, deployment.Deployment.ID)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	out.Artefacts = slices.Map(artefacts, func(row dalsql.GetDeploymentArtefactsRow) *model.Artefact {
		return &model.Artefact{
			Path:       row.Path,
			Executable: row.Executable,
			Content:    &artefactReader{id: row.ID, db: d.db},
			Digest:     sha256.FromBytes(row.Digest),
		}
	})
	return out, nil
}

func (d *DAL) CreateRequest(ctx context.Context, key model.RequestKey, addr string) error {
	if err := d.db.CreateRequest(ctx, dalsql.Origin(key.Payload.Origin), key, addr); err != nil {
		return libdal.TranslatePGError(err)
	}
	return nil
}

func (d *DAL) GetIngressRoutes(ctx context.Context, method string) ([]IngressRoute, error) {
	routes, err := d.db.GetIngressRoutes(ctx, method)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	if len(routes) == 0 {
		return nil, libdal.ErrNotFound
	}
	return slices.Map(routes, func(row dalsql.GetIngressRoutesRow) IngressRoute {
		return IngressRoute{
			Runner:     row.RunnerKey,
			Deployment: row.DeploymentKey,
			Endpoint:   row.Endpoint,
			Path:       row.Path,
			Module:     row.Module,
			Verb:       row.Verb,
		}
	}), nil
}

func (d *DAL) UpsertController(ctx context.Context, key model.ControllerKey, addr string) (int64, error) {
	id, err := d.db.UpsertController(ctx, key, addr)
	return id, libdal.TranslatePGError(err)
}

func (d *DAL) InsertCallEvent(ctx context.Context, call *CallEvent) error {
	var sourceModule, sourceVerb optional.Option[string]
	if sr, ok := call.SourceVerb.Get(); ok {
		sourceModule, sourceVerb = optional.Some(sr.Module), optional.Some(sr.Name)
	}
	var requestKey optional.Option[string]
	if rn, ok := call.RequestKey.Get(); ok {
		requestKey = optional.Some(rn.String())
	}
	var parentRequestKey optional.Option[string]
	if pr, ok := call.ParentRequestKey.Get(); ok {
		parentRequestKey = optional.Some(pr.String())
	}
	var payload ftlencryption.EncryptedTimelineColumn
	err := d.encryption.EncryptJSON(map[string]any{
		"duration_ms": call.Duration.Milliseconds(),
		"request":     call.Request,
		"response":    call.Response,
		"error":       call.Error,
		"stack":       call.Stack,
	}, &payload)
	if err != nil {
		return fmt.Errorf("failed to encrypt call payload: %w", err)
	}
	return libdal.TranslatePGError(d.db.InsertTimelineCallEvent(ctx, dalsql.InsertTimelineCallEventParams{
		DeploymentKey:    call.DeploymentKey,
		RequestKey:       requestKey,
		ParentRequestKey: parentRequestKey,
		TimeStamp:        call.Time,
		SourceModule:     sourceModule,
		SourceVerb:       sourceVerb,
		DestModule:       call.DestVerb.Module,
		DestVerb:         call.DestVerb.Name,
		Payload:          payload,
	}))
}

func (d *DAL) DeleteOldEvents(ctx context.Context, eventType EventType, age time.Duration) (int64, error) {
	count, err := d.db.DeleteOldTimelineEvents(ctx, sqltypes.Duration(age), eventType)
	return count, libdal.TranslatePGError(err)
}

func (d *DAL) GetActiveRunners(ctx context.Context) ([]Runner, error) {
	rows, err := d.db.GetActiveRunners(ctx)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	return slices.Map(rows, func(row dalsql.GetActiveRunnersRow) Runner {
		return runnerFromDB(dalsql.GetRunnerRow(row))
	}), nil
}

// Check if a deployment exists that exactly matches the given artefacts and schema.
func (*DAL) checkForExistingDeployments(ctx context.Context, tx *DAL, moduleSchema *schema.Module, artefacts []DeploymentArtefact) (model.DeploymentKey, error) {
	schemaBytes, err := schema.ModuleToBytes(moduleSchema)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("failed to marshal schema: %w", err)
	}
	existing, err := tx.db.GetDeploymentsWithArtefacts(ctx,
		sha256esToBytes(slices.Map(artefacts, func(in DeploymentArtefact) sha256.SHA256 { return in.Digest })),
		schemaBytes,
		int64(len(artefacts)),
	)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("couldn't check for existing deployment: %w", err)
	}
	if len(existing) > 0 {
		return existing[0].DeploymentKey, nil
	}
	return model.DeploymentKey{}, nil
}

func sha256esToBytes(digests []sha256.SHA256) [][]byte {
	return slices.Map(digests, func(digest sha256.SHA256) []byte { return digest[:] })
}

type artefactReader struct {
	id     int64
	db     dalsql.Querier
	offset int32
}

func (r *artefactReader) Close() error { return nil }

func (r *artefactReader) Read(p []byte) (n int, err error) {
	content, err := r.db.GetArtefactContentRange(context.Background(), r.offset+1, int32(len(p)), r.id)
	if err != nil {
		return 0, libdal.TranslatePGError(err)
	}
	copy(p, content)
	clen := len(content)
	r.offset += int32(clen)
	if clen == 0 {
		err = io.EOF
	}
	return clen, err
}
