package identity

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

// CertificateContent is used as a certificate request and also the content of a certificate.
// It is "content" that is to be signed outside of this struct.
type CertificateContent struct {
	Identity  Identity     `protobuf:"bytes,1,opt,name=identity,proto3" json:"identity,omitempty"`
	PublicKey RawPublicKey `protobuf:"bytes,2,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

func (c CertificateContent) ProtoReflect() protoreflect.Message {
	panic("unimplemented")
}

var _ proto.Message = CertificateContent{}

func (c CertificateContent) String() string {
	return fmt.Sprintf("CertficateContent(%s %x)", c.Identity, c.PublicKey)
}

type Certificate struct {
	CertificateContent
	Signature Signature
}

func ParseCertificateFromProto(protoCert *ftlv1.Certificate) (Certificate, error) {
	encoded := ParseSignedMessageFromProto(protoCert.SignedMessage)
	var certificateContent CertificateContent
	if err := proto.Unmarshal(encoded.message, &certificateContent); err != nil {
		return Certificate{}, fmt.Errorf("failed to unmarshal certificate content: %w", err)
	}

	return Certificate{
		CertificateContent: CertificateContent{
			Identity:  certificateContent.Identity,
			PublicKey: certificateContent.PublicKey,
		},
		Signature: encoded.Signature,
	}, nil
}

func (c Certificate) String() string {
	return fmt.Sprintf("Certificate(key:%x sig:%x)", c.PublicKey.Bytes, c.Signature.Bytes)
}

func (c Certificate) ToProto() *ftlv1.Certificate {
	certificateContent, err := proto.Marshal(&c.CertificateContent)
	if err != nil {
		panic(fmt.Errorf("failed to marshal certificate content: %w", err))
	}

	return &ftlv1.Certificate{
		SignedMessage: &ftlv1.SignedMessage{
			Message:   certificateContent,
			Signature: c.Signature.Bytes,
		},
	}
}

// CertifiedSignedData is sent by a node and proves identity based on a certificate.
type CertifiedSignedData struct {
	Certificate Certificate
	SignedData  SignedMessage
}

func (c CertifiedSignedData) String() string {
	return fmt.Sprintf("CertifiedSignedData data:%x signature:%x (%s)", c.SignedData.message, c.SignedData.Signature, c.Certificate)
}

// Verify against the CA and then the node certificate. Only return the data if both are valid.
func (c CertifiedSignedData) Verify(caVerifier Verifier) (Identity, []byte, error) {
	// Verify against the CA certificate.
	data, err := caVerifier.Verify(c.Certificate.SignedData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	var certificateContent ftlv1.CertificateContent
	if err = proto.Unmarshal(data, &certificateContent); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal certificate content: %w", err)
	}

	identity, err := Parse(certificateContent.Identity)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse identity: %w", err)
	}

	nodePublicKey := RawPublicKey{Bytes: certificateContent.PublicKey}
	nodeVerifier, err := NewVerifier(nodePublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create verifier: %w", err)
	}
	payload, err := nodeVerifier.Verify(c.SignedData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify signed data: %w", err)
	}

	return identity, payload, nil
}
