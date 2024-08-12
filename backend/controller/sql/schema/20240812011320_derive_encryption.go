package schema

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// Migrate20240811162246DeriveEncryption
//
// If there are settings set for DeprecatedEncryptionKeys, then check that there is also a KMSURI. If not error out.
// If there is no DeprecatedEncryptionKeys, use the NoOpEncryptor so the data will just be transformed to bytea.
//   - but... also make sure none of the data is already encrypted!
//
// keys:
// - Create a table called encryption_keys with columns id, encrypted
//
// events:
// - Add events.payload_new as type BYTEA
// - Add events.encryption_key_id as type BIGINT as a foreign key to encryption_keys.id
//
// async_calls:
// - Add async_calls.request_new as type BYTEA
// - !! REMOVE async_calls.response because we don't use it
// - Add async_calls.encryption_key_id as type BIGINT as a foreign key to encryption_keys.id
//
// topic_events:
// - topic_events.payload is already BYTEA but still create a new column called payload_new as type BYTEA
// - Add topic_events.encryption_key_id as type BIGINT as a foreign key to encryption_keys.id
//
// Iterate over all events and async_calls and decrypt using the appropriate encryption.DeprecatedEncryptable
// and re-encrypt using the new encryption.Encryptor.
// For topic.request, Similar to events and async_calls.
//
// Drop the old columns and rename the new columns to the old column names.
func Migrate20240811162246DeriveEncryption(ctx context.Context, tx *sql.Tx) error {
	encryption, err := sanityCheck(ctx, tx)
	if err != nil {
		return fmt.Errorf("sanity check failed: %w", err)
	}

	if encryption {
		if err := migrateDataEncrypted(ctx, tx); err != nil {
			return fmt.Errorf("failed to migrate encrypted data: %w", err)
		}
	} else {
		if err := migrateDataUnencrypted(ctx, tx); err != nil {
			return fmt.Errorf("failed to migrate unencrypted data: %w", err)
		}
	}

	return nil
}

func sanityCheck(ctx context.Context, tx *sql.Tx) (bool, error) {
	logKey := os.Getenv("FTL_LOG_ENCRYPTION_KEY")
	asyncKey := os.Getenv("FTL_ASYNC_ENCRYPTION_KEY")
	kmsURI := os.Getenv("FTL_KMS_URI")
	encryption := false
	if logKey != "" || asyncKey != "" {
		if kmsURI == "" {
			return false, fmt.Errorf("deprecated encryption keys are set but no KMS URI was set, refuse to migrate")
		}
		encryption = true
	}

	if err := checkTableForEncryption(ctx, tx, "SELECT payload FROM events", encryption); err != nil {
		return false, fmt.Errorf("failed to check events: %w", err)
	}
	if err := checkTableForEncryption(ctx, tx, "SELECT request FROM async_calls", encryption); err != nil {
		return false, fmt.Errorf("failed to check async_calls: %w", err)
	}
	if err := checkTableForEncryption(ctx, tx, "SELECT payload FROM topic_events", encryption); err != nil {
		return false, fmt.Errorf("failed to check topic_events: %w", err)
	}

	return encryption, nil
}

func checkTableForEncryption(ctx context.Context, tx *sql.Tx, sql string, encryption bool) error {
	rows, err := tx.QueryContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}
		if err = encryptionCheck(encryption, payload); err != nil {
			return fmt.Errorf("encryption mismatch: %w", err)
		}
	}

	return nil
}

// encryptionCheck makes sure that the data is encrypted if encryption is set to true and vice versa.
// encrypted payloads start with `{"encrypted":`
func encryptionCheck(encryption bool, payload []byte) error {
	isEncrypted := strings.HasPrefix(string(payload), `{"encrypted":`)

	if isEncrypted && !encryption {
		return fmt.Errorf("data is encrypted but encryption is not active")
	}
	if !isEncrypted && encryption {
		return fmt.Errorf("data is not encrypted but encryption is active")
	}

	return nil
}

// migrateDataEncrypted will migrate data as is to the new field unencrypted using SQL.
func migrateDataUnencrypted(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE events SET payload_new = payload;
		UPDATE async_calls SET request_new = request;
		UPDATE topic_events SET payload_new = payload;
	`); err != nil {
		return fmt.Errorf("failed to add new columns: %w", err)
	}

	return nil
}

// migrateDataEncrypted will migrate data from the old field to the new field using the appropriate encryption.
func migrateDataEncrypted(ctx context.Context, tx *sql.Tx) error {
	if err := migrateData(ctx, tx, "events", "payload", "payload_new"); err != nil {
		return fmt.Errorf("failed to migrate events: %w", err)
	}
	if err := migrateData(ctx, tx, "async_calls", "request", "request_new"); err != nil {
		return fmt.Errorf("failed to migrate async_calls: %w", err)
	}
	if err := migrateData(ctx, tx, "topic_events", "payload", "payload_new"); err != nil {
		return fmt.Errorf("failed to migrate topic_events: %w", err)
	}

	return nil
}
