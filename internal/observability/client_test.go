package observability

import (
	"path"
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestSchemaMismatch(t *testing.T) {
	dflt := resource.Default()
	assert.Equal(t, dflt.SchemaURL(), schemaURL, `change import in client.go to: semconv "go.opentelemetry.io/otel/semconv/v%s"`, path.Base(dflt.SchemaURL()))
}
