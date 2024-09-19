package artefacts

import (
	"context"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
	sets "github.com/deckarep/golang-set/v2"
	"io"
)

type ArtefactRow struct {
	ID     int64
	Digest []byte
}

type DAL interface {
	GetArtefactDigests(ctx context.Context, digests [][]byte) ([]ArtefactRow, error)
	CreateArtefact(ctx context.Context, digest []byte, content []byte) (int64, error)
	GetArtefactContentRange(ctx context.Context, start int32, count int32, iD int64) ([]byte, error)
}

type DALRegistry struct {
	db DAL
}

func NewDALRegistry(db DAL) *DALRegistry {
	return &DALRegistry{db: db}
}

func MakeKey(id int64, digest []byte) ArtefactKey {
	return ArtefactKey{id: id, Digest: sha256.FromBytes(digest)}
}

func (r *DALRegistry) GetMissingDigests(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error) {
	have, err := r.db.GetArtefactDigests(ctx, sha256esToBytes(digests))
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	haveStr := slices.Map(have, func(in ArtefactRow) sha256.SHA256 {
		return sha256.FromBytes(in.Digest)
	})
	return sets.NewSet(digests...).Difference(sets.NewSet(haveStr...)).ToSlice(), nil
}

func (r *DALRegistry) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	sha256digest := sha256.Sum(artefact.Content)
	_, err := r.db.CreateArtefact(ctx, sha256digest[:], artefact.Content)
	return sha256digest, libdal.TranslatePGError(err)
}

func (r *DALRegistry) Download(_ context.Context, key ArtefactKey) io.ReadCloser {
	return &dalArtefactStream{dal: r.db, id: key.id}
}

type dalArtefactStream struct {
	dal    DAL
	id     int64
	offset int32
}

func (s *dalArtefactStream) Close() error { return nil }

func (s *dalArtefactStream) Read(p []byte) (n int, err error) {
	content, err := s.dal.GetArtefactContentRange(context.Background(), s.offset+1, int32(len(p)), s.id)
	if err != nil {
		return 0, libdal.TranslatePGError(err)
	}
	copy(p, content)
	clen := len(content)
	s.offset += int32(clen)
	if clen == 0 {
		err = io.EOF
	}
	return clen, err
}

func sha256esToBytes(digests []sha256.SHA256) [][]byte {
	return slices.Map(digests, func(digest sha256.SHA256) []byte { return digest[:] })
}
