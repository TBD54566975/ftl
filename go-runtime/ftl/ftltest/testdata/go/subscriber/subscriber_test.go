package subscriber

import (
	"ftl/pubsub"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
)

func TestPublishToExternalModule(t *testing.T) {
	ctx := ftltest.Context(t)
	assert.NoError(t, pubsub.Topic.Publish(ctx, pubsub.Event{Value: "external"}))
	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, pubsub.Topic)))

	// Make sure we correctly made the right ref for the external module.
	assert.Equal(t, "pubsub", pubsub.Topic.Ref.Module)
}
