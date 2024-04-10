package model

import (
	"fmt"
	"regexp"
	"strings"
)

// A RequestName represents an inbound request into the cluster.
type RequestName = KeyType[RequestKeyPayload, *RequestKeyPayload]

func NewRequestName(origin Origin, key string) RequestName {
	return newKey[RequestKeyPayload](string(origin), key)
}

func ParseRequestName(name string) (RequestName, error) {
	return parseKey[RequestKeyPayload](name)
}

// Origin of a request.
type Origin string

const (
	OriginIngress Origin = "ingress"
	OriginCron    Origin = "cron"
	OriginPubsub  Origin = "pubsub"
)

func ParseOrigin(origin string) (Origin, error) {
	switch origin {
	case "ingress":
		return OriginIngress, nil
	case "cron":
		return OriginCron, nil
	case "pubsub":
		return OriginPubsub, nil
	default:
		return "", fmt.Errorf("unknown origin %q", origin)
	}
}

var requestNameNormaliserRe = regexp.MustCompile("[^a-zA-Z0-9]+")

type RequestKeyPayload struct {
	Origin Origin
	Key    string
}

func (r *RequestKeyPayload) Kind() string   { return "req" }
func (r *RequestKeyPayload) String() string { return fmt.Sprintf("%s-%s", r.Origin, r.Key) }
func (r *RequestKeyPayload) Parse(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("expected <origin>-<key> but got %q", strings.Join(parts, "-"))
	}
	origin, err := ParseOrigin(parts[0])
	if err != nil {
		return err
	}
	r.Origin = origin
	key := strings.Join(parts[1:], "-")
	r.Key = requestNameNormaliserRe.ReplaceAllString(key, "-")
	return nil
}
func (r *RequestKeyPayload) RandomBytes() int { return 10 }
