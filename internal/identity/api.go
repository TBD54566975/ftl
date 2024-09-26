package identity

type SignedData struct {
	// data is hidden here so that the only way to access it is to verify the signature.
	data      []byte
	Signature []byte
}

type PublicKey = []byte

type KeyPair interface {
	Signer() (Signer, error)
	Verifier() (Verifier, error)
	Public() (PublicKey, error)
}

type Signer interface {
	Sign(data []byte) (*SignedData, error)
	Public() (PublicKey, error)
}

type Verifier interface {
	Verify(signedData SignedData) ([]byte, error)
}
