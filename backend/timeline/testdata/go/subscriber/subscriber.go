package subscriber

import (
	"context"
	"ftl/publisher"

	"github.com/block/ftl/go-runtime/ftl"
)

//ftl:verb
//ftl:subscribe publisher.testTopic from=beginning
func Consume(ctx context.Context, req publisher.PubSubEvent) error {
	ftl.LoggerFromContext(ctx).Infof("Subscriber is consuming %v", req.Time)
	return nil
}
