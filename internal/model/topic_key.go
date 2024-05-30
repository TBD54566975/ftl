package model

import (
	"errors"
	"strings"
)

type TopicKey = KeyType[TopicPayload, *TopicPayload]

func NewTopicKey(module, name string) TopicKey {
	return newKey[TopicPayload](strings.Join([]string{module, name}, "-"))
}

func ParseTopicKey(key string) (TopicKey, error) { return parseKey[TopicPayload](key) }

type TopicPayload struct {
	Ref string
}

var _ KeyPayload = (*TopicPayload)(nil)

func (d *TopicPayload) Kind() string   { return "top" }
func (d *TopicPayload) String() string { return d.Ref }
func (d *TopicPayload) Parse(parts []string) error {
	if len(parts) == 0 {
		return errors.New("expected <module>-<name> but got empty string")
	}
	d.Ref = strings.Join(parts, "-")
	return nil
}
func (d *TopicPayload) RandomBytes() int { return 10 }
