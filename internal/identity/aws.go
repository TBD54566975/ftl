package identity

import (
	"fmt"

	// "github.com/aws/aws-sdk-go/service/kms"

	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/signature"
	"github.com/tink-crypto/tink-go/v2/tink"
)

var _ Signer = &KMSIdentitySigner{}

type KMSIdentitySigner struct {
	signer tink.Signer
}

func (k KMSIdentitySigner) Sign(data []byte) (*SignedData, error) {
	bytes, err := k.signer.Sign(data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return &SignedData{
		Data:      data,
		Signature: bytes,
	}, nil
}

var _ Verifier = &KMSIdentityVerifier{}

type KMSIdentityVerifier struct {
	verifier tink.Verifier
}

func (k KMSIdentityVerifier) Verify(signedData SignedData) error {
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

	return KMSIdentitySigner{
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

	return &KMSIdentityVerifier{
		verifier: verifier,
	}, nil
}

// TODO encrypt with KMS
// For now dump the key into the db as plaintext
func newKeyPair() (*TinkKeyPair, error) {
	keysetHandle, err := keyset.NewHandle(signature.ED25519KeyTemplate())
	if err != nil {
		return nil, fmt.Errorf("failed to create keyset handle: %w", err)
	}

	return &TinkKeyPair{
		keysetHandle: *keysetHandle,
	}, nil
}
