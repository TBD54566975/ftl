// Package dal provides a domain abstraction over the ControlPlane database.
package dal

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"io"
	"net/url"
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

	"github.com/TBD54566975/ftl/controlplane/internal/sql"
	"github.com/TBD54566975/ftl/controlplane/internal/sqltypes"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/sha256"
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
	// ErrInvalidReference is returned by select methods in the DAL when a reference to another resource is invalid.
	ErrInvalidReference = errors.New("invalid reference")
	// ErrNotFound is returned by select methods in the DAL when no results are found.
	ErrNotFound = errors.New("not found")
)

type DAL struct {
	db *sql.DB
}

func New(conn sql.DBI) *DAL {
	return &DAL{db: sql.NewDB(conn)}
}

func (d *DAL) CreateModule(ctx context.Context, language, name string) (err error) {
	_, err = d.db.CreateModule(ctx, language, name)
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
func (d *DAL) CreateDeployment(
	ctx context.Context,
	language string,
	schema *schema.Module,
	artefacts []DeploymentArtefact,
) (key ulid.ULID, err error) {
	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return ulid.ULID{}, errors.WithStack(err)
	}
	defer tx.CommitOrRollback(ctx, &err)()

	// Check if the deployment already exists and if so, return it.
	existing, err := tx.GetDeploymentsWithArtefacts(ctx,
		sha256esToBytes(slices.Map(artefacts, func(in DeploymentArtefact) sha256.SHA256 { return in.Digest })),
		len(artefacts),
	)
	if err != nil {
		return ulid.ULID{}, errors.WithStack(err)
	}
	if len(existing) > 0 {
		return existing[0].Key.ULID(), nil
	}

	artefactsByDigest := maps.FromSlice(artefacts, func(in DeploymentArtefact) (sha256.SHA256, DeploymentArtefact) {
		return in.Digest, in
	})

	schemaBytes, err := proto.Marshal(schema.ToProto())
	if err != nil {
		return ulid.ULID{}, errors.WithStack(err)
	}

	// TODO(aat): "schema" containing language?
	_, err = tx.CreateModule(ctx, language, schema.Name)

	deploymentKey := ulid.Make()
	// Create the deployment
	err = tx.CreateDeployment(ctx, sqltypes.Key(deploymentKey), schema.Name, schemaBytes)
	if err != nil {
		return ulid.ULID{}, errors.WithStack(translatePGError(err))
	}

	uploadedDigests := slices.Map(artefacts, func(in DeploymentArtefact) []byte { return in.Digest[:] })
	artefactDigests, err := tx.GetArtefactDigests(ctx, uploadedDigests)
	if err != nil {
		return ulid.ULID{}, errors.WithStack(err)
	}
	if len(artefactDigests) != len(artefacts) {
		missingDigests := strings.Join(slices.Map(artefacts, func(in DeploymentArtefact) string { return in.Digest.String() }), ", ")
		return ulid.ULID{}, errors.Errorf("missing %d artefacts: %s", len(artefacts)-len(artefactDigests), missingDigests)
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
			return ulid.ULID{}, errors.WithStack(translatePGError(err))
		}
	}

	return deploymentKey, nil
}

type Deployment struct {
	Module    string
	Language  string
	Key       ulid.ULID
	Schema    *schema.Module
	Artefacts []*Artefact
}

type Artefact struct {
	Path       string
	Executable bool
	Digest     sha256.SHA256
	// ~Zero-cost on-demand reader.
	Content io.Reader
}

func (a *Artefact) ToProto() *ftlv1.DeploymentArtefact {
	return &ftlv1.DeploymentArtefact{
		Path:       a.Path,
		Executable: a.Executable,
		Digest:     a.Digest.String(),
	}
}

func (d *DAL) GetDeployment(ctx context.Context, id ulid.ULID) (*Deployment, error) {
	deployment, err := d.db.GetDeployment(ctx, sqltypes.Key(id))
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	return d.loadDeployment(ctx, sql.GetLatestDeploymentRow(deployment))
}

func (d *DAL) GetLatestDeployment(ctx context.Context, module string) (*Deployment, error) {
	deployment, err := d.db.GetLatestDeployment(ctx, module)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	return d.loadDeployment(ctx, deployment)
}

// RegisterRunner registers a new runner.
//
// It will return ErrConflict if a runner with the same endpoint already exists.
func (d *DAL) RegisterRunner(
	ctx context.Context,
	key ulid.ULID,
	language string,
	endpoint *url.URL,
	deployment types.Option[ulid.ULID],
) error {
	var deploymentSQLKey pgtype.UUID
	if deploymentKey, ok := deployment.Get(); ok {
		deploymentSQLKey.Valid = true
		deploymentSQLKey.Bytes = deploymentKey
	}
	_, err := d.db.RegisterRunner(ctx, sqltypes.Key(key), language, endpoint.String())
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	return nil
}

// DeleteStaleRunners deletes runners that have not had heartbeats for the given duration.
func (d *DAL) DeleteStaleRunners(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.DeleteStaleRunners(ctx, pgtype.Interval{
		Microseconds: int64(age / time.Microsecond),
		Valid:        true,
	})
	return count, errors.WithStack(err)
}

