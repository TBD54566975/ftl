package encryption

import (
	"encoding/json"
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
	jsonInput := "\"hello\""

	encrypter, err := NewForKey(key)
	assert.NoError(t, err)

	encrypted, err := encrypter.EncryptJSON(jsonInput)
	assert.NoError(t, err)
	fmt.Printf("Encrypted: %s\n", encrypted)

	var decrypted json.RawMessage
	err = encrypter.DecryptJSON(encrypted, &decrypted)
	assert.NoError(t, err)
	fmt.Printf("Decrypted: %s\n", decrypted)

	var decryptedString string
	err = json.Unmarshal(decrypted, &decryptedString)
	assert.NoError(t, err)
	fmt.Printf("Decrypted string: %s\n", decryptedString)

	assert.Equal(t, jsonInput, decryptedString)
}
