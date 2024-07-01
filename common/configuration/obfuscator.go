package configuration

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/TBD54566975/ftl/internal/slices"
)

type ObfuscatorProvider interface {
	obfuscator() Obfuscator
}

// Obfuscator hides and reveals a value, but does not provide real security
// instead the aim of this Obfuscator is to make values not easily human readable while attaching a comment
//
// Obfuscation is done by XOR-ing the input with the AES key. Length of key must be 16, 24 or 32 bytes (corresponding to AES-128, AES-192 or AES-256 keys).
//
// Example obfuscation result:
// # Comments appear at top
// # Multiple lines are supported
// d144b654d69a438cf7bcaaa59dee430e51d5c3cbeb6f61d8bc3f79f3484e7234fab5280ac57678d68e6c
type Obfuscator struct {
	key []byte
}

// Obfuscate takes a value and returns an obfuscated value (encoded in hex) with a comment
func (o Obfuscator) Obfuscate(input []byte, comment string) ([]byte, error) {
	encoded, err := o.encode(input)
	if err != nil {
		return nil, err
	}
	// build output by prepending comments
	lines := []string{}
	for _, line := range strings.Split(comment, "\n") {
		lines = append(lines, "# "+line)
	}
	lines = append(lines, hex.EncodeToString(encoded))
	return []byte(strings.Join(lines, "\n")), nil
}

// Reveal takes an obfuscated value, ignores any comments (lines starting with '#') and de-obfuscates the hex encoded value
func (o Obfuscator) Reveal(input []byte) ([]byte, error) {
	// find first line which is not a comment
	hexEncoded, ok := slices.Find(strings.Split(string(input), "\n"), func(line string) bool {
		return !strings.HasPrefix(line, "#")
	})
	if !ok {
		return nil, fmt.Errorf("could not find obfuscated value")
	}
	obfuscated, err := hex.DecodeString(hexEncoded)
	if err != nil {
		return nil, fmt.Errorf("expected hexadecimal string: %w", err)
	}
	return o.decode(obfuscated)
}

// encode takes a byte slice and returns an obfuscated (XOR with AES key) byte slice
func (o Obfuscator) encode(input []byte) ([]byte, error) {
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
	return ciphertext, nil
}

// decode takes a obfuscated byte slice and decrypts
func (o Obfuscator) decode(input []byte) ([]byte, error) {
	block, err := aes.NewCipher(o.key)
	if err != nil {
		return nil, fmt.Errorf("could not create cypher for decoding obfuscation: %w", err)
	}
	if len(input) < aes.BlockSize {
		return nil, errors.New("obfuscated value too short to decode")
	}
	iv := input[:aes.BlockSize]
	input = input[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)

	var output = make([]byte, len(input))
	cfb.XORKeyStream(output, input)

	return output, nil
}
