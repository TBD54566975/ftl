package identity

import (
	"fmt"

	// "github.com/aws/aws-sdk-go/service/kms"

	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/signature"
	"github.com/tink-crypto/tink-go/v2/tink"
)

var _ Signer = &TinkSigner{}

type TinkSigner struct {
	signer tink.Signer
}

func (k TinkSigner) Sign(data []byte) (*SignedData, error) {
	bytes, err := k.signer.Sign(data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return &SignedData{
		Data:      data,
		Signature: bytes,
	}, nil
}

var _ Verifier = &TinkVerifier{}

type TinkVerifier struct {
	verifier tink.Verifier
}

func (k TinkVerifier) Verify(signedData SignedData) error {
	err := k.verifier.Verify(signedData.Signature, signedData.Data)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	return nil
}

var _ KeyPair = &TinkKeyPair{}

type TinkKeyPair struct {
	keysetHandle keyset.Handle
}

func (t TinkKeyPair) Signer() (Signer, error) {
	signer, err := signature.NewSigner(&t.keysetHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	return TinkSigner{
		signer: signer,
	}, nil
}

func (t TinkKeyPair) Verifier() (Verifier, error) {
	public, err := t.keysetHandle.Public()
	if err != nil {
		return nil, fmt.Errorf("failed to get public keyset from keyset handle: %w", err)
	}

	verifier, err := signature.NewVerifier(public)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}

	return &TinkVerifier{
		verifier: verifier,
	}, nil
}

// TODO: Remove this
func (t TinkKeyPair) Handle() keyset.Handle {
	return t.keysetHandle
}

// NewKeyPair creates a new key pair using Tink's ED25519 key template
func NewKeyPair() (*TinkKeyPair, error) {
	keysetHandle, err := keyset.NewHandle(signature.ED25519KeyTemplate())
	if err != nil {
		return nil, fmt.Errorf("failed to create keyset handle: %w", err)
	}

	return &TinkKeyPair{
		keysetHandle: *keysetHandle,
	}, nil
}
