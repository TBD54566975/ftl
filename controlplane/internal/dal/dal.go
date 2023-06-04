// Package dal provides a domain abstraction over the ControlPlane database.
package dal

import (
	"context"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	sets "github.com/deckarep/golang-set/v2"
	"github.com/jackc/pgerrcode"
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

var ErrConflict = errors.New("conflict")

type DAL struct {
	db *sql.DB
}

func New(conn sql.DBI) *DAL {
	return &DAL{db: sql.NewDB(conn)}
}

func (d *DAL) CreateModule(ctx context.Context, language, name string) (err error) {
	_, err = d.db.CreateModule(ctx, language, name)
	return errors.WithStack(err)
}

// GetMissingArtefacts returns the digests of the artefacts that are missing from the database.
func (d *DAL) GetMissingArtefacts(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error) {
	have, err := d.db.GetArtefactDigests(ctx, sha256esToBytes(digests))
	if err != nil {
		return nil, errors.WithStack(err)
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
	return sha256digest, errors.WithStack(err)
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
		return ulid.ULID{}, errors.WithStack(err)
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
			return ulid.ULID{}, errors.WithStack(err)
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
		return nil, errors.WithStack(err)
	}
	return d.loadDeployment(ctx, sql.GetLatestDeploymentRow(deployment))
}

func (d *DAL) GetLatestDeployment(ctx context.Context, module string) (*Deployment, error) {
	deployment, err := d.db.GetLatestDeployment(ctx, module)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return d.loadDeployment(ctx, deployment)
}

type RunnerID int64

// RegisterRunner registers a new runner.
//
// It will return ErrConflict if a runner with the same endpoint already exists.
func (d *DAL) RegisterRunner(ctx context.Context, key ulid.ULID, language string, endpoint *url.URL) (RunnerID, error) {
	id, err := d.db.RegisterRunner(ctx, sqltypes.Key(key), language, endpoint.String())
	if isPGConflict(err) {
		return 0, errors.Wrap(ErrConflict, "runner already registered")
	}
	return RunnerID(id), errors.WithStack(err)
}

func (d *DAL) DeleteStaleRunners(ctx context.Context, age time.Duration) (int64, error) {
	count, err := d.db.DeleteStaleRunners(ctx, pgtype.Interval{
		Microseconds: int64(age / time.Microsecond),
		Valid:        true,
	})
	return count, errors.WithStack(err)
}

func (d *DAL) HeartbeatRunner(ctx context.Context, id RunnerID) error {
	return errors.WithStack(d.db.HeartbeatRunner(ctx, int64(id)))
}

func (d *DAL) DeregisterRunner(ctx context.Context, id RunnerID) error {
	return errors.WithStack(d.db.DeregisterRunner(ctx, int64(id)))
}

type Runner struct {
	ID       RunnerID
	Language string
	Endpoint string
	// Assigned deployment key, if any.
	Deployment types.Option[ulid.ULID]
}

func (d *DAL) GetIdleRunnersForLanguage(ctx context.Context, language string) ([]Runner, error) {
	runners, err := d.db.GetIdleRunnersForLanguage(ctx, language)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return slices.Map(runners, func(row sql.Runner) Runner {
		return Runner{
			ID:       RunnerID(row.ID),
			Language: row.Language,
			Endpoint: row.Endpoint,
		}
	}), nil
}

func (d *DAL) GetRunnersForModule(ctx context.Context, module string) ([]Runner, error) {
	runners, err := d.db.GetRunnersForModule(ctx, module)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return slices.Map(runners, func(row sql.GetRunnersForModuleRow) Runner {
		return Runner{
			ID:         RunnerID(row.ID),
			Language:   row.Language,
			Endpoint:   row.Endpoint,
			Deployment: types.Some(row.DeploymentKey.ULID()),
		}
	}), nil
}

func (d *DAL) InsertDeploymentLogEntry(ctx context.Context, deployment ulid.ULID, logEntry log.Entry) error {
	logError := pgtype.Text{}
	if logEntry.Error != nil {
		logError.String = logEntry.Error.Error()
		logError.Valid = true
	}
	return errors.WithStack(d.db.InsertDeploymentLogEntry(ctx, sql.InsertDeploymentLogEntryParams{
		Key:       sqltypes.Key(deployment),
		TimeStamp: pgtype.Timestamptz{Time: logEntry.Time, Valid: true},
		Level:     int32(logEntry.Level.Severity()),
		Scope:     strings.Join(logEntry.Scope, ":"),
		Message:   logEntry.Message,
		Error:     logError,
	}))
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
		return nil, errors.WithStack(err)
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

// func fileDigest(path string) (string, error) {
// 	r, err := os.Open(path)
// 	if err != nil {
// 		return "", errors.WithStack(err)
// 	}
// 	defer r.Close()

// 	h := sha256.New()
// 	_, err = io.Copy(h, r)
// 	if err != nil {
// 		return "", errors.WithStack(err)
// 	}
// 	return hex.EncodeToString(h.Sum(nil)), nil
// }

type artefactReader struct {
	id     int64
	db     *sql.DB
	offset int32
}

func (r *artefactReader) Read(p []byte) (n int, err error) {
	content, err := r.db.GetArtefactContentRange(context.Background(), r.offset+1, int32(len(p)), r.id)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	copy(p, content)
	clen := len(content)
	r.offset += int32(clen)
	if clen == 0 {
		err = io.EOF
	}
	return clen, err
}

// isPGConflict returns true if the error is the result of unique constraint conflict.
func isPGConflict(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return true
	}
	return false
}
