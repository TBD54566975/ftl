package configuration

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

type ObfuscatorProvider interface {
	obfuscator() Obfuscator
}

// Obfuscator hides and reveals a value, but does not provide real security
// instead the aim of this Obfuscator is to make values not easily human readable
//
// Obfuscation is done by XOR-ing the input with the AES key. Length of key must be 16, 24 or 32 bytes (corresponding to AES-128, AES-192 or AES-256 keys).
type Obfuscator struct {
	key []byte
}

// Obfuscate takes a value and returns an obfuscated value (encoded in base64)
func (o Obfuscator) Obfuscate(input []byte) ([]byte, error) {
	block, err := aes.NewCipher(o.key)
	if err != nil {
		return nil, fmt.Errorf("could not create cypher for obfuscation: %w", err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(input))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("could not generate IV for obfuscation: %w", err)
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], input)
	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// Reveal takes an obfuscated value and de-obfuscates the base64 encoded value
func (o Obfuscator) Reveal(input []byte) ([]byte, error) {
	// check if the input looks like it was obfuscated
	if !strings.ContainsRune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/=", rune(input[0])) {
		// known issue: an unobfuscated value which is just a number will be considered obfuscated
		return input, nil
	}

	obfuscated, err := base64.StdEncoding.DecodeString(string(input))
	if err != nil {
		return nil, fmt.Errorf("expected base64 string: %w", err)
	}
	block, err := aes.NewCipher(o.key)
	if err != nil {
		return nil, fmt.Errorf("could not create cypher for decoding obfuscation: %w", err)
	}
	if len(obfuscated) < aes.BlockSize {
		return nil, errors.New("obfuscated value too short to decode")
	}
	iv := obfuscated[:aes.BlockSize]
	obfuscated = obfuscated[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)

	var output = make([]byte, len(obfuscated))
	cfb.XORKeyStream(output, obfuscated)

	return output, nil
}
