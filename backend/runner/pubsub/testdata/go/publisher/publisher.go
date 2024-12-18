package publisher

import (
	"context"
	"time"

	"github.com/block/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

type PartitionMapper struct{}

var _ ftl.TopicPartitionMap[PubSubEvent] = PartitionMapper{}

func (PartitionMapper) PartitionKey(event PubSubEvent) string {
	return event.Time.String()
}

//ftl:export
type TestTopic = ftl.TopicHandle[PubSubEvent, PartitionMapper]

type LocalTopic = ftl.TopicHandle[PubSubEvent, PartitionMapper]

type PubSubEvent struct {
	Time     time.Time
	Haystack string
}

//ftl:verb
func PublishTen(ctx context.Context, topic TestTopic) error {
	logger := ftl.LoggerFromContext(ctx)
	for i := 0; i < 10; i++ {
		t := time.Now()
		logger.Infof("Publishing to testTopic: %v", t)
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

//ftl:verb
func PublishTenLocal(ctx context.Context, topic LocalTopic) error {
	logger := ftl.LoggerFromContext(ctx)
	for i := 0; i < 10; i++ {
		t := time.Now()
		logger.Infof("Publishing to localTopic: %v", t)
		err := topic.Publish(ctx, PubSubEvent{Time: t})
		if err != nil {
			return err
		}
	}
	return nil
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

//ftl:verb
//ftl:subscribe localTopic from=latest
func Local(ctx context.Context, event PubSubEvent) error {
	ftl.LoggerFromContext(ctx).Infof("Consume local: %v", event.Time)
	return nil
}
