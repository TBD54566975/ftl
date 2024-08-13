package subscriber_test

import (
	"ftl/pubsub"
	"testing"

	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/alecthomas/assert/v2"
)

func TestPublishToExternalModule(t *testing.T) {
	ctx := ftltest.Context()
	assert.NoError(t, pubsub.Topic1.Publish(ctx, pubsub.Event{Value: "external"}))
	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, pubsub.Topic1)))

	// Make sure we correctly made the right ref for the external module.
	assert.Equal(t, "pubsub", pubsub.Topic1.Ref.Module)
}
