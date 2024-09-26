package identity

import (
	"bytes"
	"fmt"

	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/signature"
	"github.com/tink-crypto/tink-go/v2/tink"
)

var _ Signer = &TinkSigner{}

type TinkSigner struct {
	signer tink.Signer
}

func (k TinkSigner) Sign(data []byte) (SignedData, error) {
	bytes, err := k.signer.Sign(data)
	if err != nil {
		return SignedData{}, fmt.Errorf("failed to sign data: %w", err)
	}

	return SignedData{
		data:      data,
		Signature: bytes,
	}, nil
}

func (k TinkSigner) Public() ([]byte, error) {
	panic("implement me")
}

var _ Verifier = &TinkVerifier{}

type TinkVerifier struct {
	verifier tink.Verifier
}

func NewTinkVerifier(publicKey []byte) (Verifier, error) {
	fmt.Printf("publicKey: %s\n", string(publicKey))
	reader := keyset.NewBinaryReader(bytes.NewReader(publicKey))
	public, err := keyset.ReadWithNoSecrets(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read public keyset: %w", err)
	}

	verifier, err := signature.NewVerifier(public)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}

	return &TinkVerifier{
		verifier: verifier,
	}, nil
}

func (k TinkVerifier) Verify(signedData SignedData) ([]byte, error) {
	err := k.verifier.Verify(signedData.Signature, signedData.data)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	return signedData.data, nil
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

func (t TinkKeyPair) Public() ([]byte, error) {
	publicHandle, err := t.keysetHandle.Public()
	if err != nil {
		return nil, fmt.Errorf("failed to get public keyset from keyset handle: %w", err)
	}

	buf := new(bytes.Buffer)
	writer := keyset.NewBinaryWriter(buf)
	if err := publicHandle.WriteWithNoSecrets(writer); err != nil {
		return nil, fmt.Errorf("failed to write public keyset to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

// Handle returns the keyset handle.
// TODO: Remove this. We don't want to expose the private key.
func (t TinkKeyPair) Handle() keyset.Handle {
	return t.keysetHandle
}

// GenerateTinkKeyPair creates a new key pair using Tink's ED25519 key template
func GenerateTinkKeyPair() (*TinkKeyPair, error) {
	keysetHandle, err := keyset.NewHandle(signature.ED25519KeyTemplate())
	if err != nil {
		return nil, fmt.Errorf("failed to create keyset handle: %w", err)
	}

	return &TinkKeyPair{
		keysetHandle: *keysetHandle,
	}, nil
}

func NewTinkKeyPair(handle keyset.Handle) *TinkKeyPair {
	return &TinkKeyPair{
		keysetHandle: handle,
	}
}
