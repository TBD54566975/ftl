//go:build integration

package goplugin

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestGoBuildClearsBuildDir(t *testing.T) {
	file := "./another/.ftl/test-clear-build.tmp"
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("another"),
		in.Build("another"),
		in.WriteFile(file, []byte{1}),
		in.FileExists(file),
		in.Build("another"),
		in.ExpectError(in.FileExists(file), "no such file"),
	)
}

func TestExternalType(t *testing.T) {
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("external"),
		in.ExpectError(in.Build("external"),
			`unsupported type "time.Month" for field "Month"`,
			`unsupported external type "time.Month"; see FTL docs on using external types: block.github.io/ftl/docs/reference/externaltypes/`,
			`unsupported response type "ftl/external.ExternalResponse"`,
		),
	)
}

func TestGeneratedTypeRegistry(t *testing.T) {
	t.Skip("Skipping until there has been a release with the package change")
	expected, err := os.ReadFile("testdata/type_registry_main.go")
	assert.NoError(t, err)

	file := "other/.ftl/go/main/main.go"

	in.Run(t,
		in.WithTestDataDir("testdata"),
		// Deploy dependency
		in.CopyModule("another"),
		in.Deploy("another"),
		// Build the module under test
		in.CopyModule("other"),
		in.ExpectError(in.FileExists(file), "no such file"),
		in.Build("other"),
		// Validate the generated main.go
		in.FileContent(file, string(expected)),
	)
}
