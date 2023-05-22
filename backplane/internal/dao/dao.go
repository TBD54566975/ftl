// Package dao provides a data access object for the backplane.
package dao

import (
	"context"
	"io"
	"strings"

	"github.com/alecthomas/errors"
	sets "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/backplane/internal/sql"
	"github.com/TBD54566975/ftl/common/maps"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/common/slices"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
)

type DAO struct {
	db *sql.DB
}

func New(conn sql.DBI) *DAO {
	return &DAO{db: sql.NewDB(conn)}
}

func (d *DAO) CreateModule(ctx context.Context, language, name string) (err error) {
	_, err = d.db.CreateModule(ctx, sql.CreateModuleParams{
		Language: language,
		Name:     name,
	})
	return errors.WithStack(err)
}

// GetMissingArtefacts returns the digests of the artefacts that are missing from the database.
func (d *DAO) GetMissingArtefacts(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error) {
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
func (d *DAO) CreateArtefact(ctx context.Context, content []byte) (digest sha256.SHA256, err error) {
	sha256digest := sha256.Sum(content)
	_, err = d.db.CreateArtefact(ctx, sql.CreateArtefactParams{
		Digest:  sha256digest[:],
		Content: content,
	})
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
func (d *DAO) CreateDeployment(
	ctx context.Context,
	language string,
	schema *schema.Module,
	artefacts []DeploymentArtefact,
) (key uuid.UUID, err error) {
	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, errors.WithStack(err)
	}
	defer tx.CommitOrRollback(ctx, &err)()

	// Check if the deployment already exists and if so, return it.
	existing, err := tx.GetDeploymentsWithArtefacts(ctx, sql.GetDeploymentsWithArtefactsParams{
		Count:   len(artefacts),
		Digests: sha256esToBytes(slices.Map(artefacts, func(in DeploymentArtefact) sha256.SHA256 { return in.Digest })),
	})
	if err != nil {
		return uuid.UUID{}, errors.WithStack(err)
	}
	if len(existing) > 0 {
		return existing[0].Key, nil
	}

	artefactsByDigest := maps.FromSlice(artefacts, func(in DeploymentArtefact) (sha256.SHA256, DeploymentArtefact) {
		return in.Digest, in
	})

	schemaBytes, err := proto.Marshal(schema.ToProto())
	if err != nil {
		return uuid.UUID{}, errors.WithStack(err)
	}

	_, err = tx.CreateModule(ctx, sql.CreateModuleParams{
		Language: language, // TODO(aat): "schema" containing language?
		Name:     schema.Name,
	})

	// Create the deployment
	deploymentKey, err := tx.CreateDeployment(ctx, sql.CreateDeploymentParams{
		ModuleName: schema.Name,
		Schema:     schemaBytes,
	})
	if err != nil {
		return uuid.UUID{}, errors.WithStack(err)
	}

	uploadedDigests := slices.Map(artefacts, func(in DeploymentArtefact) []byte { return in.Digest[:] })
	artefactDigests, err := tx.GetArtefactDigests(ctx, uploadedDigests)
	if err != nil {
		return uuid.UUID{}, errors.WithStack(err)
	}
	if len(artefactDigests) != len(artefacts) {
		missingDigests := strings.Join(slices.Map(artefacts, func(in DeploymentArtefact) string { return in.Digest.String() }), ", ")
		return uuid.UUID{}, errors.Errorf("missing %d artefacts: %s", len(artefacts)-len(artefactDigests), missingDigests)
	}

	// Associate the artefacts with the deployment
	for _, row := range artefactDigests {
		artefact := artefactsByDigest[sha256.FromBytes(row.Digest)]
		err = tx.AssociateArtefactWithDeployment(ctx, sql.AssociateArtefactWithDeploymentParams{
			Key:        deploymentKey,
			ArtefactID: row.ID,
			Executable: artefact.Executable,
			Path:       artefact.Path,
		})
		if err != nil {
			return uuid.UUID{}, errors.WithStack(err)
		}
	}

	return deploymentKey, nil
}

type Deployment struct {
	Module    string
	Language  string
	Key       uuid.UUID
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

func (d *DAO) GetDeployment(ctx context.Context, id uuid.UUID) (*Deployment, error) {
	deployment, err := d.db.GetDeployment(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return d.loadDeployment(ctx, sql.GetLatestDeploymentRow(deployment))
}

func (d *DAO) GetLatestDeployment(ctx context.Context, module string) (*Deployment, error) {
	deployment, err := d.db.GetLatestDeployment(ctx, module)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return d.loadDeployment(ctx, deployment)
}

func (d *DAO) loadDeployment(ctx context.Context, deployment sql.GetLatestDeploymentRow) (*Deployment, error) {
	pm := &pschema.Module{}
	err := proto.Unmarshal(deployment.Schema, pm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	out := &Deployment{
		Module:   deployment.ModuleName,
		Language: deployment.Language,
		Key:      deployment.Key,
		Schema:   schema.ModuleFromProto(pm),
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
	content, err := r.db.GetArtefactContentRange(context.Background(), sql.GetArtefactContentRangeParams{
		Start: r.offset + 1,
		Count: int32(len(p)),
		ID:    r.id,
	})
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
