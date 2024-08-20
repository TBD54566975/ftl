package observability

import (
	"path"
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestSchemaMismatch(t *testing.T) {
	dflt := resource.Default()
	assert.Equal(t, dflt.SchemaURL(), schemaURL, `in every file that imports go.opentelemetry.io/otel/semconv, change the import to: semconv "go.opentelemetry.io/otel/semconv/v%s"`, path.Base(dflt.SchemaURL()))
}
