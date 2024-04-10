package model

import (
	"strconv"
)

type ControllerKey = KeyType[ControllerPayload, *ControllerPayload]

func NewControllerKey(hostname, port string) ControllerKey {
	return newKey[ControllerPayload](hostname, port)
}

func NewLocalControllerKey(suffix int) ControllerKey {
	return newKey[ControllerPayload]("", strconv.Itoa(suffix))
}

func ParseControllerKey(key string) (ControllerKey, error) { return parseKey[ControllerPayload](key) }

var _ KeyPayload = (*ControllerPayload)(nil)

type ControllerPayload struct {
	HostPortMixin
}

func (c *ControllerPayload) Kind() string { return "ctr" }
