package identity

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	identitypb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/identity"
)

// CertificateContent is used as a certificate request and also the content of a certificate.
// It is "content" that is to be signed outside of this struct.
type CertificateContent struct {
	Identity  Identity
	PublicKey RawPublicKey
}

func CertificateContentFromProto(proto *identitypb.CertificateContent) (CertificateContent, error) {
	identity, err := Parse(proto.Identity)
	if err != nil {
		return CertificateContent{}, fmt.Errorf("failed to parse identity: %w", err)
	}

	return CertificateContent{
		Identity:  identity,
		PublicKey: NewRawPublicKey(proto.PublicKey),
	}, nil
}

func (c CertificateContent) ToProto() *identitypb.CertificateContent {
	return &identitypb.CertificateContent{
		Identity:  c.Identity.String(),
		PublicKey: c.PublicKey.Bytes,
	}
}

func (c CertificateContent) String() string {
	return fmt.Sprintf("CertficateContent(%s %x)", c.Identity, c.PublicKey)
}

type CertificateRequest struct {
	CertificateContent
	Signature Signature
}

// CertificateRequestFromProto does not verify the signature.
func CertificateRequestFromProto(proto *identitypb.CertificateRequest) (CertificateRequest, error) {
	content, err := CertificateContentFromProto(proto.Content)
	if err != nil {
		return CertificateRequest{}, fmt.Errorf("failed to parse certificate content: %w", err)
	}

	return CertificateRequest{
		CertificateContent: content,
		Signature:          NewSignature(proto.Signature),
	}, nil
}

func (c CertificateRequest) ToProto() *identitypb.CertificateRequest {
	return &identitypb.CertificateRequest{
		Content:   c.CertificateContent.ToProto(),
		Signature: c.Signature.Bytes,
	}
}

type Certificate struct {
	CertificateContent
	Signature Signature
}

func (c Certificate) Verify(caVerifier Verifier) error {
	encoded, err := proto.Marshal(c.CertificateContent.ToProto())
	if err != nil {
		return fmt.Errorf("failed to marshal certificate content: %w", err)
	}

	if err = caVerifier.Verify(c.Signature, encoded); err != nil {
		return fmt.Errorf("failed to verify ca certificate: %w", err)
	}

	return nil
}

// func (c Certificate) ToSignedMessage() (SignedMessage, error) {
// 	// encoded, err := proto.Marshal(&c.CertificateContent.ToProto())
// 	// if err != nil {
// 	// 	return SignedMessage{}, fmt.Errorf("failed to marshal certificate content: %w", err)
// 	// }

// 	// return SignedMessage{
// 	// 	message:   encoded,
// 	// 	Signature: c.Signature,
// 	// }, nil

// 	panic("not implemented")
// }

func ParseCertificateFromProto(protoCert *identitypb.Certificate) (Certificate, error) {
	// encoded := ParseSignedMessageFromProto(protoCert.SignedMessage)
	// var certificateContent CertificateContent
	// if err := proto.Unmarshal(encoded.message, &certificateContent); err != nil {
	// 	return Certificate{}, fmt.Errorf("failed to unmarshal certificate content: %w", err)
	// }

	// return Certificate{
	// 	CertificateContent: CertificateContent{
	// 		Identity:  certificateContent.Identity,
	// 		PublicKey: certificateContent.PublicKey,
	// 	},
	// 	Signature: encoded.Signature,
	// }, nil
	//
	panic("not implemented")
}

func (c Certificate) String() string {
	return fmt.Sprintf("Certificate(key:%x sig:%x)", c.PublicKey.Bytes, c.Signature.Bytes)
}

func (c Certificate) ToProto() *identitypb.Certificate {
	return &identitypb.Certificate{
		Content:   c.CertificateContent.ToProto(),
		Signature: c.Signature.Bytes,
	}
}

// CertifiedSignedData is sent by a node and proves identity based on a certificate.
type CertifiedSignedData struct {
	Certificate Certificate
	Message     []byte
	Signature   Signature
}

func (c CertifiedSignedData) String() string {
	return fmt.Sprintf("CertifiedSignedData data:%x signature:%x (%s)", c.Message, c.Signature, c.Certificate)
}

// Verify against the CA and then the node certificate. Only return the data if both are valid.
func (c CertifiedSignedData) Verify(caVerifier Verifier) (Identity, []byte, error) {
	if err := c.Certificate.Verify(caVerifier); err != nil {
		return nil, nil, fmt.Errorf("failed to verify ca certificate cert:%s: %w", c.Certificate, err)
	}

	// Now verify the given data with the public key of the certificate.
	verifier, err := NewVerifier(c.Certificate.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create verifier: %w", err)
	}

	if err = verifier.Verify(c.Signature, c.Message); err != nil {
		return nil, nil, fmt.Errorf("failed to verify message: %w", err)
	}

	return c.Certificate.Identity, c.Message, nil
}
