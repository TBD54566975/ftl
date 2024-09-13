package encryption

import (
	"bytes"
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	ftlencryption "github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestEncryptionService(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	uri := "fake-kms://CK6YwYkBElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEJy4TIQgfCuwxA3ZZgChp_wYARABGK6YwYkBIAE"

	t.Run("EncryptDecryptJSON", func(t *testing.T) {
		service, err := New(ctx, conn, ftlencryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.NoError(t, err)

		type TestStruct struct {
			Name string
			Age  int
		}

		original := TestStruct{Name: "John Doe", Age: 30}
		var encrypted ftlencryption.EncryptedTimelineColumn
		err = service.EncryptJSON(original, &encrypted)
		assert.NoError(t, err)

		var decrypted TestStruct
		err = service.DecryptJSON(&encrypted, &decrypted)
		assert.NoError(t, err)

		assert.Equal(t, original, decrypted)
	})

	t.Run("EncryptDecryptBinary", func(t *testing.T) {
		service, err := New(ctx, conn, ftlencryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.NoError(t, err)

		original := []byte("Hello, World!")
		var encrypted ftlencryption.EncryptedTimelineColumn
		err = service.Encrypt(original, &encrypted)
		assert.NoError(t, err)

		decrypted, err := service.Decrypt(&encrypted)
		assert.NoError(t, err)

		assert.True(t, bytes.Equal(original, decrypted))
	})
}
