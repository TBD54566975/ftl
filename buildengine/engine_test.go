package buildengine_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestEngine(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	engine, err := buildengine.New(ctx, nil, nil, []string{"testdata/projects/alpha", "testdata/projects/other", "testdata/projects/another"}, nil)
	assert.NoError(t, err)

	defer engine.Close()

	// Import the schema from the third module, simulating a remote schema.
	otherSchema := &schema.Module{
		Name: "other",
		Decls: []schema.Decl{
			&schema.Data{
				Name: "EchoRequest",
				Fields: []*schema.Field{
					{Name: "name", Type: &schema.Optional{Type: &schema.String{}}, Metadata: []schema.Metadata{&schema.MetadataAlias{Alias: "name"}}},
				},
			},
			&schema.Data{
				Name: "EchoResponse",
				Fields: []*schema.Field{
					{Name: "message", Type: &schema.String{}, Metadata: []schema.Metadata{&schema.MetadataAlias{Alias: "message"}}},
				},
			},
			&schema.Verb{
				Name:     "echo",
				Request:  &schema.Ref{Module: "other", Name: "EchoRequest"},
				Response: &schema.Ref{Module: "other", Name: "EchoResponse"},
			},
		},
	}
	engine.Import(ctx, otherSchema)

	expected := map[string][]string{
		"alpha":   {"another", "other", "builtin"},
		"another": {"builtin"},
		"other":   {"another", "builtin"},
		"builtin": {},
	}
	graph, err := engine.Graph()
	assert.NoError(t, err)
	assert.Equal(t, expected, graph)
	err = engine.Build(ctx)
	assert.NoError(t, err)
}

func TestValidateConfigsAndSecretsMatch(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	projConfig, err := projectconfig.LoadConfig(ctx, []string{"testdata/projectconfigs/config-secret-validation-ftl-project.toml"})
	assert.NoError(t, err)
	engine, err := buildengine.New(ctx, nil, &projConfig, []string{"testdata/projects/configsecret"}, nil)
	assert.NoError(t, err)

	defer engine.Close()

	err = engine.Build(ctx)

	expectedErrs := []string{
		"config \"missingConfig\" is not provided in ftl-project.toml, but is required by module \"configsecret\"",
		"secret \"missingSecret\" is not provided in ftl-project.toml, but is required by module \"configsecret\"",
	}
	for _, expectedErr := range expectedErrs {
		assert.Contains(t, err.Error(), expectedErr)
	}
}
