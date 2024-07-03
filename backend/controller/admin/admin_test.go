package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestAdminService(t *testing.T) {
	config := tempConfigPath(t, "testdata/ftl-project.toml", "admin")
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm, err := cf.NewConfigurationManager(ctx, cf.ProjectConfigResolver[cf.Configuration]{Config: config})
	assert.NoError(t, err)

	sm, err := cf.New(ctx,
		cf.ProjectConfigResolver[cf.Secrets]{Config: config},
		[]cf.Provider[cf.Secrets]{
			cf.EnvarProvider[cf.Secrets]{},
			cf.InlineProvider[cf.Secrets]{},
		})
	assert.NoError(t, err)
	admin := NewAdminService(cm, sm, &diskSchemaRetriever{})
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
		existing = []byte("name = \"generated\"")
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

var testSchema = schema.MustValidate(&schema.Schema{
	Modules: []*schema.Module{
		{
			Name:     "batmobile",
			Comments: []string{"A batmobile comment"},
			Decls: []schema.Decl{
				&schema.Secret{
					Comments: []string{"top secret"},
					Name:     "owner",
					Type:     &schema.String{},
				},
				&schema.Secret{
					Comments: []string{"ultra secret"},
					Name:     "horsepower",
					Type:     &schema.Int{},
				},
				&schema.Config{
					Comments: []string{"car color"},
					Name:     "color",
					Type:     &schema.Ref{Module: "batmobile", Name: "Color"},
				},
				&schema.Config{
					Comments: []string{"car capacity"},
					Name:     "capacity",
					Type:     &schema.Ref{Module: "batmobile", Name: "Capacity"},
				},
				&schema.Enum{
					Comments: []string{"Car colors"},
					Name:     "Color",
					Type:     &schema.String{},
					Variants: []*schema.EnumVariant{
						{Name: "Black", Value: &schema.StringValue{Value: "Black"}},
						{Name: "Blue", Value: &schema.StringValue{Value: "Blue"}},
						{Name: "Green", Value: &schema.StringValue{Value: "Green"}},
					},
				},
				&schema.Enum{
					Comments: []string{"Car capacities"},
					Name:     "Capacity",
					Type:     &schema.Int{},
					Variants: []*schema.EnumVariant{
						{Name: "One", Value: &schema.IntValue{Value: int(1)}},
						{Name: "Two", Value: &schema.IntValue{Value: int(2)}},
						{Name: "Four", Value: &schema.IntValue{Value: int(4)}},
					},
				},
			},
		},
	},
})

type mockSchemaRetriever struct {
}

func (d *mockSchemaRetriever) GetActiveSchema(ctx context.Context) (*schema.Schema, error) {
	return testSchema, nil
}

func TestAdminValidation(t *testing.T) {
	config := tempConfigPath(t, "testdata/ftl-project.toml", "admin")
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm, err := cf.NewConfigurationManager(ctx, cf.ProjectConfigResolver[cf.Configuration]{Config: config})
	assert.NoError(t, err)

	sm, err := cf.New(ctx,
		cf.ProjectConfigResolver[cf.Secrets]{Config: config},
		[]cf.Provider[cf.Secrets]{
			cf.EnvarProvider[cf.Secrets]{},
			cf.InlineProvider[cf.Secrets]{},
		})
	assert.NoError(t, err)
	admin := NewAdminService(cm, sm, &mockSchemaRetriever{})
	assert.NotZero(t, admin)

	testSetConfig(t, ctx, admin, "batmobile", "color", "Black", "")
	testSetConfig(t, ctx, admin, "batmobile", "color", "Red", "JSON validation failed: Red is not a valid variant of enum batmobile.Color")
	testSetConfig(t, ctx, admin, "batmobile", "capacity", 2, "")
	testSetConfig(t, ctx, admin, "batmobile", "capacity", 3, "JSON validation failed: %!s(float64=3) is not a valid variant of enum batmobile.Capacity")

	testSetSecret(t, ctx, admin, "batmobile", "owner", "Bruce Wayne", "")
	testSetSecret(t, ctx, admin, "batmobile", "owner", 99, "JSON validation failed: owner has wrong type, expected String found float64")
	testSetSecret(t, ctx, admin, "batmobile", "horsepower", 1000, "")
	testSetSecret(t, ctx, admin, "batmobile", "horsepower", "thousand", "JSON validation failed: horsepower has wrong type, expected Int found string")

	testSetConfig(t, ctx, admin, "", "city", "Gotham", "")
	testSetSecret(t, ctx, admin, "", "universe", "DC", "")
}

// nolint
func testSetConfig(t testing.TB, ctx context.Context, admin *AdminService, module string, name string, jsonVal any, expectedError string) {
	t.Helper()
	buffer, err := json.Marshal(jsonVal)
	assert.NoError(t, err)

	configRef := &ftlv1.ConfigRef{Name: name}
	if module != "" {
		configRef.Module = &module
	}

	_, err = admin.ConfigSet(ctx, connect.NewRequest(&ftlv1.SetConfigRequest{
		Provider: ftlv1.ConfigProvider_CONFIG_INLINE.Enum(),
		Ref:      configRef,
		Value:    buffer,
	}))
	assert.EqualError(t, err, expectedError)
}

// nolint
func testSetSecret(t testing.TB, ctx context.Context, admin *AdminService, module string, name string, jsonVal any, expectedError string) {
	t.Helper()
	buffer, err := json.Marshal(jsonVal)
	assert.NoError(t, err)

	configRef := &ftlv1.ConfigRef{Name: name}
	if module != "" {
		configRef.Module = &module
	}

	_, err = admin.SecretSet(ctx, connect.NewRequest(&ftlv1.SetSecretRequest{
		Provider: ftlv1.SecretProvider_SECRET_INLINE.Enum(),
		Ref:      configRef,
		Value:    buffer,
	}))
	assert.EqualError(t, err, expectedError)
}
