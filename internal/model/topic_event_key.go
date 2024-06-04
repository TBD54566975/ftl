package model

import (
	"errors"
)

type TopicEventKey = KeyType[TopicEventPayload, *TopicEventPayload]

func NewTopicEventKey(module, topic string) TopicEventKey {
	return newKey[TopicEventPayload](module, topic)
}

func ParseTopicEventKey(key string) (TopicEventKey, error) { return parseKey[TopicEventPayload](key) }

type TopicEventPayload struct {
	Module string
	Topic  string
}

var _ KeyPayload = (*TopicEventPayload)(nil)

func (t *TopicEventPayload) Kind() string   { return "evt" }
func (t *TopicEventPayload) String() string { return t.Module + "-" + t.Topic }
func (t *TopicEventPayload) Parse(parts []string) error {
	if len(parts) != 2 {
		return errors.New("expected <module>-<topic> but got empty string")
	}
	t.Module = parts[0]
	t.Topic = parts[1]
	return nil
}
func (t *TopicEventPayload) RandomBytes() int { return 12 }
