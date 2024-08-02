package encryption

type Encryptable interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

func NewForKey(key []byte) Encryptable {
	if len(key) == 0 {
		return NoOpEncrypter{}
	}
	return Encrypter{key: key}
}

type Encrypter struct {
	key []byte
}

func (e Encrypter) Encrypt(data []byte) ([]byte, error) {
	// TODO: implement
	return data, nil
}

func (e Encrypter) Decrypt(data []byte) ([]byte, error) {
	// TODO: implement
	return data, nil
}

type NoOpEncrypter struct{}

func (e NoOpEncrypter) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (e NoOpEncrypter) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}
