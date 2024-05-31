package model

import (
	"errors"
)

type SubscriptionKey = KeyType[SubscriptionPayload, *SubscriptionPayload]

func NewSubscriptionKey(module, name string) SubscriptionKey {
	return newKey[SubscriptionPayload](module, name)
}

func ParseSubscriptionKey(key string) (SubscriptionKey, error) {
	return parseKey[SubscriptionPayload](key)
}

type SubscriptionPayload struct {
	Module string
	Name   string
}

var _ KeyPayload = (*SubscriptionPayload)(nil)

func (s *SubscriptionPayload) Kind() string   { return "sub" }
func (s *SubscriptionPayload) String() string { return s.Module + "-" + s.Name }
func (s *SubscriptionPayload) Parse(parts []string) error {
	if len(parts) != 2 {
		return errors.New("expected <module>-<name> but got empty string")
	}
	s.Module = parts[0]
	s.Name = parts[1]
	return nil
}
func (s *SubscriptionPayload) RandomBytes() int { return 10 }
