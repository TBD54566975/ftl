package encryption

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestNoOpEncryptor(t *testing.T) {
	encryptor := NoOpEncryptorNext{}

	encrypted, err := encryptor.Encrypt(TimelineSubKey, []byte("hunter2"))
	assert.NoError(t, err)

	decrypted, err := encryptor.Decrypt(TimelineSubKey, encrypted)
	assert.NoError(t, err)

	assert.Equal(t, "hunter2", string(decrypted))
}

// echo -n "fake-kms://" && tinkey create-keyset --key-template AES128_GCM --out-format binary | base64 | tr '+/' '-_' | tr -d '='
func TestKMSEncryptorFakeKMS(t *testing.T) {
	uri := "fake-kms://CKbvh_ILElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEE6tD2yE5AWYOirhmkY-r3sYARABGKbvh_ILIAE"

	encryptor, err := NewKMSEncryptorGenerateKey(uri, nil)
	assert.NoError(t, err)

	encrypted, err := encryptor.Encrypt(TimelineSubKey, []byte("hunter2"))
	assert.NoError(t, err)

	decrypted, err := encryptor.Decrypt(TimelineSubKey, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", string(decrypted))

	// Should fail to decrypt with the wrong subkey
	_, err = encryptor.Decrypt(AsyncSubKey, encrypted)
	assert.Error(t, err)
}
