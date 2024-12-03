package pubsub

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

type subscription struct {
	moduleName string
	verb       *schema.Verb
	subscriber *schema.MetadataSubscriber

	client ftlv1connect.VerbServiceClient
}

func newSubscription(ctx context.Context, moduleName string, verb *schema.Verb, subscriber *schema.MetadataSubscriber, client ftlv1connect.VerbServiceClient) (*subscription, error) {
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

	sub := &subscription{
		moduleName: moduleName,
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

func (s *subscription) kafkaTopicID() string {
	return s.subscriber.Topic.String()
}

func (s *subscription) subscribe(ctx context.Context, group sarama.ConsumerGroup) {
	// Iterate over consumer sessions.
	// ctx := context.Background()
	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		err := group.Consume(ctx, []string{s.kafkaTopicID()}, s)
		if err != nil {
			panic("could not consume: " + err.Error())
		}
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (s *subscription) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (s *subscription) Cleanup(session sarama.ConsumerGroupSession) error {
	// TODO: any last commits or is that automatic?
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (s *subscription) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger := log.FromContext(session.Context())
	for msg := range claim.Messages() {
		logger.Debugf("%s: consumed message: %v", s.verb.Name, string(msg.Value))
		s.client.Call(session.Context(), connect.NewRequest(&ftlv1.CallRequest{
			Verb: schema.RefKey{Module: s.moduleName, Name: s.verb.Name}.ToProto(),
			Body: msg.Value,
		}))
	}
	return nil
}
