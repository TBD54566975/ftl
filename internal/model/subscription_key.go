package model

import (
	"errors"
	"strings"
)

type SubscriptionKey = KeyType[SubscriptionPayload, *SubscriptionPayload]

func NewSubscriptionKey(module, name string) SubscriptionKey {
	return newKey[SubscriptionPayload](strings.Join([]string{module, name}, "-"))
}

func ParseSubscriptionKey(key string) (SubscriptionKey, error) {
	return parseKey[SubscriptionPayload](key)
}

type SubscriptionPayload struct {
	Ref string
}

var _ KeyPayload = (*SubscriptionPayload)(nil)

func (d *SubscriptionPayload) Kind() string   { return "sub" }
func (d *SubscriptionPayload) String() string { return d.Ref }
func (d *SubscriptionPayload) Parse(parts []string) error {
	if len(parts) == 0 {
		return errors.New("expected <module>-<name> but got empty string")
	}
	d.Ref = strings.Join(parts, "-")
	return nil
}
func (d *SubscriptionPayload) RandomBytes() int { return 10 }
