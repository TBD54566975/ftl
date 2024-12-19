package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/jpillora/backoff"

	"github.com/block/ftl/backend/controller/observability"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/model"
)

type consumer struct {
	moduleName          string
	deployment          model.DeploymentKey
	verb                *schema.Verb
	subscriber          *schema.MetadataSubscriber
	retryParams         schema.RetryParams
	group               sarama.ConsumerGroup
	deadLetterPublisher optional.Option[*publisher]

	verbClient     VerbClient
	timelineClient *timeline.Client
}

func newConsumer(moduleName string, verb *schema.Verb, subscriber *schema.MetadataSubscriber, deployment model.DeploymentKey,
	deadLetterPublisher optional.Option[*publisher], verbClient VerbClient, timelineClient *timeline.Client) (*consumer, error) {
	if verb.Runtime == nil {
		return nil, fmt.Errorf("subscription %s has no runtime", verb.Name)
	}
	if len(verb.Runtime.Subscription.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("subscription %s has no Kafka brokers", verb.Name)
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = true
	switch subscriber.FromOffset {
	case schema.FromOffsetBeginning, schema.FromOffsetUnspecified:
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case schema.FromOffsetLatest:
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	groupID := kafkaConsumerGroupID(moduleName, verb)
	group, err := sarama.NewConsumerGroup(verb.Runtime.Subscription.KafkaBrokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group for subscription %s: %w", verb.Name, err)
	}

	c := &consumer{
		moduleName:          moduleName,
		deployment:          deployment,
		verb:                verb,
		subscriber:          subscriber,
		group:               group,
		deadLetterPublisher: deadLetterPublisher,

		verbClient:     verbClient,
		timelineClient: timelineClient,
	}
	retryMetada, ok := slices.FindVariant[*schema.MetadataRetry](verb.Metadata)
	if ok {
		retryParams, err := retryMetada.RetryParams()
		if err != nil {
			return nil, fmt.Errorf("failed to parse retry params for subscription %s: %w", verb.Name, err)
		}
		c.retryParams = retryParams
	} else {
		c.retryParams = schema.RetryParams{}
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
	logger := log.FromContext(ctx).AppendScope("sub:" + c.verb.Name)
	logger.Debugf("Subscribing to %s", c.kafkaTopicID())

	go c.watchErrors(ctx)
	go c.subscribe(ctx)
	return nil
}

func (c *consumer) watchErrors(ctx context.Context) {
	logger := log.FromContext(ctx).AppendScope("sub:" + c.verb.Name)
	for err := range channels.IterContext(ctx, c.group.Errors()) {
		logger.Errorf(err, "Consumer group error")
	}
}

func (c *consumer) subscribe(ctx context.Context) {
	logger := log.FromContext(ctx).AppendScope("sub:" + c.verb.Name)
	// Iterate over consumer sessions.
	//
	// `Consume` should be called inside an infinite loop, when a server-side rebalance happens,
	// the consumer session will need to be recreated to get the new claims.
	for {
		select {
		case <-ctx.Done():
			c.group.Close()
			return
		default:
		}

		err := c.group.Consume(ctx, []string{c.kafkaTopicID()}, c)
		if err != nil {
			logger.Errorf(err, "Session failed for %s", c.verb.Name)
		} else {
			logger.Debugf("Ending session")
		}
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	logger := log.FromContext(session.Context()).AppendScope("sub:" + c.verb.Name)

	partitions := session.Claims()[c.kafkaTopicID()]
	logger.Debugf("Starting session with partitions [%v]", strings.Join(slices.Map(partitions, func(partition int32) string { return strconv.Itoa(int(partition)) }), ","))

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited but before the
// offsets are committed for the very last time.
func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages(). Once the Messages() channel
// is closed, the Handler must finish its processing loop and exit.
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	logger := log.FromContext(ctx).AppendScope("sub:" + c.verb.Name)

	for msg := range channels.IterContext(ctx, claim.Messages()) {
		if msg == nil {
			// Channel closed, rebalance or shutdown needed
			return nil
		}
		logger.Debugf("Consuming message with partition %v and offset %v", msg.Partition, msg.Offset)
		remainingRetries := c.retryParams.Count
		backoff := c.retryParams.MinBackoff
		for {
			err := c.call(ctx, msg.Value, int(msg.Partition), int(msg.Offset))
			if err == nil {
				break
			}
			select {
			case <-ctx.Done():
				// Do not commit the message if we did not succeed and the context is done.
				// No need to retry message either.
				logger.Errorf(err, "Failed to consume message with partition %v and offset %v", msg.Partition, msg.Offset)
				return nil
			default:
			}
			if remainingRetries == 0 {
				logger.Errorf(err, "Failed to consume message with partition %v and offset %v", msg.Partition, msg.Offset)
				if !c.publishToDeadLetterTopic(ctx, msg, err) {
					return nil
				}
				break
			}
			logger.Errorf(err, "Failed to consume message with partition %v and offset %v and will retry in %vs", msg.Partition, msg.Offset, int(backoff.Seconds()))
			time.Sleep(backoff)
			remainingRetries--
			backoff *= 2
			if backoff > c.retryParams.MaxBackoff {
				backoff = c.retryParams.MaxBackoff
			}
		}
		session.MarkMessage(msg, "")
	}
	// Rebalance or shutdown needed
	return nil
}

func (c *consumer) call(ctx context.Context, body []byte, partition, offset int) error {
	start := time.Now()

	requestKey := model.NewRequestKey(model.OriginPubsub, schema.RefKey{Module: c.moduleName, Name: c.verb.Name}.String())
	destRef := &schema.Ref{
		Module: c.moduleName,
		Name:   c.verb.Name,
	}
	req := &ftlv1.CallRequest{
		Verb: schema.RefKey{Module: c.moduleName, Name: c.verb.Name}.ToProto(),
		Body: body,
	}
	consumeEvent := timeline.PubSubConsume{
		DeploymentKey: c.deployment,
		RequestKey:    optional.Some(requestKey.String()),
		Time:          time.Now(),
		DestVerb:      optional.Some(destRef.ToRefKey()),
		Topic:         c.subscriber.Topic.String(),
		Partition:     partition,
		Offset:        offset,
	}
	defer c.timelineClient.Publish(ctx, consumeEvent)

	callEvent := &timeline.Call{
		DeploymentKey: c.deployment,
		RequestKey:    requestKey,
		StartTime:     start,
		DestVerb:      destRef,
		Callers:       []*schema.Ref{},
		Request:       req,
	}
	defer c.timelineClient.Publish(ctx, callEvent)

	resp, callErr := c.verbClient.Call(ctx, connect.NewRequest(req))
	if callErr == nil {
		if errResp, ok := resp.Msg.Response.(*ftlv1.CallResponse_Error_); ok {
			callErr = fmt.Errorf("verb call failed: %s", errResp.Error.Message)
		}
	}
	if callErr != nil {
		consumeEvent.Error = optional.Some(callErr.Error())
		callEvent.Response = result.Err[*ftlv1.CallResponse](callErr)
		observability.Calls.Request(ctx, req.Verb, start, optional.Some("verb call failed"))
		return callErr
	}
	callEvent.Response = result.Ok(resp.Msg)
	observability.Calls.Request(ctx, req.Verb, start, optional.None[string]())
	return nil
}

// publishToDeadLetterTopic tries to publish the message to the dead letter topic.
//
// If it does not succeed it will retry until it succeeds or the context is done.
// Returns true if the message was published or if there is no dead letter queue.
// Returns false if the context is done.
func (c *consumer) publishToDeadLetterTopic(ctx context.Context, msg *sarama.ConsumerMessage, callErr error) bool {
	p, ok := c.deadLetterPublisher.Get()
	if !ok {
		return true
	}

	deadLetterEvent, err := encoding.Marshal(map[string]any{
		"event": json.RawMessage(msg.Value),
		"error": callErr.Error(),
	})
	if err != nil {
		panic(fmt.Errorf("failed to marshal dead letter event for %v on partition %v and offset %v: %w", c.kafkaTopicID(), msg.Partition, msg.Offset, err))
	}

	bo := &backoff.Backoff{Min: time.Second, Max: 10 * time.Second}
	first := true
	for {
		var waitDuration time.Duration
		if first {
			first = false
		} else {
			waitDuration = bo.Duration()
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(waitDuration):
		}
		err := p.publish(ctx, deadLetterEvent, string(msg.Key), schema.Ref{Module: c.moduleName, Name: c.verb.Name})
		if err == nil {
			return true
		}
	}
}
