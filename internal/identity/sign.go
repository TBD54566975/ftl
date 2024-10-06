package identity

import (
	"bytes"
	"fmt"

	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/signature"
	"github.com/tink-crypto/tink-go/v2/tink"
)

type Signer struct {
	signer tink.Signer
}

func (k Signer) Sign(data []byte) (SignedData, error) {
	signatureBytes, err := k.signer.Sign(data)
	if err != nil {
		return SignedData{}, fmt.Errorf("failed to sign data: %w", err)
	}

	return SignedData{
		data:      data,
		Signature: NewSignature(signatureBytes),
	}, nil
}

func (k Signer) Public() (RawPublicKey, error) {
	panic("implement me")
}

type Verifier struct {
	verifier tink.Verifier
}

func NewVerifier(publicKey RawPublicKey) (Verifier, error) {
	reader := keyset.NewBinaryReader(bytes.NewReader(publicKey.Bytes))
	public, err := keyset.ReadWithNoSecrets(reader)
	if err != nil {
		return Verifier{}, fmt.Errorf("failed to read public keyset: %w", err)
	}

	verifier, err := signature.NewVerifier(public)
	if err != nil {
		return Verifier{}, fmt.Errorf("failed to create verifier: %w", err)
	}

	return Verifier{
		verifier: verifier,
	}, nil
}

func (k Verifier) Verify(signedData SignedData) ([]byte, error) {
	err := k.verifier.Verify(signedData.Signature.Bytes, signedData.data)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	return signedData.data, nil
}

type KeyPair struct {
	keysetHandle keyset.Handle
}

func (t KeyPair) Signer() (Signer, error) {
	signer, err := signature.NewSigner(&t.keysetHandle)
	if err != nil {
		return Signer{}, fmt.Errorf("failed to create signer: %w", err)
	}

	return Signer{
		signer: signer,
	}, nil
}

func (t KeyPair) Verifier() (Verifier, error) {
	public, err := t.keysetHandle.Public()
	if err != nil {
		return Verifier{}, fmt.Errorf("failed to get public keyset from keyset handle: %w", err)
	}

	verifier, err := signature.NewVerifier(public)
	if err != nil {
		return Verifier{}, fmt.Errorf("failed to create verifier: %w", err)
	}

	return Verifier{
		verifier: verifier,
	}, nil
}

func (t KeyPair) Public() (RawPublicKey, error) {
	// TODO: Maybe slow. Cache it.
	publicHandle, err := t.keysetHandle.Public()
	if err != nil {
		return RawPublicKey{}, fmt.Errorf("failed to get public keyset from keyset handle: %w", err)
	}

	buf := new(bytes.Buffer)
	writer := keyset.NewBinaryWriter(buf)
	if err := publicHandle.WriteWithNoSecrets(writer); err != nil {
		return RawPublicKey{}, fmt.Errorf("failed to write public keyset to buffer: %w", err)
	}

	publicKey := RawPublicKey{
		Bytes: buf.Bytes(),
	}
	return publicKey, nil
}

func (t KeyPair) Handle() keyset.Handle {
	return t.keysetHandle
}

// GenerateKeyPair creates a new key pair using Tink's ED25519 key template
func GenerateKeyPair() (KeyPair, error) {
	keysetHandle, err := keyset.NewHandle(signature.ED25519KeyTemplate())
	if err != nil {
		return KeyPair{}, fmt.Errorf("failed to create keyset handle: %w", err)
	}

	return KeyPair{
		keysetHandle: *keysetHandle,
	}, nil
}

func NewKeyPair(handle keyset.Handle) KeyPair {
	return KeyPair{
		keysetHandle: handle,
	}
}
