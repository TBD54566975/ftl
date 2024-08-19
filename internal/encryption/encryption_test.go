package encryption

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestNoOpEncryptor(t *testing.T) {
	encryptor := NoOpEncryptor{}

	var encrypted EncryptedTimelineColumn
	err := encryptor.Encrypt([]byte("hunter2"), &encrypted)
	assert.NoError(t, err)

	decryptedLogs, err := encryptor.Decrypt(&encrypted)
	assert.NoError(t, err)

	assert.Equal(t, "hunter2", string(decryptedLogs))
}

// echo -n "fake-kms://" && tinkey create-keyset --key-template AES128_GCM --out-format binary | base64 | tr '+/' '-_' | tr -d '='
func TestKMSEncryptorFakeKMS(t *testing.T) {
	uri := "fake-kms://CKbvh_ILElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEE6tD2yE5AWYOirhmkY-r3sYARABGKbvh_ILIAE"

	key, err := newKey(uri, nil)
	assert.NoError(t, err)

	encryptor, err := NewKMSEncryptorWithKMS(uri, nil, key)
	assert.NoError(t, err)

	var encrypted EncryptedTimelineColumn
	err = encryptor.Encrypt([]byte("hunter2"), &encrypted)
	assert.NoError(t, err)

	decrypted, err := encryptor.Decrypt(&encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", string(decrypted))

	wrongSubKey := EncryptedAsyncColumn(encrypted)
	// Should fail to decrypt with the wrong subkey
	_, err = encryptor.Decrypt(&wrongSubKey)
	assert.Error(t, err)
}
