package pubsub

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/alecthomas/types/optional"
)

type consumer struct {
	moduleName string
	deployment model.DeploymentKey
	verb       *schema.Verb
	subscriber *schema.MetadataSubscriber

	client VerbClient
}

func newConsumer(ctx context.Context, moduleName string, verb *schema.Verb, subscriber *schema.MetadataSubscriber, deployment model.DeploymentKey, client VerbClient) (*consumer, error) {
	logger := log.FromContext(ctx)
	if verb.Runtime == nil {
		return nil, fmt.Errorf("subscription %s has no runtime", verb.Name)
	}
	if len(verb.Runtime.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("subscription %s has no Kafka brokers", verb.Name)
	}

	// set up config
	// TODO: any config needed? autocommit defaults to on. What do we want?
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	// TODO: reconsider default value
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true

	group, err := sarama.NewConsumerGroup(verb.Runtime.KafkaBrokers, kafkaConsumerGroupID(moduleName, verb), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group for subscription %s: %w", verb.Name, err)
	}
	// TODO: close!
	// defer func() { _ = group.Close() }()

	sub := &consumer{
		moduleName: moduleName,
		deployment: deployment,
		verb:       verb,
		subscriber: subscriber,
		client:     client,
	}

	go func() {
		for err := range group.Errors() {
			// TODO: handle?
			logger.Errorf(err, "consumer group error for subscription %s", verb.Name)
		}
	}()

	go sub.subscribe(ctx, group)

	return sub, nil
}

func kafkaConsumerGroupID(moduleName string, verb *schema.Verb) string {
	return schema.RefKey{Module: moduleName, Name: verb.Name}.String()
}

func (c *consumer) kafkaTopicID() string {
	return c.subscriber.Topic.String()
}

func (c *consumer) subscribe(ctx context.Context, group sarama.ConsumerGroup) {
	// Iterate over consumer sessions.
	// ctx := context.Background()
	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		err := group.Consume(ctx, []string{c.kafkaTopicID()}, c)
		if err != nil {
			panic("could not consume: " + err.Error())
		}
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (s *consumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (s *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	// TODO: any last commits or is that automatic?
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	logger := log.FromContext(ctx)
	for msg := range claim.Messages() {
		// TODO: request id, what should it look like?
		requestKey := model.NewRequestKey(model.OriginPubsub, c.verb.Name)
		logger.Debugf("%s: consuming message: %v", c.verb.Name, string(msg.Value))
		timelineEvent := timeline.PubSubConsume{
			DeploymentKey: c.deployment,
			RequestKey:    optional.Some(requestKey.String()),
			Time:          time.Now(),
			DestVerb: optional.Some(schema.RefKey{
				Module: c.moduleName,
				Name:   c.verb.Name,
			}),
			Topic: c.subscriber.Topic.String(),
		}

		_, err := c.client.Call(session.Context(), connect.NewRequest(&ftlv1.CallRequest{
			Verb: schema.RefKey{Module: c.moduleName, Name: c.verb.Name}.ToProto(),
			Body: msg.Value,
		}))
		if err != nil {
			timelineEvent.Error = optional.Some(err.Error())
		}
		timeline.ClientFromContext(ctx).Publish(ctx, timelineEvent)
	}
	return nil
}
