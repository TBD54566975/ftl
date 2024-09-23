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
	"github.com/TBD54566975/ftl/internal/log"
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

	keyPair, err := svc.getKeyPair(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get key pair: %w", err)
	}

	signer, err := keyPair.Signer()
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}
	svc.signer = signer

	verifier, err := keyPair.Verifier()
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}
	svc.verifier = verifier

	return svc, nil
}

func (s Service) Sign(data []byte) (*internalidentity.SignedData, error) {
	signedData, err := s.signer.Sign(data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return signedData, nil
}

func (s Service) Verify(signedData internalidentity.SignedData) error {
	err := s.verifier.Verify(signedData)
	if err != nil {
		return fmt.Errorf("failed to verify data: %w", err)
	}

	return nil
}

func (s Service) getKeyPair(ctx context.Context) (internalidentity.KeyPair, error) {
	identity, err := s.dal.GetOnlyIdentityKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get only identity key: %w", err)
	}

	reader := keyset.NewBinaryReader(bytes.NewReader(identity.Private.Bytes()))
	aead, err := s.encryption.AEAD()
	if err != nil {
		return nil, fmt.Errorf("failed to get AEAD: %w", err)
	}

	handle, err := keyset.Read(reader, aead)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyset: %w", err)
	}

	keyPair := internalidentity.NewTinkKeyPair(*handle)
	return keyPair, nil
}

const verificationText = "My voice is my passport, verify me."

func (s Service) ensureIdentity(ctx context.Context) (err error) {
	logger := log.FromContext(ctx)
	tx, err := s.dal.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	_, err = s.dal.GetOnlyIdentityKey(ctx)
	if err != nil {
		if !libdal.IsNotFound(err) {
			return fmt.Errorf("failed to get only identity key: %w", err)
		}

		logger.Debugf("Generating identity key")
		err = s.generateAndSaveIdentity(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to generate and save identity: %w", err)
		}
	} else {
		logger.Debugf("Identity key already exists")
	}

	return nil
}

func (s Service) generateAndSaveIdentity(ctx context.Context, tx *dal.DAL) error {
	pair, err := internalidentity.GenerateTinkKeyPair()
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

	// TODO: Make this support different encryptors.
	// Might need to refactor internal/identity to access controller encryption types.
	// It's a bit tricky because you can't take out the private key from the keyset without
	// encrypting it with the AEAD.
	handle := pair.Handle()
	buf := new(bytes.Buffer)
	writer := keyset.NewBinaryWriter(buf)
	aead, err := s.encryption.AEAD()
	if err != nil {
		return fmt.Errorf("failed to get AEAD: %w", err)
	}
	if err := handle.Write(writer, aead); err != nil {
		return fmt.Errorf("failed to write keyset: %w", err)
	}
	encryptedIdentityColumn := api.EncryptedIdentityColumn{}
	encryptedIdentityColumn.Set(buf.Bytes())

	public, err := pair.Public()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	encryptedIdentity := &dal.EncryptedIdentity{
		Private:         encryptedIdentityColumn,
		Public:          public,
		VerifySignature: signed.Signature,
	}
	if err := tx.CreateOnlyIdentityKey(ctx, *encryptedIdentity); err != nil {
		return fmt.Errorf("failed to create only identity key: %w", err)
	}

	return nil
}
