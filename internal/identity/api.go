package identity

type SignedData struct {
	data      []byte
	signature []byte
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
	// TODO: Should hide the data until verified, i.e. return the data if the signature is valid.
	Verify(signedData SignedData) ([]byte, error)
}
