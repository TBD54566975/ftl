package routing

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

func TestRouting(t *testing.T) {
	events := schemaeventsource.NewUnattached()
	events.Publish(schemaeventsource.EventUpsert{
		Module: &schema.Module{
			Name: "time",
			Runtime: &schema.ModuleRuntime{
				Deployment: &schema.ModuleRuntimeDeployment{
					Endpoint:      "http://time.ftl",
					DeploymentKey: "dpl-time-sjkfislfjslfas",
				},
			},
		},
	})

	rt := New(log.ContextWithNewDefaultLogger(context.TODO()), events)
	current := rt.Current()
	assert.Equal(t, optional.Ptr(must.Get(url.Parse("http://time.ftl"))), current.GetForModule("time"))
	assert.Equal(t, optional.None[url.URL](), current.GetForModule("echo"))

	events.Publish(schemaeventsource.EventUpsert{
		Module: &schema.Module{
			Name: "echo",
			Runtime: &schema.ModuleRuntime{
				Deployment: &schema.ModuleRuntimeDeployment{
					Endpoint:      "http://echo.ftl",
					DeploymentKey: "dpl-echo-sjkfiaslfjslfs",
				},
			},
		},
	})

	time.Sleep(time.Millisecond * 250)
	current = rt.Current()
	assert.Equal(t, optional.Ptr(must.Get(url.Parse("http://echo.ftl"))), current.GetForModule("echo"))
}
