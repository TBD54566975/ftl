package identity

type KeyPair interface {
	Signer() (Signer, error)
	Verifier() (Verifier, error)
	Public() ([]byte, error)
}

type Signer interface {
	Sign(data []byte) (*SignedData, error)
	Public() ([]byte, error)
}

type SignedData struct {
	Data      []byte
	Signature []byte
}

type Verifier interface {
	Verify(signedData SignedData) error
}
