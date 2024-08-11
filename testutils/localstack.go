package testutils

import (
	"context"
	"github.com/alecthomas/assert/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"testing"
)

func NewLocalstackConfig(t *testing.T, ctx context.Context) aws.Config { // nolint: revive
	t.Helper()
	cc := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", ""))
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(cc), config.WithRegion("us-west-2"))
	assert.NoError(t, err)
	return cfg
}
