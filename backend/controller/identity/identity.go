package identity

import (
	"context"
	"database/sql"

	encryptionsvc "github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/identity/dal"
)

type KeyStoreProvider interface {
	// EnsureKey asks a provider to check for an identity key.
	// If not available, call the generateKey function to create a new key.
	// The provider should handle transactions around checking and setting the key, to prevent race conditions.
	EnsureKey(ctx context.Context, generateKey func() ([]byte, error)) ([]byte, error)
}

type Service struct {
	dal        dal.DAL
	encryption *encryptionsvc.Service
}

func New(ctx context.Context, encryption *encryptionsvc.Service, conn *sql.DB) *Service {
	svc := &Service{
		dal:        *dal.New(conn),
		encryption: encryption,
	}
	return svc
}

// func (s *Service) NewCronJobsForModule(ctx context.Context, module *schemapb.Module) ([]model.CronJob, error) {
