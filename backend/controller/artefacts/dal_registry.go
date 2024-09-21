package artefacts

import (
	"context"
	"fmt"
	"github.com/TBD54566975/ftl/backend/controller/artefacts/internal/sql"
	"io"

	sets "github.com/deckarep/golang-set/v2"

	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
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

type Service struct {
	*libdal.Handle[Service]
	db sql.Querier
}

func New(_ context.Context, conn libdal.Connection) *Service {
	return &Service{
		db: sql.New(conn),
		Handle: libdal.New(conn, func(h *libdal.Handle[Service]) *Service {
			return &Service{
				Handle: h,
				db:     sql.New(h.Connection),
			}
		}),
	}
}

func (s *Service) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	have, err := s.db.GetArtefactDigests(ctx, sha256esToBytes(digests))
	if err != nil {
		return nil, nil, libdal.TranslatePGError(err)
	}
	keys = slices.Map(have, func(in sql.GetArtefactDigestsRow) ArtefactKey {
		return ArtefactKey{ID: in.ID, Digest: sha256.FromBytes(in.Digest)}
	})
	haveStr := slices.Map(keys, func(in ArtefactKey) sha256.SHA256 {
		return in.Digest
	})
	missing = sets.NewSet(digests...).Difference(sets.NewSet(haveStr...)).ToSlice()
	return keys, missing, nil
}

func (s *Service) Upload(ctx context.Context, artefact Artefact) (sha256.SHA256, error) {
	sha256digest := sha256.Sum(artefact.Content)
	_, err := s.db.CreateArtefact(ctx, sha256digest[:], artefact.Content)
	return sha256digest, libdal.TranslatePGError(err)
}

func (s *Service) Download(ctx context.Context, digest sha256.SHA256) (io.ReadCloser, error) {
	digests := [][]byte{digest[:]}
	rows, err := s.db.GetArtefactDigests(ctx, digests)
	if err != nil {
		return nil, fmt.Errorf("unable to get artefact digests: %w", libdal.TranslatePGError(err))
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no artefact found with digest: %s", digest)
	}
	return &dalArtefactStream{db: s.db, id: rows[0].ID}, nil
}

type dalArtefactStream struct {
	db     sql.Querier
	id     int64
	offset int32
}

func (s *dalArtefactStream) Close() error { return nil }

func (s *dalArtefactStream) Read(p []byte) (n int, err error) {
	content, err := s.db.GetArtefactContentRange(context.Background(), s.offset+1, int32(len(p)), s.id)
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
