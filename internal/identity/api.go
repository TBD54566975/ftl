package identity

type KeyPair interface {
	Signer() (Signer, error)
	Verifier() (Verifier, error)
}

type Verifier interface {
	Verify(signedData SignedData) error
}

type Signer interface {
	Sign(data []byte) (*SignedData, error)
}

type SignedData struct {
	Data      []byte
	Signature []byte
}
