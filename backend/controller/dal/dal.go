// Package dal provides a data abstraction layer for the Controller
package dal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/alecthomas/types/optional"
	inprocesspubsub "github.com/alecthomas/types/pubsub"
	xmaps "golang.org/x/exp/maps"

	aregistry "github.com/TBD54566975/ftl/backend/controller/artefacts"
	dalsql "github.com/TBD54566975/ftl/backend/controller/dal/internal/sql"
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
)

func New(ctx context.Context, conn libdal.Connection, pubsub *pubsub.Service, registry aregistry.Service) *DAL {
	var d *DAL
	db := dalsql.New(conn)
	d = &DAL{
		db:       db,
		registry: registry,
		Handle: libdal.New(conn, func(h *libdal.Handle[DAL]) *DAL {
			return &DAL{
				Handle:            h,
				db:                dalsql.New(h.Connection),
				pubsub:            pubsub,
				registry:          registry,
				DeploymentChanges: d.DeploymentChanges,
			}
		}),
		DeploymentChanges: inprocesspubsub.New[DeploymentNotification](),
	}

	return d
}

type DAL struct {
	*libdal.Handle[DAL]
	db dalsql.Querier

	pubsub   *pubsub.Service
	registry aregistry.Service

	// DeploymentChanges is a Topic that receives changes to the deployments table.
	DeploymentChanges *inprocesspubsub.Topic[DeploymentNotification]
}

