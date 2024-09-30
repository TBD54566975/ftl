package identity

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
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
	signedData.Signature[0] = ^signedData.Signature[0]
	_, err = verifier.Verify(signedData)
	assert.EqualError(t, err, "failed to verify signature: verifier_factory: invalid signature")
}

func TestCertificate(t *testing.T) {
	// Set up CA
	caIdent := NewIdentity("ca", "")
	caStore, err := NewStoreNewKeys(caIdent)
	assert.NoError(t, err)
	caVerifier, err := caStore.KeyPair.Verifier()
	assert.NoError(t, err)

	// Runner generates a key pair and identity for signing
	runnerIdent := NewIdentity("runner", "echo")
	assert.NoError(t, err)
	runnerStore, err := NewStoreNewKeys(runnerIdent)
	assert.NoError(t, err)
	request, err := runnerStore.NewGetCertificateRequest()
	assert.NoError(t, err)
	fmt.Printf("runner request id %s\n", request.Request.Identity)
	fmt.Printf("runner request public key %x\n", request.Request.PublicKey)
	fmt.Printf("runner request signature %x\n", request.Signature)

	// Hand wave "send the request to the CA"
	// Hand wave "check the ID and module"

	certificate, err := caStore.SignCertificateRequest(&request)
	assert.NoError(t, err)

	// Hand wave "send the certificate to the runner"

	err = runnerStore.SetCertificate(certificate, caVerifier)
	assert.NoError(t, err)

	// Runner A constructs a certified message
	message := []byte("hello")
	certified, err := runnerStore.CertifiedSign(message)

	fmt.Printf("Certified message: %s\n", certified)
	_, data, err := certified.Verify(caVerifier)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}
