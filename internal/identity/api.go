package identity

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
)

type Signature struct {
	Bytes []byte
}

func NewSignature(b []byte) Signature {
	return Signature{Bytes: b}
}

// RawPublicKey is a public key that has not been parsed into a Tink keyset handle.
type RawPublicKey struct {
	Bytes []byte
}

func NewRawPublicKey(b []byte) RawPublicKey {
	return RawPublicKey{Bytes: b}
}

// Decode here is used for parsing the public key from a base64 string when passed
// as a command line argument or environment variable.
func (pk RawPublicKey) Decode(ctx *kong.DecodeContext) error {
	var b64 string
	err := ctx.Scan.PopValueInto("string", &b64)
	if err != nil {
		return fmt.Errorf("failed to pop public key: %w", err)
	}

	pk.Bytes, err = base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return fmt.Errorf("failed to decode base64 public key: %w", err)
	}

	ctx.Value.Target.Set(reflect.ValueOf(pk))
	return nil
}
