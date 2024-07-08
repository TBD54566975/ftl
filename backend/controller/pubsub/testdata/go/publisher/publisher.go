package publisher

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

//ftl:export
var Test_topic = ftl.Topic[PubSubEvent]("test_topic")

type PubSubEvent struct {
	Time time.Time
}

//ftl:verb
func PublishTen(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	for i := 0; i < 10; i++ {
		t := time.Now()
		logger.Infof("Publishing %v", t)
		err := Test_topic.Publish(ctx, PubSubEvent{Time: t})
		if err != nil {
			return err
		}
	}
	return nil
}

//ftl:verb
func PublishOne(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	t := time.Now()
	logger.Infof("Publishing %v", t)
	return Test_topic.Publish(ctx, PubSubEvent{Time: t})
}

//ftl:export
var Topic2 = ftl.Topic[PubSubEvent]("topic2")

//ftl:verb
func PublishOneToTopic2(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	t := time.Now()
	logger.Infof("Publishing to topic2 %v", t)
	return Topic2.Publish(ctx, PubSubEvent{Time: t})
}
