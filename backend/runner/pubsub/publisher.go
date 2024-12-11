package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type publisher struct {
	module     string
	deployment model.DeploymentKey
	topic      *schema.Topic
	producer   sarama.SyncProducer

	timelineClient *timeline.Client
}

func newPublisher(module string, t *schema.Topic, deployment model.DeploymentKey, timelineClient *timeline.Client) (*publisher, error) {
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
		module:     module,
		deployment: deployment,
		topic:      t,
		producer:   producer,

		timelineClient: timelineClient,
	}, nil
}

func (p *publisher) publish(ctx context.Context, data []byte, key string, caller schema.Ref) error {
	// TODO: fill in details
	timelineEvent := timeline.PubSubPublish{
		DeploymentKey: p.deployment,
		// RequestKey    optional.Option[string]
		Time:       time.Now(),
		SourceVerb: caller,
		Topic:      p.topic.Name, // or is this a ref?
		Request:    data,
	}
	// TODO: write offset and partition to timeline event
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic.Runtime.TopicID,
		Value: sarama.ByteEncoder(data),
		Key:   sarama.StringEncoder(key),
	})
	if err != nil {
		timelineEvent.Error = optional.Some(err.Error())
		return fmt.Errorf("failed to publish message: %w", err)
	}
	p.timelineClient.Publish(ctx, timelineEvent)
	return nil
}
