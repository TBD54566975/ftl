package encryption

import "encoding/json"

type Encryptable interface {
	Encrypt(data []byte) ([]byte, error)
	EncryptJSON(input map[string]any) (json.RawMessage, error)
	Decrypt(data []byte) ([]byte, error)
	DecryptJSON(input json.RawMessage, output any) error
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

func (e Encrypter) EncryptJSON(input map[string]any) (json.RawMessage, error) {
	// TODO: implement
	return input, nil
}

func (e Encrypter) Decrypt(data []byte) ([]byte, error) {
	// TODO: implement
	return data, nil
}

func (e Encrypter) DecryptJSON(input json.RawMessage, output any) error {
	// TODO: implement
	return nil
}

type NoOpEncrypter struct{}

func (e NoOpEncrypter) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (e NoOpEncrypter) EncryptJSON(input map[string]any) (json.RawMessage, error) {
	return input, nil
}

func (e NoOpEncrypter) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (e NoOpEncrypter) DecryptJSON(input json.RawMessage, output any) error {
	return nil
}
