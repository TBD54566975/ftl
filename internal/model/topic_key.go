package model

import (
	"errors"
)

type TopicKey = KeyType[TopicPayload, *TopicPayload]

func NewTopicKey(module, name string) TopicKey {
	return newKey[TopicPayload](module, name)
}

func ParseTopicKey(key string) (TopicKey, error) { return parseKey[TopicPayload](key) }

type TopicPayload struct {
	Module string
	Name   string
}

var _ KeyPayload = (*TopicPayload)(nil)

func (t *TopicPayload) Kind() string   { return "top" }
func (t *TopicPayload) String() string { return t.Module + "-" + t.Name }
func (t *TopicPayload) Parse(parts []string) error {
	if len(parts) != 2 {
		return errors.New("expected <module>-<name> but got empty string")
	}
	t.Module = parts[0]
	t.Name = parts[1]
	return nil
}
func (t *TopicPayload) RandomBytes() int { return 10 }
