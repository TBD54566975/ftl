package publisher

import (
	"context"
	"time"

	"github.com/block/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

//ftl:export
type TestTopic = ftl.TopicHandle[PubSubEvent, ftl.SinglePartitionMap[PubSubEvent]]

type PubSubEvent struct {
	Time time.Time
}

//ftl:verb
func Publish(ctx context.Context, topic TestTopic) error {
	logger := ftl.LoggerFromContext(ctx)
	t := time.Now()
	logger.Infof("Publishing %v", t)
	return topic.Publish(ctx, PubSubEvent{Time: t})
}
