package buildengine

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestGoBuildClearsBuildDir(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "test"},
		},
	}
	bctx := buildContext{
		moduleDir: "testdata/another",
		buildDir:  ".ftl",
		sch:       sch,
	}
	testBuildClearsBuildDir(t, bctx)
}

func TestExternalType(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	bctx := buildContext{
		moduleDir: "testdata/external",
		buildDir:  ".ftl",
		sch:       &schema.Schema{},
	}
	testBuild(t, bctx, "", "unsupported external type", []assertion{
		assertBuildProtoErrors(
			"unsupported external type \"time.Month\"",
			"unsupported type \"time.Month\" for field \"Month\"",
			"unsupported response type \"ftl/external.ExternalResponse\"",
		),
	})
}

func TestGoModVersion(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "highgoversion", Decls: []schema.Decl{
				&schema.Data{Name: "EchoReq"},
				&schema.Data{Name: "EchoResp"},
				&schema.Verb{
					Name:     "echo",
					Request:  &schema.Ref{Name: "EchoRequest"},
					Response: &schema.Ref{Name: "EchoResponse"},
				},
			}},
		},
	}
	bctx := buildContext{
		moduleDir: "testdata/highgoversion",
		buildDir:  ".ftl",
		sch:       sch,
	}
	testBuild(t, bctx, fmt.Sprintf("go version %q is not recent enough for this module, needs minimum version \"9000.1.1\"", runtime.Version()[2:]), "", []assertion{})
}

func TestGeneratedTypeRegistry(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			{Name: "another", Decls: []schema.Decl{
				&schema.Enum{
					Name:   "TypeEnum",
					Export: true,
					Variants: []*schema.EnumVariant{
						{Name: "A", Value: &schema.TypeValue{Value: &schema.Int{}}},
						{Name: "B", Value: &schema.TypeValue{Value: &schema.String{}}},
					},
				},
				&schema.Enum{
					Name:   "SecondTypeEnum",
					Export: true,
					Variants: []*schema.EnumVariant{
						{Name: "One", Value: &schema.TypeValue{Value: &schema.Int{}}},
						{Name: "Two", Value: &schema.TypeValue{Value: &schema.String{}}},
					},
				},
				&schema.Data{
					Name:   "TransitiveTypeEnum",
					Export: true,
					Fields: []*schema.Field{
						{Name: "TypeEnumRef", Type: &schema.Ref{Name: "SecondTypeEnum", Module: "another"}},
					},
				},
			}},
		},
	}
	expected, err := os.ReadFile("testdata/type_registry_main.go")
	assert.NoError(t, err)
	bctx := buildContext{
		moduleDir: "testdata/other",
		buildDir:  ".ftl",
		sch:       sch,
	}
	testBuild(t, bctx, "", "", []assertion{
		assertGeneratedMain(string(expected)),
	})
}
