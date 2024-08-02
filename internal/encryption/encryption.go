package encryption

import "encoding/json"

type Encryptable interface {
	EncryptJSON(input any) (json.RawMessage, error)
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

func (e Encrypter) EncryptJSON(input any) (json.RawMessage, error) {
	// TODO: implement
	return json.RawMessage{}, nil
}

func (e Encrypter) DecryptJSON(input json.RawMessage, output any) error {
	// TODO: implement
	return nil
}

type NoOpEncrypter struct{}

func (e NoOpEncrypter) EncryptJSON(input any) (json.RawMessage, error) {
	return json.RawMessage{}, nil
}

func (e NoOpEncrypter) DecryptJSON(input json.RawMessage, output any) error {
	return nil
}
