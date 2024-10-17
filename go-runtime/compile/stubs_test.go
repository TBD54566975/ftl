package compile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/schema"
)

func setUp(t *testing.T) (ctx context.Context, projectRoot, fakeGoModDir string) {
	t.Helper()

	ctx = log.ContextWithNewDefaultLogger(context.Background())

	projectRoot = t.TempDir()
	fakeDir := filepath.Join(projectRoot, "fakegomod")
	err := os.MkdirAll(fakeDir, 0700)
	assert.NoError(t, err)
	ftlPath, err := filepath.Abs("../..")
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(fakeDir, "go.mod"), []byte(fmt.Sprintf(`
	module ftl/fake
	go 1.23.0

	replace github.com/TBD54566975/ftl => %s
	`, ftlPath)), 0600)
	// err := copy.Copy(filepath.Join("testdata", "go", "time"), filepath.Join(projectRoot, "time"))
	assert.NoError(t, err)
	return ctx, projectRoot, fakeDir
}

func TestGenerateGoStubs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	modules := []*schema.Module{
		schema.Builtins(),
		{Name: "other", Decls: []schema.Decl{
			&schema.Enum{
				Comments: []string{"This is an enum.", "", "It has 3 variants."},
				Name:     "Color",
				Export:   true,
				Type:     &schema.String{},
				Variants: []*schema.EnumVariant{
					{Name: "Red", Value: &schema.StringValue{Value: "Red"}},
					{Name: "Blue", Value: &schema.StringValue{Value: "Blue"}},
					{Name: "Green", Value: &schema.StringValue{Value: "Green"}},
				},
			},
			&schema.Enum{
				Name:   "ColorInt",
				Export: true,
				Type:   &schema.Int{},
				Variants: []*schema.EnumVariant{
					{Name: "RedInt", Value: &schema.IntValue{Value: 0}},
					{Name: "BlueInt", Value: &schema.IntValue{Value: 1}},
					{Name: "GreenInt", Value: &schema.IntValue{Value: 2}},
				},
			},
			&schema.Enum{
				Comments: []string{"This is type enum."},
				Name:     "TypeEnum",
				Export:   true,
				Variants: []*schema.EnumVariant{
					{Name: "A", Value: &schema.TypeValue{Value: &schema.Int{}}},
					{Name: "B", Value: &schema.TypeValue{Value: &schema.String{}}},
				},
			},
			&schema.Data{Name: "EchoRequest", Export: true},
			&schema.Data{
				Comments: []string{"This is an echo data response."},
				Name:     "EchoResponse", Export: true},
			&schema.Verb{
				Name:     "echo",
				Export:   true,
				Request:  &schema.Ref{Name: "EchoRequest"},
				Response: &schema.Ref{Name: "EchoResponse"},
			},
			&schema.Data{Name: "SinkReq", Export: true},
			&schema.Verb{
				Comments: []string{"This is a sink verb.", "", "Here is another line for this comment!"},
				Name:     "sink",
				Export:   true,
				Request:  &schema.Ref{Name: "SinkReq"},
				Response: &schema.Unit{},
			},
			&schema.Data{Name: "SourceResp", Export: true},
			&schema.Verb{
				Name:     "source",
				Export:   true,
				Request:  &schema.Unit{},
				Response: &schema.Ref{Name: "SourceResp"},
			},
			&schema.Verb{
				Name:     "nothing",
				Export:   true,
				Request:  &schema.Unit{},
				Response: &schema.Unit{},
			},
		}},
		{Name: "test"},
	}

	expected := `// Code generated by FTL. DO NOT EDIT.

package other

import (
  "context"

  "github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

var _ = context.Background

// This is an enum.
//
// It has 3 variants.
//
//ftl:enum
type Color string
const (
  Red Color = "Red"
  Blue Color = "Blue"
  Green Color = "Green"
)

//ftl:enum
type ColorInt int
const (
  RedInt ColorInt = 0
  BlueInt ColorInt = 1
  GreenInt ColorInt = 2
)

// This is type enum.
//
//ftl:enum
type TypeEnum interface { typeEnum() }

type A int

func (A) typeEnum() {}

type B string

func (B) typeEnum() {}

type EchoRequest struct {
}

// This is an echo data response.
//
type EchoResponse struct {
}

//ftl:verb
type EchoClient func(context.Context, EchoRequest) (EchoResponse, error)

type SinkReq struct {
}

// This is a sink verb.
//
// Here is another line for this comment!
//
//ftl:verb
type SinkClient func(context.Context, SinkReq)

type SourceResp struct {
}

//ftl:verb
type SourceClient func(context.Context) (SourceResp, error)

//ftl:verb
type NothingClient func(context.Context) error

func init() {
  reflection.Register(
    reflection.SumType[TypeEnum](
      *new(A),
      *new(B),
    ),
  )
}
`
	ctx, projectRoot, fakeDir := setUp(t)
	for _, module := range modules {
		path := filepath.Join(projectRoot, ".ftl", "go", "modules", module.Name)
		err := os.MkdirAll(path, 0700)
		assert.NoError(t, err)

		// Generate stubs needs a go.mod file to check for go version
		// Force each of these test modules to point to time's dir just so it can read the go.mod file
		config := moduleconfig.AbsModuleConfig{
			Language: "go",
			Dir:      fakeDir,
		}
		err = GenerateStubs(ctx, path, module, config, optional.None[moduleconfig.AbsModuleConfig]())
		assert.NoError(t, err)
	}

	generatedPath := filepath.Join(projectRoot, ".ftl/go/modules/other/external_module.go")
	fileContent, err := os.ReadFile(generatedPath)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(fileContent))
}

func TestMetadataImportsExcluded(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	modules := []*schema.Module{
		schema.Builtins(),
		{Name: "test", Decls: []schema.Decl{
			&schema.Data{
				Comments: []string{"Request data type."},
				Name:     "Req", Export: true},
			&schema.Data{Name: "Resp", Export: true},
			&schema.Verb{
				Comments: []string{"This is a verb."},
				Name:     "call",
				Export:   true,
				Request:  &schema.Ref{Name: "Req"},
				Response: &schema.Ref{Name: "Resp"},
				Metadata: []schema.Metadata{
					&schema.MetadataCalls{Calls: []*schema.Ref{{Name: "verb", Module: "other"}}},
				},
			},
		}},
	}

	expected := `// Code generated by FTL. DO NOT EDIT.

package test

import (
  "context"
)

var _ = context.Background

// Request data type.
//
type Req struct {
}

type Resp struct {
}

// This is a verb.
//
//ftl:verb
type CallClient func(context.Context, Req) (Resp, error)
`

	ctx, projectRoot, fakeDir := setUp(t)

	for _, module := range modules {
		path := filepath.Join(projectRoot, ".ftl", "go", "modules", module.Name)
		err := os.MkdirAll(path, 0700)
		assert.NoError(t, err)

		// Generate stubs needs a go.mod file to check for go version
		// Force each of these test modules to point to time's dir just so it can read the go.mod file
		config := moduleconfig.AbsModuleConfig{
			Language: "go",
			Dir:      fakeDir,
		}
		err = GenerateStubs(ctx, path, module, config, optional.None[moduleconfig.AbsModuleConfig]())
		assert.NoError(t, err)
	}

	generatedPath := filepath.Join(projectRoot, ".ftl/go/modules/test/external_module.go")
	fileContent, err := os.ReadFile(generatedPath)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(fileContent))
}