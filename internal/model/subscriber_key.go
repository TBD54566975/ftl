package model

import (
	"errors"
	"strings"
)

type SubscriberKey = KeyType[SubscriberPayload, *SubscriberPayload]

func NewSubscriberKey(module, subscription, sink string) SubscriberKey {
	return newKey[SubscriberPayload](strings.Join([]string{module, subscription, sink}, "-"))
}

func ParseSubscriberKey(key string) (SubscriberKey, error) {
	return parseKey[SubscriberPayload](key)
}

type SubscriberPayload struct {
	Ref string
}

var _ KeyPayload = (*SubscriberPayload)(nil)

func (d *SubscriberPayload) Kind() string   { return "subr" }
func (d *SubscriberPayload) String() string { return d.Ref }
func (d *SubscriberPayload) Parse(parts []string) error {
	if len(parts) == 0 {
		return errors.New("expected <module>-<subscription>-<sink> but got empty string")
	}
	d.Ref = strings.Join(parts, "-")
	return nil
}
func (d *SubscriberPayload) RandomBytes() int { return 10 }
