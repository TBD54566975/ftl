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
	Time     time.Time
	Haystack string
}

//ftl:verb
func PublishTen(ctx context.Context, topic TestTopic) error {
	logger := ftl.LoggerFromContext(ctx)
	for i := 0; i < 10; i++ {
		t := time.Now()
		logger.Infof("Publishing %v", t)
		err := topic.Publish(ctx, PubSubEvent{Time: t})
		if err != nil {
			return err
		}
	}
	return nil
}

//ftl:verb
func PublishOne(ctx context.Context, topic TestTopic) error {
	logger := ftl.LoggerFromContext(ctx)
	t := time.Now()
	logger.Infof("Publishing %v", t)
	return topic.Publish(ctx, PubSubEvent{Time: t})
}

//ftl:export
type Topic2 = ftl.TopicHandle[PubSubEvent, ftl.SinglePartitionMap[PubSubEvent]]

//ftl:data
type PublishOneToTopic2Request struct {
	Haystack string
}

//ftl:verb
func PublishOneToTopic2(ctx context.Context, req PublishOneToTopic2Request, topic Topic2) error {
	logger := ftl.LoggerFromContext(ctx)
	t := time.Now()
	logger.Infof("Publishing to topic_2 %v", t)
	return topic.Publish(ctx, PubSubEvent{
		Time:     t,
		Haystack: req.Haystack,
	})
}
