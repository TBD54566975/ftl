package identity

import (
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

	err = verifier.Verify(*signedData)
	assert.NoError(t, err)

	// Now fail it just for sanity
	signedData.Signature[0] = ^signedData.Signature[0]
	err = verifier.Verify(*signedData)
	assert.EqualError(t, err, "failed to verify signature: verifier_factory: invalid signature")
}
