package encryption

import (
	"bytes"
	"github.com/tink-crypto/tink-go/v2/aead"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestNoOpEncryptor(t *testing.T) {
	encryptor := NoOpEncryptorNext{}

	encrypted, err := encryptor.Encrypt(LogsSubKey, []byte("hunter2"))
	assert.NoError(t, err)

	decrypted, err := encryptor.Decrypt(LogsSubKey, encrypted)
	assert.NoError(t, err)

	assert.Equal(t, "hunter2", string(decrypted))
}

// echo -n "fake-kms://" && tinkey create-keyset --key-template AES128_GCM --out-format binary | base64 | tr '+/' '-_' | tr -d '='
func TestKMSEncryptorFakeKMS(t *testing.T) {
	uri := "fake-kms://CKbvh_ILElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEE6tD2yE5AWYOirhmkY-r3sYARABGKbvh_ILIAE"

	encryptor, err := NewKMSEncryptorGenerateKey(uri, nil)
	assert.NoError(t, err)

	encrypted, err := encryptor.Encrypt(LogsSubKey, []byte("hunter2"))
	assert.NoError(t, err)

	decrypted, err := encryptor.Decrypt(LogsSubKey, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", string(decrypted))

	// Should fail to decrypt with the wrong subkey
	_, err = encryptor.Decrypt(AsyncSubKey, encrypted)
	assert.Error(t, err)
}

func BenchmarkDeriveAndEncrypt(b *testing.B) {
	uri := "fake-kms://CKbvh_ILElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEE6tD2yE5AWYOirhmkY-r3sYARABGKbvh_ILIAE"
	clearText := []byte(uri)
	encryptor, err := NewKMSEncryptorGenerateKey(uri, nil)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		encrypted, err := encryptor.Encrypt(LogsSubKey, clearText)
		if err != nil {
			b.Fatal(err)
		}

		decrypted, err := encryptor.Decrypt(LogsSubKey, encrypted)
		if err != nil {
			b.Fatal(err)
		}

		if !bytes.Equal(clearText, decrypted) {
			b.Fatal("decrypted text does not match clear text")
		}
	}
}

func BenchmarkCacheDeriveAndEncrypt(b *testing.B) {
	uri := "fake-kms://CKbvh_ILElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEE6tD2yE5AWYOirhmkY-r3sYARABGKbvh_ILIAE"
	clearText := []byte(uri)
	encryptor, err := NewKMSEncryptorGenerateKey(uri, nil)
	if err != nil {
		b.Fatal(err)
	}

	derived, err := deriveKeyset(encryptor.root, []byte(LogsSubKey))
	if err != nil {
		b.Fatal(err)
	}

	primitive, err := aead.New(derived)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		encrypted, err := primitive.Encrypt(clearText, nil)
		if err != nil {
			b.Fatal(err)
		}

		decrypted, err := primitive.Decrypt(encrypted, nil)
		if err != nil {
			b.Fatal(err)
		}

		if !bytes.Equal(clearText, decrypted) {
			b.Fatal("decrypted text does not match clear text")
		}
	}
}