func (d *DAL) GetDeploymentStatus(ctx context.Context) ([]dalmodel.Deployment, error) {
	deployments, err := d.db.GetActiveDeployments(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get active deployments: %w", libdal.TranslatePGError(err))
	}
	statusDeployments, err := slices.MapErr(deployments, func(in dalsql.GetActiveDeploymentsRow) (dalmodel.Deployment, error) {
		labels := model.Labels{}
		err = json.Unmarshal(in.Deployment.Labels, &labels)
		if err != nil {
			return dalmodel.Deployment{}, fmt.Errorf("%q: invalid labels in database: %w", in.ModuleName, err)
		}
		return dalmodel.Deployment{
			Key:         in.Deployment.Key,
			Module:      in.ModuleName,
			Language:    in.Language,
			MinReplicas: int(in.Deployment.MinReplicas),
			Schema:      in.Deployment.Schema,
			Labels:      labels,
		}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not parse deployments: %w", err)
	}

	return statusDeployments, nil
}

func (d *DAL) UpsertModule(ctx context.Context, language, name string) (err error) {
	_, err = d.db.UpsertModule(ctx, language, name)
	return libdal.TranslatePGError(err)
}

// CreateDeployment (possibly) creates a new deployment and associates
// previously created artefacts with it.
//
// If an existing deployment with identical artefacts exists, it is returned.
func (d *DAL) CreateDeployment(ctx context.Context, language string, moduleSchema *schema.Module, artefacts []dalmodel.DeploymentArtefact) (key model.DeploymentKey, err error) {
	logger := log.FromContext(ctx)

	// Start the parent transaction
	tx, err := d.Begin(ctx)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	existingDeployment, err := tx.checkForExistingDeployments(ctx, tx, moduleSchema, artefacts)
	if err != nil {
		return model.DeploymentKey{}, err
	} else if !existingDeployment.IsZero() {
		logger.Tracef("Returning existing deployment %s", existingDeployment)
		return existingDeployment, nil
	}

	artefactsByDigest := maps.FromSlice(artefacts, func(in dalmodel.DeploymentArtefact) (sha256.SHA256, dalmodel.DeploymentArtefact) {
		return in.Digest, in
	})

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
	err = tx.db.CreateDeployment(ctx, moduleSchema.Name, moduleSchema, deploymentKey)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("failed to create deployment: %w", libdal.TranslatePGError(err))
	}

	uploadedDigests := slices.Map(artefacts, func(in dalmodel.DeploymentArtefact) sha256.SHA256 { return sha256.FromBytes(in.Digest[:]) })
	keys, missing, err := tx.registry.GetDigestsKeys(ctx, uploadedDigests)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("failed to get artefact digests: %w", err)
	}
	if len(missing) > 0 {
		m := slices.Reduce(missing, "", func(join string, in sha256.SHA256) string {
			return fmt.Sprintf("%s, %s", join, in.String())
		})
		return model.DeploymentKey{}, fmt.Errorf("missing digests %s", m)
	}

	// Associate the artefacts with the deployment
	for _, row := range keys {
		artefact := artefactsByDigest[row.Digest]
		err = tx.db.AssociateArtefactWithDeployment(ctx, dalsql.AssociateArtefactWithDeploymentParams{
			Key:        deploymentKey,
			Digest:     row.Digest[:],
			Executable: artefact.Executable,
			Path:       artefact.Path,
		})
		if err != nil {
			return model.DeploymentKey{}, fmt.Errorf("failed to associate artefact with deployment: %w", libdal.TranslatePGError(err))
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
	timeline.Publish(ctx, timeline.DeploymentUpdated{
		DeploymentKey:   key,
		MinReplicas:     minReplicas,
		PrevMinReplicas: int(deployment.MinReplicas),
	})

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
		err = tx.db.SetDeploymentDesiredReplicas(ctx, newDeploymentKey, int32(minReplicas))
		if err != nil {
			return fmt.Errorf("replace deployment failed to set new deployment replicas from %v to %v: %w", oldDeployment.Key, newDeploymentKey, libdal.TranslatePGError(err))
		}
		err = tx.deploymentWillDeactivate(ctx, oldDeployment.Key)
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

	timeline.Publish(ctx, timeline.DeploymentCreated{
		DeploymentKey:      newDeploymentKey,
		Language:           newDeployment.Language,
		ModuleName:         newDeployment.ModuleName,
		MinReplicas:        minReplicas,
		ReplacedDeployment: replacedDeploymentKey,
	})
	if err != nil {
		return fmt.Errorf("replace deployment failed to create event: %w", libdal.TranslatePGError(err))
	}
	return nil
}

// deploymentWillActivate is called whenever a deployment goes from min_replicas=0 to min_replicas>0.
//
// When replacing a deployment, this should be called first before calling deploymentWillDeactivate on the old deployment.
// This allows the new deployment to migrate from the old deployment (such as subscriptions).
func (d *DAL) deploymentWillActivate(ctx context.Context, key model.DeploymentKey) error {
	module, err := d.db.GetSchemaForDeployment(ctx, key)
	if err != nil {
		return fmt.Errorf("could not get schema: %w", libdal.TranslatePGError(err))
	}
	err = d.pubsub.CreateSubscriptions(ctx, key, module)
	if err != nil {
		return err
	}
	err = d.pubsub.CreateSubscribers(ctx, key, module)
	if err != nil {
		return fmt.Errorf("could not create subscribers: %w", err)
	}
	return nil
}

// deploymentWillDeactivate is called whenever a deployment goes to min_replicas=0.
//
// it may be called when min_replicas was already 0
func (d *DAL) deploymentWillDeactivate(ctx context.Context, key model.DeploymentKey) error {
	err := d.pubsub.RemoveSubscriptionsAndSubscribers(ctx, key)
	if err != nil {
		return fmt.Errorf("could not remove subscriptions and subscribers: %w", err)
	}
	return nil
}

// GetActiveDeployments returns all active deployments.
func (d *DAL) GetActiveDeployments(ctx context.Context) ([]dalmodel.Deployment, error) {
	rows, err := d.db.GetActiveDeployments(ctx)
	if err != nil {
		if libdal.IsNotFound(err) {
			return nil, nil
		}
		return nil, libdal.TranslatePGError(err)
	}
	return slices.Map(rows, func(in dalsql.GetActiveDeploymentsRow) dalmodel.Deployment {
		return dalmodel.Deployment{
			Key:         in.Deployment.Key,
			Module:      in.ModuleName,
			Language:    in.Language,
			MinReplicas: int(in.Deployment.MinReplicas),
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

	schemaMap := map[string]*schema.Module{}
	for _, dep := range deployments {
		if _, ok := schemaMap[dep.Module]; !ok {
			// We only take the older ones
			// If new ones exist they are not live yet
			// Or the old ones would be gone
			schemaMap[dep.Module] = dep.Schema
		}
	}
	fullSchema := &schema.Schema{Modules: xmaps.Values(schemaMap)}
	sch, err := schema.ValidateSchema(fullSchema)
	if err != nil {
		return nil, fmt.Errorf("could not validate schema: %w", err)
	}
	return sch, nil
}

// UpdateModuleSchema updates the schema for a deployment in place.
//
// Note that this is racey as the deployment can be updated by another process. This will go away once we ditch the DB.
func (d *DAL) UpdateModuleSchema(ctx context.Context, deployment model.DeploymentKey, module *schema.Module) error {
	err := d.db.UpdateDeploymentSchema(ctx, module, deployment)
	if err != nil {
		return fmt.Errorf("failed to update deployment schema: %w", err)
	}
	return nil
}

func (d *DAL) GetDeploymentsWithMinReplicas(ctx context.Context) ([]dalmodel.Deployment, error) {
	rows, err := d.db.GetDeploymentsWithMinReplicas(ctx)
	if err != nil {
		if libdal.IsNotFound(err) {
			return nil, nil
		}
		return nil, libdal.TranslatePGError(err)
	}
	return slices.Map(rows, func(in dalsql.GetDeploymentsWithMinReplicasRow) dalmodel.Deployment {
		return dalmodel.Deployment{
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

// GetActiveDeploymentSchemasByDeploymentKey returns the schema for all active deployments by deployment key.
//
// model.DeploymentKey is not used directly as a key as it's not a valid map key.
func (d *DAL) GetActiveDeploymentSchemasByDeploymentKey(ctx context.Context) (map[string]*schema.Module, error) {
	rows, err := d.db.GetActiveDeploymentSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get active deployments: %w", libdal.TranslatePGError(err))
	}
	return maps.FromSlice(rows, func(in dalsql.GetActiveDeploymentSchemasRow) (string, *schema.Module) {
		return in.Key.String(), in.Schema
	}), nil
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
			Digest:     sha256.FromBytes(row.Digest),
		}
	})
	return out, nil
}

// Check if a deployment exists that exactly matches the given artefacts and schema.
func (*DAL) checkForExistingDeployments(ctx context.Context, tx *DAL, moduleSchema *schema.Module, artefacts []dalmodel.DeploymentArtefact) (model.DeploymentKey, error) {
	schemaBytes, err := schema.ModuleToBytes(moduleSchema)
	if err != nil {
		return model.DeploymentKey{}, fmt.Errorf("failed to marshal schema: %w", err)
	}
	existing, err := tx.db.GetDeploymentsWithArtefacts(ctx,
		sha256esToBytes(slices.Map(artefacts, func(in dalmodel.DeploymentArtefact) sha256.SHA256 { return in.Digest })),
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
