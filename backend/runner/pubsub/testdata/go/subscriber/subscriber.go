package subscriber

import (
	"context"
	"fmt"
	"time"

	"ftl/publisher"

	"github.com/TBD54566975/ftl/common/reflection"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:verb
//ftl:subscribe publisher.testTopic from=beginning
func Consume(ctx context.Context, req publisher.PubSubEvent) error {
	ftl.LoggerFromContext(ctx).Infof("Consume: %v", req.Time)
	return nil
}

//ftl:verb
//ftl:subscribe publisher.testTopic from=latest
func ConsumeFromLatest(ctx context.Context, req publisher.PubSubEvent) error {
	ftl.LoggerFromContext(ctx).Infof("ConsumeFromLatest: %v", req.Time)
	return nil
}

//ftl:verb
//ftl:subscribe publisher.topic2 from=beginning
//ftl:retry 2 1s 1s
func ConsumeButFailAndRetry(ctx context.Context, req publisher.PubSubEvent) error {
	return fmt.Errorf("always error: event %v", req.Time)
}

//ftl:verb
func PublishToExternalModule(ctx context.Context) error {
	// Get around compile-time checks
	externalTopic := ftl.TopicHandle[publisher.PubSubEvent, ftl.SinglePartitionMap[publisher.PubSubEvent]]{Ref: reflection.Ref{Module: "publisher", Name: "testTopic"}.ToSchema()}
	return externalTopic.Publish(ctx, publisher.PubSubEvent{Time: time.Now()})
}
