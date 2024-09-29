package identity

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
)

type SignedData struct {
	// data is hidden here so that the only way to access it is to verify the signature.
	data      []byte
	Signature []byte
}

// NewSignedData ensures that the data is signed correctly.
func NewSignedData(verifier Verifier, data, signature []byte) (SignedData, error) {
	signedData := SignedData{data: data, Signature: signature}

	_, err := verifier.Verify(signedData)
	if err != nil {
		return SignedData{}, fmt.Errorf("failed to verify data: %w", err)
	}

	return signedData, nil
}

type PublicKey struct {
	Bytes []byte
}

func NewPublicKey(b []byte) PublicKey {
	return PublicKey{Bytes: b}
}

func (pk PublicKey) Decode(ctx *kong.DecodeContext) error {
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
