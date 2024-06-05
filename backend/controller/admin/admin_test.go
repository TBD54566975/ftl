package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestAdminService(t *testing.T) {
	config := tempConfigPath(t, "testdata/ftl-project.toml", "admin")
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm, err := cf.NewConfigurationManager(ctx, cf.ProjectConfigResolver[cf.Configuration]{Config: []string{config}})
	assert.NoError(t, err)

	sm, err := cf.New(ctx,
		cf.ProjectConfigResolver[cf.Secrets]{Config: []string{config}},
		[]cf.Provider[cf.Secrets]{
			cf.EnvarProvider[cf.Secrets]{},
			cf.InlineProvider[cf.Secrets]{},
		})
	assert.NoError(t, err)
	admin := NewAdminService(cm, sm)
	assert.NotZero(t, admin)

	expectedEnvarValue, err := json.MarshalIndent(map[string]string{"bar": "barfoo"}, "", "  ")
	assert.NoError(t, err)

	testAdminConfigs(t, ctx, "FTL_CONFIG_YmFy", admin, []expectedEntry{
		{Ref: cf.Ref{Name: "bar"}, Value: string(expectedEnvarValue)},
		{Ref: cf.Ref{Name: "foo"}, Value: `"foobar"`},
		{Ref: cf.Ref{Name: "mutable"}, Value: `"helloworld"`},
		{Ref: cf.Ref{Module: optional.Some[string]("echo"), Name: "default"}, Value: `"anonymous"`},
	})

	testAdminSecrets(t, ctx, "FTL_SECRET_YmFy", admin, []expectedEntry{
		{Ref: cf.Ref{Name: "bar"}, Value: string(expectedEnvarValue)},
		{Ref: cf.Ref{Name: "foo"}, Value: `"foobarsecret"`},
	})
}

type expectedEntry struct {
	Ref   cf.Ref
	Value string
}

func tempConfigPath(t *testing.T, existingPath string, prefix string) string {
	t.Helper()
	config := filepath.Join(t.TempDir(), fmt.Sprintf("%s-ftl-project.toml", prefix))
	var existing []byte
	var err error
	if existingPath == "" {
		existing = []byte{}
	} else {
		existing, err = os.ReadFile(existingPath)
		assert.NoError(t, err)
	}
	err = os.WriteFile(config, existing, 0600)
	assert.NoError(t, err)
	return config
}

// nolint
func testAdminConfigs(
	t *testing.T,
	ctx context.Context,
	envarName string,
	admin *AdminService,
	entries []expectedEntry,
) {
	t.Helper()
	t.Setenv(envarName, "eyJiYXIiOiJiYXJmb28ifQ") // bar={"bar": "barfoo"}

	module := ""
	includeValues := true
	resp, err := admin.ConfigList(ctx, connect.NewRequest(&ftlv1.ListConfigRequest{
		Module:        &module,
		IncludeValues: &includeValues,
	}))
	assert.NoError(t, err)
	assert.NotZero(t, resp)

	configs := resp.Msg.Configs
	assert.Equal(t, len(entries), len(configs))

	for _, entry := range entries {
		module := entry.Ref.Module.Default("")
		ref := &ftlv1.ConfigRef{
			Module: &module,
			Name:   entry.Ref.Name,
		}
		resp, err := admin.ConfigGet(ctx, connect.NewRequest(&ftlv1.GetConfigRequest{Ref: ref}))
		assert.NoError(t, err)
		assert.Equal(t, entry.Value, string(resp.Msg.Value))
	}
}

// nolint
func testAdminSecrets(
	t *testing.T,
	ctx context.Context,
	envarName string,
	admin *AdminService,
	entries []expectedEntry,
) {
	t.Helper()
	t.Setenv(envarName, "eyJiYXIiOiJiYXJmb28ifQ") // bar={"bar": "barfoo"}

	module := ""
	includeValues := true
	resp, err := admin.SecretsList(ctx, connect.NewRequest(&ftlv1.ListSecretsRequest{
		Module:        &module,
		IncludeValues: &includeValues,
	}))
	assert.NoError(t, err)
	assert.NotZero(t, resp)

	secrets := resp.Msg.Secrets
	assert.Equal(t, len(entries), len(secrets))

	for _, entry := range entries {
		module := entry.Ref.Module.Default("")
		ref := &ftlv1.ConfigRef{
			Module: &module,
			Name:   entry.Ref.Name,
		}
		resp, err := admin.SecretGet(ctx, connect.NewRequest(&ftlv1.GetSecretRequest{Ref: ref}))
		assert.NoError(t, err)
		assert.Equal(t, entry.Value, string(resp.Msg.Value))
	}
}
