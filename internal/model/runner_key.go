package model

import (
	"strconv"
)

type RunnerKey = KeyType[RunnerPayload, *RunnerPayload]

func NewRunnerKey(hostname, port string) RunnerKey {
	return newKey[RunnerPayload](hostname, port)
}

func NewLocalRunnerKey(suffix int) RunnerKey {
	return newKey[RunnerPayload]("", strconv.Itoa(suffix))
}

func ParseRunnerKey(key string) (RunnerKey, error) { return parseKey[RunnerPayload](key) }

type RunnerPayload struct {
	HostPortMixin
}

var _ KeyPayload = (*RunnerPayload)(nil)

func (r *RunnerPayload) Kind() string { return "rnr" }
