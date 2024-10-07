package identity

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/model"
)

func TestBasics(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	assert.NoError(t, err)

	signer, err := keyPair.Signer()
	assert.NoError(t, err)

	data := []byte("hunter2")
	signedData, err := signer.Sign(data)
	assert.NoError(t, err)

	verifier, err := keyPair.Verifier()
	assert.NoError(t, err)

	data, err = verifier.Verify(signedData)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", string(data))

	// Now fail it just for sanity
	signedData.Signature.Bytes[0] = ^signedData.Signature.Bytes[0]
	_, err = verifier.Verify(signedData)
	assert.EqualError(t, err, "failed to verify signature: verifier_factory: invalid signature")
}

func TestCertificate(t *testing.T) {
	// Set up CA
	caIdent := NewController()
	caStore, err := NewStoreNewKeys(caIdent)
	assert.NoError(t, err)
	caVerifier, err := caStore.KeyPair.Verifier()
	assert.NoError(t, err)

	// Runner generates a key pair and identity for signing
	runnerKey := model.NewRunnerKey("hostname", "1234")
	runnerIdent := NewRunner(runnerKey, "echo")
	assert.NoError(t, err)
	runnerStore, err := NewStoreNewKeys(runnerIdent)
	assert.NoError(t, err)
	request, err := runnerStore.NewGetCertificateRequest()
	assert.NoError(t, err)

	// Hand wave "send the request to the CA"
	// Hand wave "check the ID and module"

	certificate, err := caStore.SignCertificateRequest(&request)
	assert.NoError(t, err)

	// Hand wave "send the certificate to the runner"

	err = runnerStore.SetCertificate(certificate, caVerifier)
	assert.NoError(t, err)

	// Runner constructs a certified message
	message := []byte("hello")
	certified, err := runnerStore.CertifiedSign(message)
	assert.NoError(t, err)

	id, data, err := certified.Verify(caVerifier)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(data))
	assert.Equal(t, runnerIdent.String(), id.String())
}
