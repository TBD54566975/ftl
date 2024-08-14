//go:build integration

package encryption

import (
	"context"
	"fmt"
	"testing"
	"time"

	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/internal/testutils"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awsv1 "github.com/aws/aws-sdk-go/aws"
	awsv1credentials "github.com/aws/aws-sdk-go/aws/credentials"
	awsv1session "github.com/aws/aws-sdk-go/aws/session"
	awsv1kms "github.com/aws/aws-sdk-go/service/kms"
)

func WithEncryption() in.Option {
	return in.WithEnvar("FTL_KMS_URI", "fake-kms://CKbvh_ILElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEE6tD2yE5AWYOirhmkY-r3sYARABGKbvh_ILIAE")
}

func TestEncryptionForLogs(t *testing.T) {
	in.Run(t,
		WithEncryption(),
		in.CopyModule("encryption"),
		in.Deploy("encryption"),
		in.Call[map[string]interface{}, any]("encryption", "echo", map[string]interface{}{"name": "Alice"}, nil),

		// confirm that we can read an event for that call
		func(t testing.TB, ic in.TestContext) {
			in.Infof("Read Logs")
			resp, err := ic.Console.GetEvents(ic.Context, connect.NewRequest(&pbconsole.EventsQuery{
				Limit: 10,
			}))
			assert.NoError(t, err, "could not get events")
			_, ok := slices.Find(resp.Msg.Events, func(e *pbconsole.Event) bool {
				call, ok := e.Entry.(*pbconsole.Event_Call)
				if !ok {
					return false
				}
				assert.Contains(t, call.Call.Request, "Alice", "request does not contain expected value")

				return true
			})
			assert.True(t, ok, "could not find event")
		},

		// confirm that we can't find that raw request string in the table
		in.QueryRow("ftl", "SELECT COUNT(*) FROM timeline WHERE type = 'call'", int64(1)),
		func(t testing.TB, ic in.TestContext) {
			values := in.GetRow(t, ic, "ftl", "SELECT payload FROM timeline WHERE type = 'call' LIMIT 1", 1)
			payload, ok := values[0].([]byte)
			assert.True(t, ok, "could not convert payload to string")
			assert.NotContains(t, string(payload), "Alice", "raw request string should not be stored in the table")
		},
	)
}

func TestEncryptionForPubSub(t *testing.T) {
	in.Run(t,
		WithEncryption(),
		in.CopyModule("encryption"),
		in.Deploy("encryption"),
		in.Call[map[string]interface{}, any]("encryption", "publish", map[string]interface{}{"name": "AliceInWonderland"}, nil),

		in.Sleep(4*time.Second),

		// check that the event was published with an encrypted request
		in.QueryRow("ftl", "SELECT COUNT(*) FROM topic_events", int64(1)),
		func(t testing.TB, ic in.TestContext) {
			values := in.GetRow(t, ic, "ftl", "SELECT payload FROM topic_events", 1)
			payload, ok := values[0].([]byte)
			assert.True(t, ok, "could not convert payload to string")
			assert.NotContains(t, string(payload), "AliceInWonderland", "raw request string should not be stored in the table")
		},
		validateAsyncCall("consume", "AliceInWonderland"),
	)
}

func TestEncryptionForFSM(t *testing.T) {
	in.Run(t,
		WithEncryption(),
		in.CopyModule("encryption"),
		in.Deploy("encryption"),
		// "Rosebud" goes from created -> paid to test normal encryption
		in.Call[map[string]interface{}, any]("encryption", "beginFsm", map[string]interface{}{"name": "Rosebud"}, nil),
		in.Sleep(2*time.Second),
		in.Call[map[string]interface{}, any]("encryption", "transitionToPaid", map[string]interface{}{"name": "Rosebud"}, nil),
		in.Sleep(2*time.Second),
		validateAsyncCall("created", "Rosebud"),
		validateAsyncCall("paid", "Rosebud"),

		// "Next" goes from created -> nextAndSleep to test fsm next event encryption
		in.Call[map[string]interface{}, any]("encryption", "beginFsm", map[string]interface{}{"name": "Next"}, nil),
		in.Sleep(2*time.Second),
		in.Call[map[string]interface{}, any]("encryption", "transitionToNextAndSleep", map[string]interface{}{"name": "Next"}, nil),
		func(t testing.TB, ic in.TestContext) {
			in.QueryRow("ftl", "SELECT COUNT(*) FROM fsm_next_event LIMIT 1", int64(1))(t, ic)
			values := in.GetRow(t, ic, "ftl", "SELECT request FROM fsm_next_event LIMIT 1", 1)
			request, ok := values[0].([]byte)
			assert.True(t, ok, "could not convert payload to bytes")
			assert.NotContains(t, string(request), "Next", "raw request string should not be stored in the table")
		},
	)
}

func validateAsyncCall(verb string, sensitive string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE verb = 'encryption.%s' AND state = 'success'", verb), int64(1))(t, ic)

		values := in.GetRow(t, ic, "ftl", fmt.Sprintf("SELECT request FROM async_calls WHERE verb = 'encryption.%s' AND state = 'success'", verb), 1)
		request, ok := values[0].([]byte)
		assert.True(t, ok, "could not convert payload to bytes")
		assert.NotContains(t, string(request), sensitive, "raw request string should not be stored in the table")
	}
}

func TestKMSEncryptorLocalstack(t *testing.T) {
	endpoint := "http://localhost:4566"

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	cfg := testutils.NewLocalstackConfig(t, ctx)
	v2client := kms.NewFromConfig(cfg, func(o *kms.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
	createKey, err := v2client.CreateKey(ctx, &kms.CreateKeyInput{})
	assert.NoError(t, err)
	uri := fmt.Sprintf("aws-kms://%s", *createKey.KeyMetadata.Arn)
	fmt.Printf("URI: %s\n", uri)

	// tink does not support awsv2 yet so here be dragons
	// https://github.com/tink-crypto/tink-go-awskms/issues/2
	s := awsv1session.Must(awsv1session.NewSession())
	v1client := awsv1kms.New(s, &awsv1.Config{
		Credentials: awsv1credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    awsv1.String(endpoint),
		Region:      awsv1.String("us-west-2"),
	})

	encryptor, err := NewKMSEncryptorGenerateKey(uri, v1client)
	assert.NoError(t, err)

	var encrypted EncryptedTimelineColumn
	err = encryptor.Encrypt([]byte("hunter2"), &encrypted)
	assert.NoError(t, err)

	decrypted, err := encryptor.Decrypt(&encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", string(decrypted))

	// Should fail to decrypt with the wrong subkey
	wrongSubKey := EncryptedAsyncColumn(encrypted)
	_, err = encryptor.Decrypt(&wrongSubKey)
	assert.Error(t, err)
}
