package buildengine_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestEngine(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	engine, err := buildengine.New(ctx, nil, []string{"testdata/projects/alpha", "testdata/projects/other", "testdata/projects/another"}, nil)
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

func TestDissallowMultipleEngines(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	engine, err := buildengine.New(ctx, nil, []string{"testdata/projects/alpha"}, nil)
	assert.NoError(t, err)
	defer engine.Close()

	_, err = buildengine.New(ctx, nil, []string{"testdata/projects/alpha"}, nil)
	assert.Error(t, err)
}