// DeregisterRunner deregisters the given runner.
func (d *DAL) DeregisterRunner(ctx context.Context, key ulid.ULID) error {
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
	Key      ulid.ULID
	Language string
	Endpoint string
	State    RunnerState
	// Assigned deployment key, if any.
	Deployment types.Option[ulid.ULID]
}

// ReserveRunnerForDeployment reserves a runner for the given deployment.
//
// Once a runner is reserved, it will be unavailable for other reservations
// or deployments and will not be returned by GetIdleRunnersForLanguage.
func (d *DAL) ReserveRunnerForDeployment(ctx context.Context, language string, deployment ulid.ULID) (Runner, error) {
	runner, err := d.db.ReserveRunners(ctx, language, 1, sqltypes.Key(deployment))
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
		Key:      runner.Key.ULID(),
		Language: runner.Language,
		Endpoint: runner.Endpoint,
		State:    RunnerState(runner.State),
	}, nil
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
			Key:      row.Key.ULID(),
			Language: row.Language,
			Endpoint: row.Endpoint,
			State:    RunnerState(row.State),
		}
	}), nil
}

// GetRunnersForModule returns all runners for the given module.
//
// If no runners are available, it will return an empty slice.
func (d *DAL) GetRunnersForModule(ctx context.Context, module string) ([]Runner, error) {
	runners, err := d.db.GetRunnersForModule(ctx, module)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	if len(runners) == 0 {
		return nil, errors.WithStack(ErrNotFound)
	}
	return slices.Map(runners, func(row sql.GetRunnersForModuleRow) Runner {
		return Runner{
			Key:        row.Key.ULID(),
			Language:   row.Language,
			Endpoint:   row.Endpoint,
			State:      RunnerState(row.State),
			Deployment: types.Some(row.DeploymentKey.ULID()),
		}
	}), nil
}

// GetRoutingTable returns the endpoints for all runners for the given module.
func (d *DAL) GetRoutingTable(ctx context.Context, module string) ([]string, error) {
	routes, err := d.db.GetRoutingTable(ctx, module)
	if len(routes) == 0 {
		return nil, errors.WithStack(ErrNotFound)
	}
	return routes, errors.WithStack(translatePGError(err))
}

type RunnerState string

// Runner states.
const (
	RunnerStateIdle     = RunnerState(sql.RunnersStateIdle)
	RunnerStateClaimed  = RunnerState(sql.RunnersStateClaimed)
	RunnerStateReserved = RunnerState(sql.RunnersStateReserved)
	RunnerStateAssigned = RunnerState(sql.RunnersStateAssigned)
)

// UpdateRunner updates the state of the given Runner.
func (d *DAL) UpdateRunner(ctx context.Context, runnerKey ulid.ULID, state RunnerState, deploymentKey types.Option[ulid.ULID]) error {
	var deploymentKeyField pgtype.UUID
	if key, ok := deploymentKey.Get(); ok {
		deploymentKeyField.Valid = true
		deploymentKeyField.Bytes = key
	}
	deploymentID, err := d.db.UpdateRunner(ctx, sqltypes.Key(runnerKey), sql.RunnersState(state), deploymentKeyField)
	if err != nil {
		return errors.WithStack(translatePGError(err))
	}
	if deploymentKey.Ok() && !deploymentID.Valid {
		return errors.Errorf("deployment %s not found", deploymentKey)
	}
	return nil
}

func (d *DAL) GetRunnerState(ctx context.Context, runnerKey ulid.ULID) (RunnerState, error) {
	state, err := d.db.GetRunnerState(ctx, sqltypes.Key(runnerKey))
	if err != nil {
		return "", errors.WithStack(translatePGError(err))
	}
	return RunnerState(state), nil
}

func (d *DAL) ExpireRunnerReservations(ctx context.Context) (int64, error) {
	count, err := d.db.ExpireRunnerReservations(ctx)
	return count, errors.WithStack(translatePGError(err))
}

func (d *DAL) InsertDeploymentLogEntry(ctx context.Context, deployment ulid.ULID, logEntry log.Entry) error {
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

func (d *DAL) loadDeployment(ctx context.Context, deployment sql.GetLatestDeploymentRow) (*Deployment, error) {
	pm := &pschema.Module{}
	err := proto.Unmarshal(deployment.Schema, pm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	module, err := schema.ModuleFromProto(pm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	out := &Deployment{
		Module:   deployment.ModuleName,
		Language: deployment.Language,
		Key:      deployment.Key.ULID(),
		Schema:   module,
	}
	artefacts, err := d.db.GetDeploymentArtefacts(ctx, deployment.ID)
	if err != nil {
		return nil, errors.WithStack(translatePGError(err))
	}
	out.Artefacts = slices.Map(artefacts, func(row sql.GetDeploymentArtefactsRow) *Artefact {
		return &Artefact{
			Path:       row.Path,
			Executable: row.Executable,
			Content:    &artefactReader{id: row.ID, db: d.db},
			Digest:     sha256.FromBytes(row.Digest),
		}
	})
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
			return errors.Wrap(ErrInvalidReference, strings.TrimSuffix(strings.TrimPrefix(pgErr.ConstraintName, pgErr.TableName+"_"), "_id_fkey"))
		} else if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrConflict
		}
	} else if isNotFound(err) {
		return ErrNotFound
	}
	return err
}
