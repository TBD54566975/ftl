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
func PublishTen(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	for i := 0; i < 10; i++ {
		t := time.Now()
		logger.Infof("Publishing %v", t)
		err := topic.Publish(ctx, PubSubEvent{Time: t})
		if err != nil {
			return err
		}
		time.Sleep(time.Microsecond * 20)
	}
	return nil
}

//ftl:verb
func PublishOne(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	for i := 0; i < 10; i++ {
		t := time.Now()
		logger.Infof("Publishing %v", t)
		err := topic.Publish(ctx, PubSubEvent{Time: t})
		if err != nil {
			return err
		}
		time.Sleep(time.Microsecond * 20)
	}
	return nil
}
