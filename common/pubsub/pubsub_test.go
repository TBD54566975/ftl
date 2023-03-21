package pubsub

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestPubsub(t *testing.T) {
	pubsub := New[string]()
	ch := make(chan string, 64)
	pubsub.Subscribe(ch)
	pubsub.Publish("hello")
	select {
	case msg := <-ch:
		assert.Equal(t, "hello", msg)

	case <-time.After(time.Millisecond * 100):
		t.Fail()
	}
	_ = pubsub.Close()
	assert.Panics(t, func() { pubsub.Subscribe(ch) })
	assert.Panics(t, func() { pubsub.Unsubscribe(ch) })
	assert.Panics(t, func() { pubsub.Publish("hello") })
}
