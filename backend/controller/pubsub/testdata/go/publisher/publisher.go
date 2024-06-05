package publisher

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

//ftl:export
var topic = ftl.Topic[PubSubEvent]("test_topic")

type PubSubEvent struct {
	Time time.Time
}

//ftl:verb
func Publish(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	t := time.Now()
	logger.Infof("Publishing %v", t)
	return topic.Publish(ctx, PubSubEvent{Time: t})
}
