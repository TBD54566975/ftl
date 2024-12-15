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
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/common/slices"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

type consumer struct {
	moduleName  string
	deployment  model.DeploymentKey
	verb        *schema.Verb
	subscriber  *schema.MetadataSubscriber
	retryParams schema.RetryParams

	verbClient     VerbClient
	timelineClient *timeline.Client
}

func newConsumer(moduleName string, verb *schema.Verb, subscriber *schema.MetadataSubscriber, deployment model.DeploymentKey, verbClient VerbClient, timelineClient *timeline.Client) (*consumer, error) {
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
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = true
	switch c.subscriber.FromOffset {
	case schema.FromOffsetBeginning, schema.FromOffsetUnspecified:
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case schema.FromOffsetLatest:
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
	groupID := kafkaConsumerGroupID(c.moduleName, c.verb)
	log.FromContext(ctx).Infof("Subscribing to topic %s for %s with offset %v", c.kafkaTopicID(), groupID, config.Consumer.Offsets.Initial)

	group, err := sarama.NewConsumerGroup(c.verb.Runtime.Subscription.KafkaBrokers, groupID, config)
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
	// `Consume` should be called inside an infinite loop, when a server-side rebalance happens,
	// the consumer session will need to be recreated to get the new claims.
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

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited but before the
// offsets are committed for the very last time.
func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages(). Once the Messages() channel
// is closed, the Handler must finish its processing loop and exit.
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	logger := log.FromContext(ctx)
	for msg := range claim.Messages() {
		logger.Debugf("%s: consuming message %v:%v", c.verb.Name, msg.Partition, msg.Offset)
		remainingRetries := c.retryParams.Count
		backoff := c.retryParams.MinBackoff
		for {
			err := c.call(ctx, msg.Value, int(msg.Partition), int(msg.Offset))
			if err == nil {
				break
			}
			if remainingRetries == 0 {
				logger.Errorf(err, "%s: failed to consume message %v:%v", c.verb.Name, msg.Partition, msg.Offset)
				break
			}
			logger.Errorf(err, "%s: failed to consume message %v:%v: retrying in %vs", c.verb.Name, msg.Partition, msg.Offset, int(backoff.Seconds()))
			time.Sleep(backoff)
			remainingRetries--
			backoff *= 2
			if backoff > c.retryParams.MaxBackoff {
				backoff = c.retryParams.MaxBackoff
			}
		}
		session.MarkMessage(msg, "")
	}
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
