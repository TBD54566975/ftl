package subscriber

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ftl/builtin"
	"ftl/publisher"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
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
//ftl:subscribe publisher.topic2 from=beginning deadletter
//ftl:retry 2 1s 1s
func ConsumeButFailAndRetry(ctx context.Context, req publisher.PubSubEvent) error {
	return fmt.Errorf("always error: event %v", req.Time)
}

//ftl:verb
//ftl:subscribe consumeButFailAndRetryFailed from=beginning
func ConsumeFromDeadLetter(ctx context.Context, req builtin.FailedEvent[publisher.PubSubEvent]) error {
	ftl.LoggerFromContext(ctx).Infof("ConsumeFromDeadLetter: %v", req.Event.Time)
	return nil
}

//ftl:verb
func PublishToExternalModule(ctx context.Context) error {
	// Get around compile-time checks
	externalTopic := ftl.TopicHandle[publisher.PubSubEvent, ftl.SinglePartitionMap[publisher.PubSubEvent]]{Ref: reflection.Ref{Module: "publisher", Name: "testTopic"}.ToSchema()}
	return externalTopic.Publish(ctx, publisher.PubSubEvent{Time: time.Now()})
}

//ftl:verb
//ftl:subscribe publisher.slowTopic from=beginning
func ConsumeSlow(ctx context.Context, req publisher.PubSubEvent) error {
	versionDescription := "This deployment is TheFirstDeployment"
	if strings.Contains(versionDescription, "TheFirstDeployment") {
		ftl.LoggerFromContext(ctx).Infof("ConsumeSlow first deployment (will sleep 5s): %v", req.Time)
		time.Sleep(5 * time.Second)
		return nil
	}
	ftl.LoggerFromContext(ctx).Infof("ConsumeSlow second deployment (immediate): %v", req.Time)
	return nil
}
