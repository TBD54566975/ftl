package model

import (
	"errors"
)

type SubscriberKey = KeyType[SubscriberPayload, *SubscriberPayload]

func NewSubscriberKey(module, subscription, sink string) SubscriberKey {
	return newKey[SubscriberPayload](module, subscription)
}

func ParseSubscriberKey(key string) (SubscriberKey, error) {
	return parseKey[SubscriberPayload](key)
}

type SubscriberPayload struct {
	Module       string
	Subscription string
}

var _ KeyPayload = (*SubscriberPayload)(nil)

func (s *SubscriberPayload) Kind() string   { return "subr" }
func (s *SubscriberPayload) String() string { return s.Module + "-" + s.Subscription }
func (s *SubscriberPayload) Parse(parts []string) error {
	if len(parts) != 2 {
		return errors.New("expected <module>-<subscription> but got empty string")
	}
	s.Module = parts[0]
	s.Subscription = parts[1]
	return nil
}
func (s *SubscriberPayload) RandomBytes() int { return 10 }
