package encryption

import (
	"fmt"
	"github.com/alecthomas/assert/v2"
	"testing"
)

const key = `{
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

func TestNewEncrypter(t *testing.T) {
	encrpyter, err := NewForKey(key)
	assert.NoError(t, err)

	encrypted, err := encrpyter.EncryptJSON("\"hello\"")
	assert.NoError(t, err)

	fmt.Printf("Encrypted: %s\n", encrypted)

	var decrypted string
	err = encrpyter.DecryptJSON(encrypted, &decrypted)
	assert.NoError(t, err)
}
