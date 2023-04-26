package dao

import (
	"context"
	"crypto/sha256"

	"github.com/alecthomas/errors"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/backplane/internal/sql"
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

// func (d *DAO) GetArtefectDiff(ctx context.Context, haveDigests []string) (needDigests []string, err error) {
// 	needDigests, err := d.db.GetArtefactDigests(ctx, haveDigests)
// 	return needDigests, errors.WithStack(err)
// }

// CreateArtefact inserts a new artefact into the database and returns its ID.
func (d *DAO) CreateArtefact(ctx context.Context, path string, executable bool, content []byte) (id int64, err error) {
	sha256digest := sha256.Sum256(content)
	id, err = d.db.CreateArtefact(ctx, sql.CreateArtefactParams{
		Executable: executable,
		Path:       path,
		Digest:     sha256digest[:],
		Content:    content,
	})
	return id, errors.WithStack(err)
}

// CreateDeployment creates a new deployment and associates previously created artefacts with it.
func (d *DAO) CreateDeployment(
	ctx context.Context,
	module string,
	schema *schema.Module,
	artefacts []int64,
) (err error) {
	// Start the transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.CommitOrRollback(ctx, &err)()

	schemaBytes, err := proto.Marshal(schema.ToProto())
	if err != nil {
		return errors.WithStack(err)
	}

	// Create the deployment
	deploymentID, err := tx.CreateDeployment(ctx, sql.CreateDeploymentParams{
		ModuleName: module,
		Schema:     schemaBytes,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	// Associate the artefacts with the deployment
	for _, artefactID := range artefacts {
		err = tx.AssociateArtefactWithDeployment(ctx, sql.AssociateArtefactWithDeploymentParams{
			DeploymentID: deploymentID,
			ArtefactID:   artefactID,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
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
	Content    []byte
}

func (d *DAO) GetLatestDeployment(ctx context.Context, module string) (*Deployment, error) {
	deployment, err := d.db.GetLatestDeployment(ctx, module)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pm := &pschema.Module{}
	err = proto.Unmarshal(deployment.Schema, pm)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	out := &Deployment{
		Module:   deployment.ModuleName,
		Language: deployment.Language,
		Key:      deployment.Key,
		Schema:   schema.ProtoToModule(pm),
	}
	artefacts, err := d.db.GetArtefactsForDeployment(ctx, deployment.ID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, artefact := range artefacts {
		out.Artefacts = append(out.Artefacts, &Artefact{
			Path:       artefact.Path,
			Executable: artefact.Executable,
			Content:    artefact.Content,
		})
	}
	return out, nil
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
