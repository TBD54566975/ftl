package buildengine_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

func TestGraph(t *testing.T) {
	ctx, cancel := context.WithCancel(log.ContextWithNewDefaultLogger(context.Background()))
	t.Cleanup(cancel)
	projConfig := projectconfig.Config{
		Path: filepath.Join(t.TempDir(), "ftl-project.toml"),
		Name: "test",
	}
	engine, err := buildengine.New(ctx, nil, schemaeventsource.NewUnattached(), projConfig, []string{"testdata/alpha", "testdata/other", "testdata/another"})
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
