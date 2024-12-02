package pubsub

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/TBD54566975/ftl/internal/schema"
)

type publisher struct {
	topic    *schema.Topic
	producer sarama.SyncProducer
}

func newPublisher(t *schema.Topic) (*publisher, error) {
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
		topic:    t,
		producer: producer,
	}, nil
}

func (p *publisher) publish(data []byte, key string) error {
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
