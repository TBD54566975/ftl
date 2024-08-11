package encryption

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/testutils"
	"github.com/alecthomas/assert/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awsv1 "github.com/aws/aws-sdk-go/aws"
	awsv1credentials "github.com/aws/aws-sdk-go/aws/credentials"
	awsv1session "github.com/aws/aws-sdk-go/aws/session"
	awsv1kms "github.com/aws/aws-sdk-go/service/kms"
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

	encryptor, err := NewDeprecatedKeyKeyOrURI(streamingKey)
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
	encryptor := NoOpEncryptor{}

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

func TestKmsEncryptorLocalstack(t *testing.T) {
	endpoint := "http://localhost:4566"

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	cfg := testutils.NewLocalstackConfig(t, ctx)
	v2client := kms.NewFromConfig(cfg, func(o *kms.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
	createKey, err := v2client.CreateKey(ctx, &kms.CreateKeyInput{})
	assert.NoError(t, err)
	uri := fmt.Sprintf("aws-kms://%s", *createKey.KeyMetadata.Arn)
	fmt.Printf("URI: %s\n", uri)

	// tink does not support awsv2 yet so here be dragons
	// https://github.com/tink-crypto/tink-go-awskms/issues/2
	s := awsv1session.Must(awsv1session.NewSession())
	v1client := awsv1kms.New(s, &awsv1.Config{
		Credentials: awsv1credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    aws.String(endpoint),
		Region:      aws.String("us-west-2"),
	})

	encryptor, err := NewKmsEncryptorGenerateKey(uri, v1client)
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
