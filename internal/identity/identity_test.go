package identity

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBasics(t *testing.T) {
	keyPair, err := GenerateTinkKeyPair()
	assert.NoError(t, err)

	signer, err := keyPair.Signer()
	assert.NoError(t, err)

	data := []byte("hunter2")
	signedData, err := signer.Sign(data)
	assert.NoError(t, err)

	verifier, err := keyPair.Verifier()
	assert.NoError(t, err)

	data, err = verifier.Verify(*signedData)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", string(data))

	// Now fail it just for sanity
	signedData.signature[0] = ^signedData.signature[0]
	_, err = verifier.Verify(*signedData)
	assert.EqualError(t, err, "failed to verify signature: verifier_factory: invalid signature")
}

func TestCertificate(t *testing.T) {
	// Set up CA
	caStore, err := NewStore()
	assert.NoError(t, err)
	caSigner, err := caStore.KeyPair.Signer()
	assert.NoError(t, err)
	caPublicKey, err := caStore.KeyPair.Public()
	assert.NoError(t, err)

	// Runner generates a key pair and identity for signing
	runnerStore, err := NewStore()
	assert.NoError(t, err)
	runnerSigner, err := runnerStore.KeyPair.Signer()
	assert.NoError(t, err)
	runnerIdentity, err := Parse("r:rnr-1234:echo")
	assert.NoError(t, err)
	runnerSignedData, err := Sign(runnerSigner, runnerIdentity)
	fmt.Printf("Runner signed data: %s\n", runnerSignedData)
	assert.NoError(t, err)
	runnerPublicKey, err := runnerStore.KeyPair.Public()
	assert.NoError(t, err)

	// Hand wave "check the ID and module"

	// Sign the certificate
	certificate, err := SignCertificateRequest(caSigner, runnerPublicKey, *runnerSignedData)
	assert.NoError(t, err)

	fmt.Printf("Certificate: %s\n", certificate)

	// Runner A constructs a message for runner B
	message := []byte("hello")
	signedMessage, err := runnerSigner.Sign(message)
	certified := CertifiedSignedData{
		Certificate: *certificate,
		SignedData:  *signedMessage,
	}

	// Hand wave "send the certificate to the runner"

	// Runner B verifies the message
	caVerifier, err := NewTinkVerifier(caPublicKey)
	assert.NoError(t, err)

	data, err := certified.Verify(caVerifier)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}
