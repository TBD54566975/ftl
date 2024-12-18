package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/rpc"
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
	logger := log.FromContext(ctx).Scope("topic:" + p.topic.Name)
	requestKey, err := rpc.RequestKeyFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get request key: %w", err)
	}
	var requestKeyStr optional.Option[string]
	if r, ok := requestKey.Get(); ok {
		requestKeyStr = optional.Some(r.String())
	}

	timelineEvent := timeline.PubSubPublish{
		DeploymentKey: p.deployment,
		RequestKey:    requestKeyStr,
		Time:          time.Now(),
		SourceVerb:    caller,
		Topic:         p.topic.Name,
		Request:       data,
	}

	partition, offset, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic.Runtime.TopicID,
		Value: sarama.ByteEncoder(data),
		Key:   sarama.StringEncoder(key),
	})
	if err != nil {
		timelineEvent.Error = optional.Some(err.Error())
		logger.Errorf(err, "Failed to publish message")
		return fmt.Errorf("failed to publish message: %w", err)
	}
	timelineEvent.Partition = int(partition)
	timelineEvent.Offset = int(offset)
	p.timelineClient.Publish(ctx, timelineEvent)
	logger.Debugf("Published to partition %v with offset %v)", partition, offset)
	return nil
}
