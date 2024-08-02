package encryption

type EncryptionManager interface {
	Encrypt(plain string) (string, error)

	Decrypt(encrypted string) (string, error)
}
