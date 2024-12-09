package pubsub

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"

	ftlpubsubv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	ftlv1pubsubconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/pubsub/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

type publisher struct {
	module   string
	topic    *schema.Topic
	producer sarama.SyncProducer
}

func newPublisher(module string, t *schema.Topic) (*publisher, error) {
	if t.Runtime == nil {
		return nil, fmt.Errorf("topic %s has no runtime", t.Name)
	}
	if len(t.Runtime.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("topic %s has no Kafka brokers", t.Name)
	}
	if t.Runtime.TopicID == "" {
		return nil, fmt.Errorf("topic %s has no topic ID", t.Name)
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewSyncProducer(t.Runtime.KafkaBrokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer for topic %s: %w", t.Name, err)
	}
	return &publisher{
		module:   module,
		topic:    t,
		producer: producer,
	}, nil
}

func (p *publisher) publish(ctx context.Context, data []byte, key string, caller schema.RefKey) error {
	if err := p.publishToController(ctx, data, caller); err != nil {
		return err
	}
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic.Runtime.TopicID,
		Value: sarama.ByteEncoder(data),
		Key:   sarama.StringEncoder(key),
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// publishToController publishes the data to the controller (old pubsub implementation)
//
// This is to keep pubsub working while we transition fully to Kafka for pubsub.
func (p *publisher) publishToController(ctx context.Context, data []byte, caller schema.RefKey) error {
	client := rpc.ClientFromContext[ftlv1pubsubconnect.LegacyPubsubServiceClient](ctx)
	_, err := client.PublishEvent(ctx, connect.NewRequest(&ftlpubsubv1.PublishEventRequest{
		Topic:  &schemapb.Ref{Module: p.module, Name: p.topic.Name},
		Caller: caller.Name,
		Body:   data,
	}))
	if err != nil {
		return fmt.Errorf("failed to publish event to controller: %w", err)
	}
	return nil
}
