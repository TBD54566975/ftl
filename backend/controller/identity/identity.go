package identity

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"

	"github.com/tink-crypto/tink-go/v2/keyset"

	encryptionsvc "github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/identity/dal"
	"github.com/TBD54566975/ftl/backend/libdal"
	internalidentity "github.com/TBD54566975/ftl/internal/identity"
)

type Service struct {
	dal        dal.DAL
	encryption *encryptionsvc.Service
	signer     internalidentity.Signer
	verifier   internalidentity.Verifier
}

func New(ctx context.Context, encryption *encryptionsvc.Service, conn *sql.DB) (*Service, error) {
	svc := &Service{
		dal:        *dal.New(conn),
		encryption: encryption,
	}

	err := svc.ensureIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure identity: %w", err)
	}

	return svc, nil
}

const verificationText = "My voice is my passport, verify me."

func (s Service) ensureIdentity(ctx context.Context) error {
	tx, err := s.dal.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = s.dal.GetOnlyIdentityKey(ctx)
	if err != nil {
		if !libdal.IsNotFound(err) {
			return fmt.Errorf("failed to get only identity key: %w", err)
		}

		// Not found! Generate a new identity key
		pair, err := internalidentity.NewKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate key pair: %w", err)
		}

		signer, err := pair.Signer()
		if err != nil {
			return fmt.Errorf("failed to create signer: %w", err)
		}

		signed, err := signer.Sign([]byte(verificationText))
		if err != nil {
			return fmt.Errorf("failed to sign verification: %w", err)
		}

		verifier, err := pair.Verifier()
		if err != nil {
			return fmt.Errorf("failed to create verifier: %w", err)
		}

		// For total sanity, verify immediately
		if err = verifier.Verify(*signed); err != nil {
			return fmt.Errorf("failed to verify signed verification: %w", err)
		}

		// This is what i need to do via interfaces/generics
		handle := pair.Handle()
		buf := new(bytes.Buffer)
		writer := keyset.NewBinaryWriter(buf)
		aead := s.encryption.AEAD()
		handle.Write(writer, aead)
		var encryptedIdentityColumn api.EncryptedIdentityColumn
		s.encryption.EncryptKeyPair(pair, &encryptedIdentityColumn)

		fmt.Println("buf: ", buf.Bytes())
		panic("stop")

		// var encryptedIdentity api.EncryptedIdentityColumn
		// encrypted, err := s.encryption.Encrypt()
	}

	return nil
}
