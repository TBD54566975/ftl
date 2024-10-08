package identity

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
)

// type SignedMessage struct {
// 	// message is hidden here to prevent misuse.
// 	// Use Verifier.Verify() to get the message while verifying the signature.
// 	message   []byte
// 	Signature Signature
// }

// // NewSignedMessage ensures that the data is signed correctly.
// func NewSignedMessage(verifier Verifier, data []byte, signature Signature) (SignedMessage, error) {
// 	signedMessage := SignedMessage{message: data, Signature: signature}

// 	_, err := verifier.Verify(signedMessage)
// 	if err != nil {
// 		return SignedMessage{}, fmt.Errorf("failed to verify signed message: %w", err)
// 	}

// 	return signedMessage, nil
// }

// func (sm SignedMessage) VerifiedMessage(verifier Verifier) ([]byte, error) {
// 	message, err := verifier.Verify(sm)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to verify signed message: %w", err)
// 	}
// 	return message, nil
// }

// // UnverifiedMessage returns the message without checking the signature.
// // Don't use this unless there's a good reason to do so (e.g. needing the public key before having access to the signature)
// func (sm SignedMessage) UnverifiedMessage() []byte {
// 	return sm.message
// }

// func (sm SignedMessage) ToProto() *identitypb.SignedMessage {
// 	return &identitypb.SignedMessage{
// 		Message:   sm.message,
// 		Signature: sm.Signature.Bytes,
// 	}
// }

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
