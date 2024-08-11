package encryption

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

const streamingKey = `{
    "primaryKeyId": 1720777699,
    "key": [{
        "keyData": {
            "typeUrl": "type.googleapis.com/google.crypto.tink.AesCtrHmacStreamingKey",
            "keyMaterialType": "SYMMETRIC",
            "value": "Eg0IgCAQIBgDIgQIAxAgGiDtesd/4gCnQdTrh+AXodwpm2b6BFJkp043n+8mqx0YGw=="
        },
        "outputPrefixType": "RAW",
        "keyId": 1720777699,
        "status": "ENABLED"
    }]
}`

func TestDeprecatedNewEncryptor(t *testing.T) {
	jsonInput := "\"hello\""

	encryptor, err := NewForKeyOrURI(streamingKey)
	assert.NoError(t, err)

	encrypted, err := encryptor.EncryptJSON(jsonInput)
	assert.NoError(t, err)
	fmt.Printf("Encrypted: %s\n", encrypted)

	var decrypted json.RawMessage
	err = encryptor.DecryptJSON(encrypted, &decrypted)
	assert.NoError(t, err)
	fmt.Printf("Decrypted: %s\n", decrypted)

	var decryptedString string
	err = json.Unmarshal(decrypted, &decryptedString)
	assert.NoError(t, err)
	fmt.Printf("Decrypted string: %s\n", decryptedString)

	assert.Equal(t, jsonInput, decryptedString)
}

// tinkey create-keyset --key-template HKDF_SHA256_DERIVES_AES256_GCM
const key = `{"primaryKeyId":2304101620,"key":[{"keyData":{"typeUrl":"type.googleapis.com/google.crypto.tink.PrfBasedDeriverKey","value":"El0KMXR5cGUuZ29vZ2xlYXBpcy5jb20vZ29vZ2xlLmNyeXB0by50aW5rLkhrZGZQcmZLZXkSJhICCAMaIDnEx9gPgeF32LQYjFYNSZe8b9KUl41Xy6to8MqKcSjBGAEaOgo4CjB0eXBlLmdvb2dsZWFwaXMuY29tL2dvb2dsZS5jcnlwdG8udGluay5BZXNHY21LZXkSAhAgGAE=","keyMaterialType":"SYMMETRIC"},"status":"ENABLED","keyId":2304101620,"outputPrefixType":"TINK"}]}`

func TestPlaintextEncryptor(t *testing.T) {
	encryptor, err := NewPlaintextEncryptor(key)
	assert.NoError(t, err)

	encrypted, err := encryptor.Encrypt(Logs, []byte("hunter2"))
	assert.NoError(t, err)
	fmt.Printf("Encrypted: %s\n", encrypted)

	decrypted, err := encryptor.Decrypt(Logs, encrypted)
	assert.NoError(t, err)
	fmt.Printf("Decrypted: %s\n", decrypted)

	assert.Equal(t, "hunter2", string(decrypted))

	// Should fail to decrypt with the wrong subkey
	_, err = encryptor.Decrypt(Async, encrypted)
	assert.Error(t, err)

}

func TestNoOpEncryptor(t *testing.T) {
	encryptor := NoOpEncryptorNext{}

	encrypted, err := encryptor.Encrypt(Logs, []byte("hunter2"))
	assert.NoError(t, err)

	decrypted, err := encryptor.Decrypt(Logs, encrypted)
	assert.NoError(t, err)

	assert.Equal(t, "hunter2", string(decrypted))
}

func TestKmsEncryptorFakeKMS(t *testing.T) {
	uri := "fake-kms://CM2b3_MDElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEIK75t5L-adlUwVhWvRuWUwYARABGM2b3_MDIAE"

	encryptor, err := NewKmsEncryptorGenerateKey(uri, nil)
	assert.NoError(t, err)

	encrypted, err := encryptor.Encrypt(Logs, []byte("hunter2"))
	assert.NoError(t, err)

	decrypted, err := encryptor.Decrypt(Logs, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", string(decrypted))

	// Should fail to decrypt with the wrong subkey
	_, err = encryptor.Decrypt(Async, encrypted)
	assert.Error(t, err)
}
