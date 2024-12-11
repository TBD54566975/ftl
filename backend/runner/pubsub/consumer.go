package pubsub

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"

	"github.com/TBD54566975/ftl/backend/controller/observability"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type consumer struct {
	moduleName string
	deployment model.DeploymentKey
	verb       *schema.Verb
	subscriber *schema.MetadataSubscriber

	client VerbClient
}

func newConsumer(moduleName string, verb *schema.Verb, subscriber *schema.MetadataSubscriber, deployment model.DeploymentKey, client VerbClient) (*consumer, error) {
	if verb.Runtime == nil {
		return nil, fmt.Errorf("subscription %s has no runtime", verb.Name)
	}
	if len(verb.Runtime.Subscription.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("subscription %s has no Kafka brokers", verb.Name)
	}

	c := &consumer{
		moduleName: moduleName,
		deployment: deployment,
		verb:       verb,
		subscriber: subscriber,
		client:     client,
	}

	return c, nil
}

func kafkaConsumerGroupID(moduleName string, verb *schema.Verb) string {
	return schema.RefKey{Module: moduleName, Name: verb.Name}.String()
}

func (c *consumer) kafkaTopicID() string {
	return c.subscriber.Topic.String()
}

func (c *consumer) Begin(ctx context.Context) error {
	// set up config
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	switch c.subscriber.FromOffset {
	case schema.FromOffsetBeginning, schema.FromOffsetUnspecified:
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case schema.FromOffsetLatest:
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true

	group, err := sarama.NewConsumerGroup(c.verb.Runtime.Subscription.KafkaBrokers, kafkaConsumerGroupID(c.moduleName, c.verb), config)
	if err != nil {
		return fmt.Errorf("failed to create consumer group for subscription %s: %w", c.verb.Name, err)
	}
	go c.watchErrors(ctx, group)
	go c.subscribe(ctx, group)
	return nil
}

func (c *consumer) watchErrors(ctx context.Context, group sarama.ConsumerGroup) {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-group.Errors():
			logger.Errorf(err, "error in consumer group %s", c.verb.Name)
		}
	}
}

func (c *consumer) subscribe(ctx context.Context, group sarama.ConsumerGroup) {
	logger := log.FromContext(ctx)
	defer group.Close()
	// Iterate over consumer sessions.
	//
	// `Consume` should be called inside an infinite loop, when a
	// server-side rebalance happens, the consumer session will need to be
	// recreated to get the new claims.
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := group.Consume(ctx, []string{c.kafkaTopicID()}, c)
		if err != nil {
			logger.Errorf(err, "consume session failed for %s", c.verb.Name)
		} else {
			logger.Debugf("Ending consume session for subscription %s", c.verb.Name)
		}
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	logger := log.FromContext(session.Context())

	partitions := session.Claims()[kafkaConsumerGroupID(c.moduleName, c.verb)]
	logger.Debugf("Starting consume session for subscription %s with partitions: [%v]", c.verb.Name, strings.Join(slices.Map(partitions, func(partition int32) string { return strconv.Itoa(int(partition)) }), ","))

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	logger := log.FromContext(ctx)
	for msg := range claim.Messages() {
		start := time.Now()

		// TODO: request id, what should it look like?
		requestKey := model.NewRequestKey(model.OriginPubsub, c.verb.Name)
		logger.Debugf("%s: consuming message (%v:%v): %v", c.verb.Name, msg.Partition, msg.Offset, string(msg.Value))
		destRef := &schema.Ref{
			Module: c.moduleName,
			Name:   c.verb.Name,
		}
		req := &ftlv1.CallRequest{
			Verb: schema.RefKey{Module: c.moduleName, Name: c.verb.Name}.ToProto(),
			Body: msg.Value,
		}
		consumeEvent := timeline.PubSubConsume{
			DeploymentKey: c.deployment,
			RequestKey:    optional.Some(requestKey.String()),
			Time:          time.Now(),
			DestVerb:      optional.Some(destRef.ToRefKey()),
			Topic:         c.subscriber.Topic.String(),
		}
		callEvent := &timeline.Call{
			DeploymentKey: c.deployment,
			RequestKey:    requestKey,
			StartTime:     start,
			DestVerb:      destRef,
			Callers:       []*schema.Ref{},
			Request:       req,
		}

		resp, err := c.client.Call(session.Context(), connect.NewRequest(req))
		if err != nil {
			consumeEvent.Error = optional.Some(err.Error())
			callEvent.Response = result.Err[*ftlv1.CallResponse](err)
			observability.Calls.Request(ctx, req.Verb, start, optional.Some("verb call failed"))
		} else {
			callEvent.Response = result.Ok(resp.Msg)
			observability.Calls.Request(ctx, req.Verb, start, optional.None[string]())
		}
		session.MarkMessage(msg, "")
		timeline.ClientFromContext(ctx).Publish(ctx, consumeEvent)
		timeline.ClientFromContext(ctx).Publish(ctx, callEvent)
	}
	return nil
}
